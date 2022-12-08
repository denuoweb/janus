package transformer

import (
	"context"

	"github.com/labstack/echo"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

//ProxyETHGetHashrate implements ETHProxy
type ProxyETHMining struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHMining) Method() string {
	return "eth_mining"
}

func (p *ProxyETHMining) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	return p.request(c.Request().Context())
}

func (p *ProxyETHMining) request(ctx context.Context) (*eth.MiningResponse, eth.JSONRPCError) {
	htmlcoinresp, err := p.Htmlcoin.GetMining(ctx)
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	// htmlcoin res -> eth res
	return p.ToResponse(htmlcoinresp), nil
}

func (p *ProxyETHMining) ToResponse(htmlcoinresp *htmlcoin.GetMiningResponse) *eth.MiningResponse {
	ethresp := eth.MiningResponse(htmlcoinresp.Staking)
	return &ethresp
}
