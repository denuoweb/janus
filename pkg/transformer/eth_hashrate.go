package transformer

import (
	"math"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

//ProxyETHGetHashrate implements ETHProxy
type ProxyETHHashrate struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHHashrate) Method() string {
	return "eth_hashrate"
}

func (p *ProxyETHHashrate) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	return p.request()
}

func (p *ProxyETHHashrate) request() (*eth.HashrateResponse, eth.JSONRPCError) {
	htmlcoinresp, err := p.Htmlcoin.GetHashrate()
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	// htmlcoin res -> eth res
	return p.ToResponse(htmlcoinresp), nil
}

func (p *ProxyETHHashrate) ToResponse(htmlcoinresp *htmlcoin.GetHashrateResponse) *eth.HashrateResponse {
	hexVal := hexutil.EncodeUint64(math.Float64bits(htmlcoinresp.Difficulty))
	ethresp := eth.HashrateResponse(hexVal)
	return &ethresp
}
