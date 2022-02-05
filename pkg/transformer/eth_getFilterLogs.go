package transformer

import (
	"math/big"

	"github.com/labstack/echo"
	"github.com/htmlcoin/janus/pkg/eth"
)

// ProxyETHGetFilterLogs implements ETHProxy
type ProxyETHGetFilterLogs struct {
	*ProxyETHGetFilterChanges
}

func (p *ProxyETHGetFilterLogs) Method() string {
	return "eth_getFilterLogs"
}

func (p *ProxyETHGetFilterLogs) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {

	filter, err := processFilter(p.ProxyETHGetFilterChanges, rawreq)
	if err != nil {
		return nil, err
	}

	switch filter.Type {
	case eth.NewFilterTy:
		return p.request(filter)
	default:
		return nil, eth.NewInvalidParamsError("filter not found")
	}
}

func (p *ProxyETHGetFilterLogs) request(filter *eth.Filter) (htmlcoinresp eth.GetFilterChangesResponse, err eth.JSONRPCError) {
	htmlcoinresp = make(eth.GetFilterChangesResponse, 0)

	_lastBlockNumber, ok := filter.Data.Load("lastBlockNumber")
	if !ok {
		return htmlcoinresp, eth.NewCallbackError("Could not get lastBlockNumber")
	}
	lastBlockNumber := _lastBlockNumber.(uint64)

	_toBlock, ok := filter.Data.Load("toBlock")
	if !ok {
		return htmlcoinresp, eth.NewCallbackError("Could not get toBlock")
	}
	toBlock := _toBlock.(uint64)

	searchLogsReq, err := p.ProxyETHGetFilterChanges.toSearchLogsReq(filter, big.NewInt(int64(lastBlockNumber)), big.NewInt(int64(toBlock)))
	if err != nil {
		return nil, err
	}

	return p.ProxyETHGetFilterChanges.doSearchLogs(searchLogsReq)

}
