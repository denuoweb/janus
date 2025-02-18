package transformer

import (
	"context"

	"github.com/dcb9/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/denuoweb/janus/pkg/eth"
	"github.com/denuoweb/janus/pkg/htmlcoin"
)

// ProxyNetPeerCount implements ETHProxy
type ProxyNetPeerCount struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyNetPeerCount) Method() string {
	return "net_peerCount"
}

func (p *ProxyNetPeerCount) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	return p.request(c.Request().Context())
}

func (p *ProxyNetPeerCount) request(ctx context.Context) (*eth.NetPeerCountResponse, eth.JSONRPCError) {
	peerInfos, err := p.GetPeerInfo(ctx)
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	resp := eth.NetPeerCountResponse(hexutil.EncodeUint64(uint64(len(peerInfos))))
	return &resp, nil
}
