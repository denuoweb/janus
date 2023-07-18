package transformer

import (
	"context"

	"github.com/labstack/echo"
	"github.com/denuoweb/janus/pkg/eth"
	"github.com/denuoweb/janus/pkg/htmlcoin"
	"github.com/denuoweb/janus/pkg/utils"
)

// ProxyETHSendRawTransaction implements ETHProxy
type ProxyETHSendRawTransaction struct {
	*htmlcoin.Htmlcoin
}

var _ ETHProxy = (*ProxyETHSendRawTransaction)(nil)

func (p *ProxyETHSendRawTransaction) Method() string {
	return "eth_sendRawTransaction"
}

func (p *ProxyETHSendRawTransaction) Request(req *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var params eth.SendRawTransactionRequest
	if err := unmarshalRequest(req.Params, &params); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}
	if params[0] == "" {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError("invalid parameter: raw transaction hexed string is empty")
	}

	return p.request(c.Request().Context(), params)
}

func (p *ProxyETHSendRawTransaction) request(ctx context.Context, params eth.SendRawTransactionRequest) (eth.SendRawTransactionResponse, eth.JSONRPCError) {
	var (
		htmlcoinHexedRawTx = utils.RemoveHexPrefix(params[0])
		req            = htmlcoin.SendRawTransactionRequest([1]string{htmlcoinHexedRawTx})
	)

	htmlcoinresp, err := p.Htmlcoin.SendRawTransaction(ctx, &req)
	if err != nil {
		if err == htmlcoin.ErrVerifyAlreadyInChain {
			// already committed
			// we need to send back the tx hash
			rawTx, err := p.Htmlcoin.DecodeRawTransaction(ctx, htmlcoinHexedRawTx)
			if err != nil {
				p.GetErrorLogger().Log("msg", "Error decoding raw transaction for duplicate raw transaction", "err", err)
				return eth.SendRawTransactionResponse(""), eth.NewCallbackError(err.Error())
			}
			htmlcoinresp = &htmlcoin.SendRawTransactionResponse{Result: rawTx.Hash}
		} else {
			return eth.SendRawTransactionResponse(""), eth.NewCallbackError(err.Error())
		}
	} else {
		p.GenerateIfPossible()
	}

	resp := *htmlcoinresp
	ethHexedTxHash := utils.AddHexPrefix(resp.Result)
	return eth.SendRawTransactionResponse(ethHexedTxHash), nil
}
