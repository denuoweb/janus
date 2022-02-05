package transformer

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

// 22000
var NonContractVMGasLimit = "0x55f0"
var ErrExecutionReverted = errors.New("execution reverted")

var GAS_BUFFER = 1.10

// ProxyETHEstimateGas implements ETHProxy
type ProxyETHEstimateGas struct {
	*ProxyETHCall
}

func (p *ProxyETHEstimateGas) Method() string {
	return "eth_estimateGas"
}

func (p *ProxyETHEstimateGas) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var ethreq eth.CallRequest
	if jsonErr := unmarshalRequest(rawreq.Params, &ethreq); jsonErr != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(jsonErr.Error())
	}

	if ethreq.Data == "" {
		response := eth.EstimateGasResponse(NonContractVMGasLimit)
		return &response, nil
	}

	// when supplying this parameter to callcontract to estimate gas in the htmlcoin api
	// if there isn't enough gas specified here, the result will be an exception
	// Excepted = "OutOfGasIntrinsic"
	// Gas = "the supplied value"
	// this is different from geth's behavior
	// which will return a used gas value that is higher than the incoming gas parameter
	// so we set this to nil so that callcontract will return the actual gas estimate
	ethreq.Gas = nil

	// eth req -> htmlcoin req
	htmlcoinreq, jsonErr := p.ToRequest(&ethreq)
	if jsonErr != nil {
		return nil, jsonErr
	}

	// htmlcoin [code: -5] Incorrect address occurs here
	htmlcoinresp, err := p.CallContract(htmlcoinreq)
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	return p.toResp(htmlcoinresp)
}

func (p *ProxyETHEstimateGas) toResp(htmlcoinresp *htmlcoin.CallContractResponse) (*eth.EstimateGasResponse, eth.JSONRPCError) {
	if htmlcoinresp.ExecutionResult.Excepted != "None" {
		return nil, eth.NewCallbackError(ErrExecutionReverted.Error())
	}
	gas := eth.EstimateGasResponse(hexutil.EncodeUint64(uint64(float64(htmlcoinresp.ExecutionResult.GasUsed) * GAS_BUFFER)))
	p.GetDebugLogger().Log(p.Method(), gas)
	return &gas, nil
}
