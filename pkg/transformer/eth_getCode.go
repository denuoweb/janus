package transformer

import (
	"context"

	"github.com/labstack/echo"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
	"github.com/htmlcoin/janus/pkg/utils"
)

// ProxyETHGetCode implements ETHProxy
type ProxyETHGetCode struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHGetCode) Method() string {
	return "eth_getCode"
}

func (p *ProxyETHGetCode) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var req eth.GetCodeRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	return p.request(c.Request().Context(), &req)
}

func (p *ProxyETHGetCode) request(ctx context.Context, ethreq *eth.GetCodeRequest) (eth.GetCodeResponse, eth.JSONRPCError) {
	htmlcoinreq := htmlcoin.GetAccountInfoRequest(utils.RemoveHexPrefix(ethreq.Address))

	htmlcoinresp, err := p.GetAccountInfo(ctx, &htmlcoinreq)
	if err != nil {
		if err == htmlcoin.ErrInvalidAddress {
			/**
			// correct response for an invalid address
			{
				"jsonrpc": "2.0",
				"id": 123,
				"result": "0x"
			}
			**/
			return "0x", nil
		} else {
			return "", eth.NewCallbackError(err.Error())
		}
	}

	// htmlcoin res -> eth res
	return eth.GetCodeResponse(utils.AddHexPrefix(htmlcoinresp.Code)), nil
}
