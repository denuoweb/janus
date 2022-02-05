package transformer

import (
	"encoding/json"
	"math/big"

	"github.com/labstack/echo"

	"github.com/htmlcoin/janus/pkg/conversion"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
	"github.com/htmlcoin/janus/pkg/utils"
)

// ProxyETHGetFilterChanges implements ETHProxy
type ProxyETHGetFilterChanges struct {
	*htmlcoin.Htmlcoin
	filter *eth.FilterSimulator
}

func (p *ProxyETHGetFilterChanges) Method() string {
	return "eth_getFilterChanges"
}

func (p *ProxyETHGetFilterChanges) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {

	filter, err := processFilter(p, rawreq)
	if err != nil {
		return nil, err
	}

	switch filter.Type {
	case eth.NewFilterTy:
		return p.requestFilter(filter)
	case eth.NewBlockFilterTy:
		return p.requestBlockFilter(filter)
	case eth.NewPendingTransactionFilterTy:
		fallthrough
	default:
		return nil, eth.NewInvalidParamsError("Unknown filter type")
	}
}

func (p *ProxyETHGetFilterChanges) requestBlockFilter(filter *eth.Filter) (htmlcoinresp eth.GetFilterChangesResponse, err eth.JSONRPCError) {
	htmlcoinresp = make(eth.GetFilterChangesResponse, 0)

	_lastBlockNumber, ok := filter.Data.Load("lastBlockNumber")
	if !ok {
		return htmlcoinresp, eth.NewCallbackError("Could not get lastBlockNumber")
	}
	lastBlockNumber := _lastBlockNumber.(uint64)

	blockCountBigInt, blockErr := p.GetBlockCount()
	if blockErr != nil {
		return htmlcoinresp, eth.NewCallbackError(blockErr.Error())
	}
	blockCount := blockCountBigInt.Uint64()

	differ := blockCount - lastBlockNumber

	hashes := make(eth.GetFilterChangesResponse, differ)
	for i := range hashes {
		blockNumber := new(big.Int).SetUint64(lastBlockNumber + uint64(i) + 1)

		resp, err := p.GetBlockHash(blockNumber)
		if err != nil {
			return htmlcoinresp, eth.NewCallbackError(err.Error())
		}

		hashes[i] = utils.AddHexPrefix(string(resp))
	}

	htmlcoinresp = hashes
	filter.Data.Store("lastBlockNumber", blockCount)
	return
}

func (p *ProxyETHGetFilterChanges) requestFilter(filter *eth.Filter) (htmlcoinresp eth.GetFilterChangesResponse, err eth.JSONRPCError) {
	htmlcoinresp = make(eth.GetFilterChangesResponse, 0)

	_lastBlockNumber, ok := filter.Data.Load("lastBlockNumber")
	if !ok {
		return htmlcoinresp, eth.NewCallbackError("Could not get lastBlockNumber")
	}
	lastBlockNumber := _lastBlockNumber.(uint64)

	blockCountBigInt, blockErr := p.GetBlockCount()
	if blockErr != nil {
		return htmlcoinresp, eth.NewCallbackError(blockErr.Error())
	}
	blockCount := blockCountBigInt.Uint64()

	differ := blockCount - lastBlockNumber

	if differ == 0 {
		return eth.GetFilterChangesResponse{}, nil
	}

	searchLogsReq, err := p.toSearchLogsReq(filter, big.NewInt(int64(lastBlockNumber+1)), big.NewInt(int64(blockCount)))
	if err != nil {
		return nil, err
	}

	return p.doSearchLogs(searchLogsReq)
}

func (p *ProxyETHGetFilterChanges) doSearchLogs(req *htmlcoin.SearchLogsRequest) (eth.GetFilterChangesResponse, eth.JSONRPCError) {
	resp, err := conversion.SearchLogsAndFilterExtraTopics(p.Htmlcoin, req)
	if err != nil {
		return nil, err
	}

	receiptToResult := func(receipt *htmlcoin.TransactionReceipt) []interface{} {
		logs := conversion.ExtractETHLogsFromTransactionReceipt(receipt, receipt.Log)
		res := make([]interface{}, len(logs))
		for i := range res {
			res[i] = logs[i]
		}
		return res
	}
	results := make(eth.GetFilterChangesResponse, 0)
	for _, receipt := range resp {
		r := htmlcoin.TransactionReceipt(receipt)
		results = append(results, receiptToResult(&r)...)
	}

	return results, nil
}

func (p *ProxyETHGetFilterChanges) toSearchLogsReq(filter *eth.Filter, from, to *big.Int) (*htmlcoin.SearchLogsRequest, eth.JSONRPCError) {
	ethreq := filter.Request.(*eth.NewFilterRequest)
	var err error
	var addresses []string
	if ethreq.Address != nil {
		if isBytesOfString(ethreq.Address) {
			var addr string
			if err = json.Unmarshal(ethreq.Address, &addr); err != nil {
				// TODO: Correct error code?
				return nil, eth.NewInvalidParamsError(err.Error())
			}
			addresses = append(addresses, addr)
		} else {
			if err = json.Unmarshal(ethreq.Address, &addresses); err != nil {
				// TODO: Correct error code?
				return nil, eth.NewInvalidParamsError(err.Error())
			}
		}
		for i := range addresses {
			addresses[i] = utils.RemoveHexPrefix(addresses[i])
		}
	}

	htmlcoinreq := &htmlcoin.SearchLogsRequest{
		Addresses: addresses,
		FromBlock: from,
		ToBlock:   to,
	}

	topics, ok := filter.Data.Load("topics")
	if ok {
		htmlcoinreq.Topics = topics.([]htmlcoin.SearchLogsTopic)
	}

	return htmlcoinreq, nil
}
