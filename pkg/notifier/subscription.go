package notifier

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"strings"
	"sync"
	"time"

	"github.com/htmlcoin/janus/pkg/conversion"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

type subscriptionInformation struct {
	*Subscription
	params     *eth.EthSubscriptionRequest
	mutex      sync.RWMutex
	ctx        context.Context
	cancelFunc context.CancelFunc
	running    bool
	htmlcoin       *htmlcoin.Htmlcoin
}

func (s *subscriptionInformation) run() {
	if s.params == nil {
		return
	}

	if strings.ToLower(s.params.Method) != "logs" {
		return
	}

	s.mutex.Lock()
	if s.running {
		s.mutex.Unlock()
		return
	}
	s.running = true
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		defer s.mutex.Unlock()

		s.running = false
	}()

	var nextBlock interface{}
	nextBlock = nil
	translatedTopics, err := eth.TranslateTopics(s.params.Params.Topics)
	if err != nil {
		s.htmlcoin.GetDebugLogger().Log("msg", "Error translating logs topics", "error", err)
		return
	}
	ethAddresses, err := s.params.Params.GetAddresses()
	if err != nil {
		s.htmlcoin.GetDebugLogger().Log("msg", "Error translating logs addresses", "error", err)
		return
	}
	stringAddresses := make([]string, len(ethAddresses))
	for i, ethAddress := range ethAddresses {
		if strings.HasPrefix(ethAddress.String(), "0x") {
			stringAddresses[i] = strings.TrimPrefix(ethAddress.String(), "0x")
		} else {
			stringAddresses[i] = ethAddress.String()
		}
	}

	htmlcoinTopics := htmlcoin.NewSearchLogsTopics(translatedTopics)
	req := &htmlcoin.WaitForLogsRequest{
		FromBlock: nextBlock,
		ToBlock:   nil,
		Filter: htmlcoin.WaitForLogsFilter{
			Addresses: &stringAddresses,
			Topics:    &htmlcoinTopics,
		},
	}

	if s.htmlcoin.Chain() == htmlcoin.ChainRegTest || s.htmlcoin.Chain() == htmlcoin.ChainTest {
		req.MinimumConfirmations = 0
	}

	// this throttles HTMLCOIN api calls if waitforlogs is returning very quickly a lot
	limitToXApiCalls := 5
	inYSeconds := 10 * time.Second
	// if a HTMLCOIN API call returns quicker than this, we will wait until this time is reached
	// this prevents spamming the HTMLCOIN node too much
	minimumTimeBetweenCalls := 100 * time.Millisecond

	rolling := newRollingLimit(limitToXApiCalls)

	// duplicate logs are only to be sent on a reorg
	// the previous log that was sent on the old chain is sent with a `removed: true`
	// then the new log is sent
	// that functionality isn't supported in this implementation yet
	// however, in order to not send duplicate logs we can do that with a simple hash map
	// each hash is a 128bit MD5 hash, the hashing algorithim doesn't really matter here
	// as this is only for preventing duplicate logs being sent over a websocket
	// there are 8000 bits in a kilobyte, thats enough for 62.5 hashes
	// 1MB = 1024KB = 1024KB/1KB = 1024 * 62.5 = 64,000 hashes
	// when proving this as a service, that can add up if tens of thousands are using the service
	// we want to put an upper limit on ram usage for an ip/connection
	// we could also put an absolute upper limit on ram usage for this feature
	// TODO: Deal with RAM usage here when Janus gets large enough
	// some kind of FIFO hashmap?
	sentHashes := make(map[string]bool)

	failures := 0
	for {
		req.FromBlock = nextBlock
		timeBeforeCall := time.Now()
		rolling.Push(&timeBeforeCall)
		resp, err := s.htmlcoin.WaitForLogs(s.ctx, req)
		timeAfterCall := time.Now()
		if err == nil {
			nextBlock = int(resp.NextBlock)
			reqSearchLogs := htmlcoin.SearchLogsRequest{
				FromBlock: big.NewInt(int64(resp.NextBlock - 1)),
				ToBlock:   big.NewInt(int64(resp.NextBlock - 1)),
				Addresses: *req.Filter.Addresses,
				Topics:    *req.Filter.Topics,
			}
			receiptsSearchLogs, err := s.htmlcoin.SearchLogs(s.ctx, &reqSearchLogs)
			if err != nil {
				s.htmlcoin.GetErrorLogger().Log("msg", "Error calling searchLogs", "subscriptionId", s.id, "error", err)
				return
			}
			for _, htmlcoinLog := range receiptsSearchLogs {
				htmlcoinLogs := htmlcoinLog.Log
				logs := conversion.FilterHtmlcoinLogs(stringAddresses, htmlcoinTopics, htmlcoinLogs)
				ethLogs := conversion.ExtractETHLogsFromTransactionReceipt(htmlcoinLog, logs)
				for _, ethLog := range ethLogs {
					subscription := &eth.EthSubscription{
						SubscriptionID: s.Subscription.id,
						Result:         ethLog,
					}
					hash := computeHash(subscription)
					if _, ok := sentHashes[hash]; !ok {
						sentHashes[hash] = true
						s.htmlcoin.GetDebugLogger().Log("subscriptionId", s.id, "msg", "notifying of logs")
						jsonRpcNotification, err := eth.NewJSONRPCNotification("eth_subscription", subscription)
						if err != nil {
							s.htmlcoin.GetErrorLogger().Log("subscriptionId", s.id, "err", err)
							return
						}
						s.Send(jsonRpcNotification)
					}
				}
			}
			oldest := rolling.Oldest()
			a := time.Now()
			if oldest != nil && a.Sub(*oldest.(*time.Time)) < inYSeconds {
				// too many request returning successfully too quickly, slow them down
				failures = failures + 1
			} else {
				failures = 0
			}
		} else {
			// error occurred
			s.htmlcoin.GetDebugLogger().Log("subscriptionId", s.id, "err", err)
			failures = failures + 1
		}

		done := s.ctx.Done()

		select {
		case <-done:
			// err is wrapped so we can't detect (err == context.Cancelled)
			s.htmlcoin.GetDebugLogger().Log("subscriptionId", s.id, "msg", "context closed, dropping subscription")
			return
		default:
		}

		backoffTime := getBackoff(failures, 0, 15*time.Second)

		timeCallTook := timeAfterCall.Sub(timeAfterCall)
		if timeCallTook < minimumTimeBetweenCalls {
			timeLeftUntilMinimumTimeBetweenCallsReached := minimumTimeBetweenCalls - timeCallTook
			backoffTime = time.Duration(math.Max(float64(backoffTime), float64(timeLeftUntilMinimumTimeBetweenCallsReached)))
		}

		if backoffTime > 0 {
			s.htmlcoin.GetDebugLogger().Log("subscriptionId", s.id, "msg", fmt.Sprintf("backing off for %d miliseconds", backoffTime/time.Millisecond))
		}

		select {
		case <-done:
			return
		case <-time.After(backoffTime):
			// ok, try again
		}
	}
}

