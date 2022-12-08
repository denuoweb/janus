package transformer

import (
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

// ProxyETHBlockNumber implements ETHProxy
type ProxyETHBlockNumber struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHBlockNumber) Method() string {
	return "eth_blockNumber"
}

func (p *ProxyETHBlockNumber) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	return p.request(c, 5)
}

func (p *ProxyETHBlockNumber) request(c echo.Context, retries int) (*eth.BlockNumberResponse, eth.JSONRPCError) {
	htmlcoinresp, err := p.Htmlcoin.GetBlockCount(c.Request().Context())
	if err != nil {
		if retries > 0 && strings.Contains(err.Error(), htmlcoin.ErrTryAgain.Error()) {
			ctx := c.Request().Context()
			t := time.NewTimer(500 * time.Millisecond)
			select {
			case <-ctx.Done():
				return nil, eth.NewCallbackError(err.Error())
			case <-t.C:
				// fallthrough
			}
			return p.request(c, retries-1)
		}
		return nil, eth.NewCallbackError(err.Error())
	}

	// htmlcoin res -> eth res
	return p.ToResponse(htmlcoinresp), nil
}

func (p *ProxyETHBlockNumber) ToResponse(htmlcoinresp *htmlcoin.GetBlockCountResponse) *eth.BlockNumberResponse {
	hexVal := hexutil.EncodeBig(htmlcoinresp.Int)
	ethresp := eth.BlockNumberResponse(hexVal)
	return &ethresp
}
