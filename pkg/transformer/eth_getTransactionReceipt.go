package transformer

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/htmlcoin/janus/pkg/conversion"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
	"github.com/htmlcoin/janus/pkg/utils"
)

var STATUS_SUCCESS = "0x1"
var STATUS_FAILURE = "0x0"

// ProxyETHGetTransactionReceipt implements ETHProxy
type ProxyETHGetTransactionReceipt struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHGetTransactionReceipt) Method() string {
	return "eth_getTransactionReceipt"
}

func (p *ProxyETHGetTransactionReceipt) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var req eth.GetTransactionReceiptRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}
	if req == "" {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError("empty transaction hash")
	}
	var (
		txHash  = utils.RemoveHexPrefix(string(req))
		htmlcoinReq = htmlcoin.GetTransactionReceiptRequest(txHash)
	)
	return p.request(&htmlcoinReq)
}

func (p *ProxyETHGetTransactionReceipt) request(req *htmlcoin.GetTransactionReceiptRequest) (*eth.GetTransactionReceiptResponse, eth.JSONRPCError) {
	htmlcoinReceipt, err := p.Htmlcoin.GetTransactionReceipt(string(*req))
	if err != nil {
		ethTx, _, getRewardTransactionErr := getRewardTransactionByHash(p.Htmlcoin, string(*req))
		if getRewardTransactionErr != nil {
			errCause := errors.Cause(err)
			if errCause == htmlcoin.EmptyResponseErr {
				return nil, nil
			}
			p.Htmlcoin.GetDebugLogger().Log("msg", "Transaction does not exist", "txid", string(*req))
			return nil, eth.NewCallbackError(err.Error())
		}
		return &eth.GetTransactionReceiptResponse{
			TransactionHash:  ethTx.Hash,
			TransactionIndex: ethTx.TransactionIndex,
			BlockHash:        ethTx.BlockHash,
			BlockNumber:      ethTx.BlockNumber,
			// TODO: This is higher than GasUsed in geth but does it matter?
			CumulativeGasUsed: NonContractVMGasLimit,
			EffectiveGasPrice: "0x0",
			GasUsed:           NonContractVMGasLimit,
			From:              ethTx.From,
			To:                ethTx.To,
			Logs:              []eth.Log{},
			LogsBloom:         eth.EmptyLogsBloom,
			Status:            STATUS_SUCCESS,
		}, nil
	}

	ethReceipt := &eth.GetTransactionReceiptResponse{
		TransactionHash:   utils.AddHexPrefix(htmlcoinReceipt.TransactionHash),
		TransactionIndex:  hexutil.EncodeUint64(htmlcoinReceipt.TransactionIndex),
		BlockHash:         utils.AddHexPrefix(htmlcoinReceipt.BlockHash),
		BlockNumber:       hexutil.EncodeUint64(htmlcoinReceipt.BlockNumber),
		ContractAddress:   utils.AddHexPrefixIfNotEmpty(htmlcoinReceipt.ContractAddress),
		CumulativeGasUsed: hexutil.EncodeUint64(htmlcoinReceipt.CumulativeGasUsed),
		GasUsed:           hexutil.EncodeUint64(htmlcoinReceipt.GasUsed),
		From:              utils.AddHexPrefixIfNotEmpty(htmlcoinReceipt.From),
		To:                utils.AddHexPrefixIfNotEmpty(htmlcoinReceipt.To),

		// TODO: researching
		// ! Temporary accept this value to be always zero, as it is at eth logs
		LogsBloom: eth.EmptyLogsBloom,
	}

	status := STATUS_FAILURE
	if htmlcoinReceipt.Excepted == "None" {
		status = STATUS_SUCCESS
	} else {
		p.Htmlcoin.GetDebugLogger().Log("transaction", ethReceipt.TransactionHash, "msg", "transaction excepted", "message", htmlcoinReceipt.Excepted)
	}
	ethReceipt.Status = status

	r := htmlcoin.TransactionReceipt(*htmlcoinReceipt)
	ethReceipt.Logs = conversion.ExtractETHLogsFromTransactionReceipt(&r, r.Log)

	htmlcoinTx, err := p.Htmlcoin.GetRawTransaction(htmlcoinReceipt.TransactionHash, false)
	if err != nil {
		p.GetDebugLogger().Log("msg", "couldn't get transaction", "err", err)
		return nil, eth.NewCallbackError("couldn't get transaction")
	}
	decodedRawHtmlcoinTx, err := p.Htmlcoin.DecodeRawTransaction(htmlcoinTx.Hex)
	if err != nil {
		p.GetDebugLogger().Log("msg", "couldn't decode raw transaction", "err", err)
		return nil, eth.NewCallbackError("couldn't decode raw transaction")
	}
	if decodedRawHtmlcoinTx.IsContractCreation() {
		ethReceipt.To = ""
	} else {
		ethReceipt.ContractAddress = ""
	}

	// TODO: researching
	// - The following code reason is unknown (see original comment)
	// - Code temporary commented, until an error occures
	// ! Do not remove
	// // contractAddress : DATA, 20 Bytes - The contract address created, if the transaction was a contract creation, otherwise null.
	// if status != "0x1" {
	// 	// if failure, should return null for contractAddress, instead of the zero address.
	// 	ethTxReceipt.ContractAddress = ""
	// }

	return ethReceipt, nil
}
