package transformer

import (
	"context"
	"math"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/denuoweb/janus/pkg/eth"
	"github.com/denuoweb/janus/pkg/htmlcoin"
)

//ProxyETHGetHashrate implements ETHProxy
type ProxyETHHashrate struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHHashrate) Method() string {
	return "eth_hashrate"
}

func (p *ProxyETHHashrate) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	return p.request(c.Request().Context())
}

func (p *ProxyETHHashrate) request(ctx context.Context) (*eth.HashrateResponse, eth.JSONRPCError) {
	htmlcoinresp, err := p.Htmlcoin.GetHashrate(ctx)
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
