package transformer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

// ProxyETHEstimateGas implements ETHProxy
type ProxyETHTxCount struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHTxCount) Method() string {
	return "eth_getTransactionCount"
}

func (p *ProxyETHTxCount) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {

	/* not sure we need this. Need to figure out how to best unmarshal this in the future. For now this will work.
	var req eth.GetTransactionCountRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		return nil, err
	}*/
	htmlcoinresp, err := p.Htmlcoin.GetTransactionCount("", "")
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	// htmlcoin res -> eth res
	return p.response(htmlcoinresp), nil
}

func (p *ProxyETHTxCount) response(htmlcoinresp *big.Int) string {
	return hexutil.EncodeBig(htmlcoinresp)
}
