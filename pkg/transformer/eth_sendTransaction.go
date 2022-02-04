package transformer

import (
	"github.com/labstack/echo"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
	"github.com/htmlcoin/janus/pkg/utils"
	"github.com/shopspring/decimal"
)

var MinimumGasLimit = int64(22000)

// ProxyETHSendTransaction implements ETHProxy
type ProxyETHSendTransaction struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHSendTransaction) Method() string {
	return "eth_sendTransaction"
}

func (p *ProxyETHSendTransaction) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var req eth.SendTransactionRequest
	err := unmarshalRequest(rawreq.Params, &req)
	if err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	if req.Gas != nil && req.Gas.Int64() < MinimumGasLimit {
		p.GetLogger().Log("msg", "Gas limit is too low", "gasLimit", req.Gas.String())
	}

	var result interface{}
	var jsonErr eth.JSONRPCError

	if req.IsCreateContract() {
		result, jsonErr = p.requestCreateContract(&req)
	} else if req.IsSendEther() {
		result, jsonErr = p.requestSendToAddress(&req)
	} else if req.IsCallContract() {
		result, jsonErr = p.requestSendToContract(&req)
	} else {
		return nil, eth.NewInvalidParamsError("Unknown operation")
	}

	if err == nil {
		p.GenerateIfPossible()
	}

	return result, jsonErr
}

func (p *ProxyETHSendTransaction) requestSendToContract(ethtx *eth.SendTransactionRequest) (*eth.SendTransactionResponse, eth.JSONRPCError) {
	gasLimit, gasPrice, err := EthGasToHtmlcoin(ethtx)
	if err != nil {
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	amount := decimal.NewFromFloat(0.0)
	if ethtx.Value != "" {
		var err error
		amount, err = EthValueToHtmlcoinAmount(ethtx.Value, ZeroSatoshi)
		if err != nil {
			return nil, eth.NewInvalidParamsError(err.Error())
		}
	}

	htmlcoinreq := htmlcoin.SendToContractRequest{
		ContractAddress: utils.RemoveHexPrefix(ethtx.To),
		Datahex:         utils.RemoveHexPrefix(ethtx.Data),
		Amount:          amount,
		GasLimit:        gasLimit,
		GasPrice:        gasPrice,
	}

	if from := ethtx.From; from != "" && utils.IsEthHexAddress(from) {
		from, err = p.FromHexAddress(from)
		if err != nil {
			return nil, eth.NewCallbackError(err.Error())
		}
		htmlcoinreq.SenderAddress = from
	}

	var resp *htmlcoin.SendToContractResponse
	if err := p.Htmlcoin.Request(htmlcoin.MethodSendToContract, &htmlcoinreq, &resp); err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	ethresp := eth.SendTransactionResponse(utils.AddHexPrefix(resp.Txid))
	return &ethresp, nil
}

func (p *ProxyETHSendTransaction) requestSendToAddress(req *eth.SendTransactionRequest) (*eth.SendTransactionResponse, eth.JSONRPCError) {
	getHtmlcoinWalletAddress := func(addr string) (string, error) {
		if utils.IsEthHexAddress(addr) {
			return p.FromHexAddress(utils.RemoveHexPrefix(addr))
		}
		return addr, nil
	}

	from, err := getHtmlcoinWalletAddress(req.From)
	if err != nil {
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	to, err := getHtmlcoinWalletAddress(req.To)
	if err != nil {
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	amount, err := EthValueToHtmlcoinAmount(req.Value, ZeroSatoshi)
	if err != nil {
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	p.GetDebugLogger().Log("msg", "successfully converted from wei to HTML", "wei", req.Value, "htmlcoin", amount)

	htmlcoinreq := htmlcoin.SendToAddressRequest{
		Address:       to,
		Amount:        amount,
		SenderAddress: from,
	}

	var htmlcoinresp htmlcoin.SendToAddressResponse
	if err := p.Htmlcoin.Request(htmlcoin.MethodSendToAddress, &htmlcoinreq, &htmlcoinresp); err != nil {
		// this can fail with:
		// "error": {
		//   "code": -3,
		//   "message": "Sender address does not have any unspent outputs"
		// }
		// this can happen if there are enough coins but some required are untrusted
		// you can get the trusted coin balance via getbalances rpc call
		return nil, eth.NewCallbackError(err.Error())
	}

	ethresp := eth.SendTransactionResponse(utils.AddHexPrefix(string(htmlcoinresp)))

	return &ethresp, nil
}

func (p *ProxyETHSendTransaction) requestCreateContract(req *eth.SendTransactionRequest) (*eth.SendTransactionResponse, eth.JSONRPCError) {
	gasLimit, gasPrice, err := EthGasToHtmlcoin(req)
	if err != nil {
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	htmlcoinreq := &htmlcoin.CreateContractRequest{
		ByteCode: utils.RemoveHexPrefix(req.Data),
		GasLimit: gasLimit,
		GasPrice: gasPrice,
	}

	if req.From != "" {
		from := req.From
		if utils.IsEthHexAddress(from) {
			from, err = p.FromHexAddress(from)
			if err != nil {
				return nil, eth.NewCallbackError(err.Error())
			}
		}

		htmlcoinreq.SenderAddress = from
	}

	var resp *htmlcoin.CreateContractResponse
	if err := p.Htmlcoin.Request(htmlcoin.MethodCreateContract, htmlcoinreq, &resp); err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	ethresp := eth.SendTransactionResponse(utils.AddHexPrefix(string(resp.Txid)))

	return &ethresp, nil
}
