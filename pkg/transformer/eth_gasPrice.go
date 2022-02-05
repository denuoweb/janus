package transformer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

// ProxyETHEstimateGas implements ETHProxy
type ProxyETHGasPrice struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHGasPrice) Method() string {
	return "eth_gasPrice"
}

func (p *ProxyETHGasPrice) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	htmlcoinresp, err := p.Htmlcoin.GetGasPrice()
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	// htmlcoin res -> eth res
	return p.response(htmlcoinresp), nil
}

func (p *ProxyETHGasPrice) response(htmlcoinresp *big.Int) string {
	// 34 GWEI is the minimum price that HTMLCOIN will confirm tx with
	return hexutil.EncodeBig(convertFromSatoshiToWei(htmlcoinresp))
}