// Compute hash for the json serialization of the passed in argument
func computeHash(value interface{}) string {
	b, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	md5Hash := md5.Sum(b)
	return string(md5Hash[:])
}

func getBackoff(count int, min time.Duration, max time.Duration) time.Duration {
	maxFailures := 10
	if count == 0 {
		return min
	}

	if count > maxFailures {
		return max
	}

	return ((max - min) / time.Duration(maxFailures)) * time.Duration(count)
}

// implementes an array with a rolling index that returns the oldest inserted element
type rollingLimit struct {
	index int
	limit int
	times []interface{}
}

func newRollingLimit(limit int) *rollingLimit {
	roll := &rollingLimit{
		index: 0,
		limit: limit,
		times: []interface{}{},
	}

	for i := 0; i < limit; i++ {
		roll.times = append(roll.times, nil)
	}

	return roll
}

func (r *rollingLimit) oldest() int {
	return (r.index + 1) % r.limit
}

func (r *rollingLimit) newest() int {
	return r.index
}

func (r *rollingLimit) bump() int {
	r.index = r.oldest()
	return r.index
}

func (r *rollingLimit) Oldest() interface{} {
	return r.times[r.oldest()]
}

func (r *rollingLimit) Push(t interface{}) {
	r.times[r.bump()] = t
}
