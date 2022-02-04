package transformer

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

// ProxyETHNetVersion implements ETHProxy
type ProxyETHNetVersion struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHNetVersion) Method() string {
	return "net_version"
}

func (p *ProxyETHNetVersion) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	return p.request()
}

func (p *ProxyETHNetVersion) request() (*eth.NetVersionResponse, eth.JSONRPCError) {
	networkID, err := getChainId(p.Htmlcoin)
	if err != nil {
		return nil, err
	}
	response := eth.NetVersionResponse(hexutil.EncodeBig(networkID))
	return &response, nil
}
