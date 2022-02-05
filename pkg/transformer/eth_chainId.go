package transformer

import (
	"math/big"
	"strings"

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
	var htmlcoinresp *htmlcoin.GetBlockChainInfoResponse
	if err := p.Request(htmlcoin.MethodGetBlockChainInfo, nil, &htmlcoinresp); err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	var chainId *big.Int
	switch strings.ToLower(htmlcoinresp.Chain) {
	case "main":
		chainId = big.NewInt(4444)
	case "test":
		chainId = big.NewInt(4445)
	case "regtest":
		chainId = big.NewInt(4446)
	default:
		chainId = big.NewInt(4446)
		p.GetDebugLogger().Log("msg", "Unknown chain "+htmlcoinresp.Chain)
	}

	return chainId, nil
}
