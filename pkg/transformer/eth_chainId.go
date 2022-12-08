package transformer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

type ProxyETHChainId struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHChainId) Method() string {
	return "eth_chainId"
}

func (p *ProxyETHChainId) Request(req *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	chainId, err := getChainId(p.Htmlcoin)
	if err != nil {
		return nil, err
	}
	return eth.ChainIdResponse(hexutil.EncodeBig(chainId)), nil
}

func getChainId(p *htmlcoin.Htmlcoin) (*big.Int, eth.JSONRPCError) {
	return big.NewInt(int64(p.ChainId())), nil
}
