package transformer

import (
	"context"
	"math/big"

	"github.com/labstack/echo"
	"github.com/denuoweb/janus/pkg/eth"
	"github.com/denuoweb/janus/pkg/htmlcoin"
	"github.com/denuoweb/janus/pkg/utils"
)

// ProxyETHCall implements ETHProxy
type ProxyETHCall struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHCall) Method() string {
	return "eth_call"
}

func (p *ProxyETHCall) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var req eth.CallRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Is this correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	return p.request(c.Request().Context(), &req)
}

func (p *ProxyETHCall) request(ctx context.Context, ethreq *eth.CallRequest) (interface{}, eth.JSONRPCError) {
	// eth req -> htmlcoin req
	htmlcoinreq, jsonErr := p.ToRequest(ethreq)
	if jsonErr != nil {
		return nil, jsonErr
	}
	if htmlcoinreq.GasLimit != nil && htmlcoinreq.GasLimit.Cmp(big.NewInt(40000000)) > 0 {
		htmlcoinresp := eth.CallResponse("0x")
		p.Htmlcoin.GetLogger().Log("msg", "Caller gas above allowance, capping", "requested", htmlcoinreq.GasLimit.Int64(), "cap", "40,000,000")
		return &htmlcoinresp, nil
	}

	htmlcoinresp, err := p.CallContract(ctx, htmlcoinreq)
	if err != nil {
		if err == htmlcoin.ErrInvalidAddress {
			htmlcoinresp := eth.CallResponse("0x")
			return &htmlcoinresp, nil
		}

		return nil, eth.NewCallbackError(err.Error())
	}

	// htmlcoin res -> eth res
	return p.ToResponse(htmlcoinresp), nil
}

func (p *ProxyETHCall) ToRequest(ethreq *eth.CallRequest) (*htmlcoin.CallContractRequest, eth.JSONRPCError) {
	from := ethreq.From
	var err error
	if utils.IsEthHexAddress(from) {
		from, err = p.FromHexAddress(from)
		if err != nil {
			return nil, eth.NewCallbackError(err.Error())
		}
	}

	var gasLimit *big.Int
	if ethreq.Gas != nil {
		gasLimit = ethreq.Gas.Int
	}

	if gasLimit != nil && gasLimit.Int64() < MinimumGasLimit {
		p.GetLogger().Log("msg", "Gas limit is too low", "gasLimit", gasLimit.String())
	}

	return &htmlcoin.CallContractRequest{
		To:       ethreq.To,
		From:     from,
		Data:     ethreq.Data,
		GasLimit: gasLimit,
	}, nil
}

func (p *ProxyETHCall) ToResponse(qresp *htmlcoin.CallContractResponse) interface{} {
	if qresp.ExecutionResult.Output == "" {
		return eth.NewJSONRPCError(
			-32000,
			"Revert: executionResult output is empty",
			nil,
		)
	}

	data := utils.AddHexPrefix(qresp.ExecutionResult.Output)
	htmlcoinresp := eth.CallResponse(data)
	return &htmlcoinresp

}
