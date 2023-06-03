package server

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

var ErrNoHtmlcoinConnections = errors.New("htmlcoind has no connections")
var ErrCannotGetConnectedChain = errors.New("Cannot detect chain htmlcoind is connected to")
var ErrBlockSyncingSeemsStalled = errors.New("Block syncing seems stalled")
var ErrLostLotsOfBlocks = errors.New("Lost a lot of blocks, expected block height to be higher")
var ErrLostFewBlocks = errors.New("Lost a few blocks, expected block height to be higher")

func (s *Server) testConnectionToHtmlcoind() error {
	networkInfo, err := s.htmlcoinRPCClient.GetNetworkInfo(s.htmlcoinRPCClient.GetContext())
	if err == nil {
		// chain can theoretically block forever if htmlcoind isn't up
		// but then GetNetworkInfo would be erroring
		chainChan := make(chan string)
		getChainTimeout := time.NewTimer(10 * time.Second)
		go func(ch chan string) {
			chain := s.htmlcoinRPCClient.Chain()
			chainChan <- chain
		}(chainChan)

		select {
		case chain := <-chainChan:
			if chain == htmlcoin.ChainRegTest {
				// ignore how many connections there are
				return nil
			}
			if networkInfo.Connections == 0 {
				s.logger.Log("liveness", "Htmlcoind has no network connections")
				return ErrNoHtmlcoinConnections
			}
			break
		case <-getChainTimeout.C:
			s.logger.Log("liveness", "Htmlcoind getnetworkinfo request timed out")
			return ErrCannotGetConnectedChain
		}
	} else {
		s.logger.Log("liveness", "Htmlcoind getnetworkinfo errored", "err", err)
	}
	return err
}

func (s *Server) testLogEvents() error {
	_, err := s.htmlcoinRPCClient.GetTransactionReceipt(s.htmlcoinRPCClient.GetContext(), "0000000000000000000000000000000000000000000000000000000000000000")
	if err == htmlcoin.ErrInternalError {
		s.logger.Log("liveness", "-logevents might not be enabled")
		return errors.Wrap(err, "-logevents might not be enabled")
	}
	return nil
}

func (s *Server) testBlocksSyncing() error {
	s.blocksMutex.RLock()
	nextBlockCheck := s.nextBlockCheck
	lastBlockStatus := s.lastBlockStatus
	s.blocksMutex.RUnlock()
	now := time.Now()
	if nextBlockCheck == nil {
		nextBlockCheckTime := time.Now().Add(-30 * time.Minute)
		nextBlockCheck = &nextBlockCheckTime
	}
	if nextBlockCheck.After(now) {
		if lastBlockStatus != nil {
			s.logger.Log("liveness", "blocks syncing", "err", lastBlockStatus)
		}
		return lastBlockStatus
	}
	s.blocksMutex.Lock()
	if s.nextBlockCheck != nil && nextBlockCheck != s.nextBlockCheck {
		// multiple threads were waiting on write lock
		s.blocksMutex.Unlock()
		return s.testBlocksSyncing()
	}
	defer s.blocksMutex.Unlock()

	blockChainInfo, err := s.htmlcoinRPCClient.GetBlockChainInfo(s.htmlcoinRPCClient.GetContext())
	if err != nil {
		s.logger.Log("liveness", "getblockchainfo request failed", "err", err)
		return err
	}

	nextBlockCheckTime := time.Now().Add(5 * time.Minute)
	s.nextBlockCheck = &nextBlockCheckTime

	if blockChainInfo.Blocks == s.lastBlock {
		// stalled
		nextBlockCheckTime = time.Now().Add(15 * time.Second)
		s.nextBlockCheck = &nextBlockCheckTime
		s.lastBlockStatus = ErrBlockSyncingSeemsStalled
	} else if blockChainInfo.Blocks < s.lastBlock {
		// lost some blocks...?
		if s.lastBlock-blockChainInfo.Blocks > 10 {
			// lost a lot of blocks
			// probably a real problem
			s.lastBlock = 0
			nextBlockCheckTime = time.Now().Add(60 * time.Second)
			s.nextBlockCheck = &nextBlockCheckTime
			s.logger.Log("liveness", "Lost lots of blocks")
			s.lastBlockStatus = ErrLostLotsOfBlocks
		} else {
			// lost a few blocks
			// could be htmlcoind nodes out of sync behind a load balancer
			nextBlockCheckTime = time.Now().Add(10 * time.Second)
			s.nextBlockCheck = &nextBlockCheckTime
			s.logger.Log("liveness", "Lost a few blocks")
			s.lastBlockStatus = ErrLostFewBlocks
		}
	} else {
		// got a higher block height than last time
		s.lastBlock = blockChainInfo.Blocks
		nextBlockCheckTime = time.Now().Add(90 * time.Second)
		s.nextBlockCheck = &nextBlockCheckTime
		s.lastBlockStatus = nil
	}

	return s.lastBlockStatus
}

func (s *Server) testHtmlcoindErrorRate() error {
	minimumSuccessRate := float32(*s.healthCheckPercent / 100)
	htmlcoinSuccessRate := s.htmlcoinRequestAnalytics.GetSuccessRate()

	if htmlcoinSuccessRate < minimumSuccessRate {
		s.logger.Log("liveness", "htmlcoind request success rate is low", "rate", htmlcoinSuccessRate)
		return errors.New(fmt.Sprintf("htmlcoind request success rate is %f<%f", htmlcoinSuccessRate, minimumSuccessRate))
	} else {
		return nil
	}
}

func (s *Server) testJanusErrorRate() error {
	minimumSuccessRate := float32(*s.healthCheckPercent / 100)
	ethSuccessRate := s.ethRequestAnalytics.GetSuccessRate()

	if ethSuccessRate < minimumSuccessRate {
		s.logger.Log("liveness", "client eth success rate is low", "rate", ethSuccessRate)
		return errors.New(fmt.Sprintf("client eth request success rate is %f<%f", ethSuccessRate, minimumSuccessRate))
	} else {
		return nil
	}
}
