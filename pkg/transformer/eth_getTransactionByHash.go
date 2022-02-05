package transformer

import (
	"encoding/json"
	"math/big"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
	"github.com/htmlcoin/janus/pkg/utils"
)

// ProxyETHGetTransactionByHash implements ETHProxy
type ProxyETHGetTransactionByHash struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHGetTransactionByHash) Method() string {
	return "eth_getTransactionByHash"
}

func (p *ProxyETHGetTransactionByHash) Request(req *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var txHash eth.GetTransactionByHashRequest
	if err := json.Unmarshal(req.Params, &txHash); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError("couldn't unmarshal request")
	}
	if txHash == "" {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError("transaction hash is empty")
	}

	htmlcoinReq := &htmlcoin.GetTransactionRequest{
		TxID: utils.RemoveHexPrefix(string(txHash)),
	}
	return p.request(htmlcoinReq)
}

func (p *ProxyETHGetTransactionByHash) request(req *htmlcoin.GetTransactionRequest) (*eth.GetTransactionByHashResponse, eth.JSONRPCError) {
	ethTx, err := getTransactionByHash(p.Htmlcoin, req.TxID)
	if err != nil {
		return nil, err
	}
	return ethTx, nil
}

// TODO: think of returning flag if it's a reward transaction for miner
func getTransactionByHash(p *htmlcoin.Htmlcoin, hash string) (*eth.GetTransactionByHashResponse, eth.JSONRPCError) {
	htmlcoinTx, err := p.GetTransaction(hash)
	var ethTx *eth.GetTransactionByHashResponse
	if err != nil {
		if errors.Cause(err) != htmlcoin.ErrInvalidAddress {
			return nil, eth.NewCallbackError(err.Error())
		}
		var rawHtmlcoinTx *htmlcoin.GetRawTransactionResponse
		ethTx, rawHtmlcoinTx, err = getRewardTransactionByHash(p, hash)
		if err != nil {
			if errors.Cause(err) == htmlcoin.ErrInvalidAddress {
				return nil, nil
			}
			rawTx, err := p.GetRawTransaction(hash, false)
			if err != nil {
				if errors.Cause(err) == htmlcoin.ErrInvalidAddress {
					return nil, nil
				}
				return nil, eth.NewCallbackError(err.Error())
			} else {
				p.GetDebugLogger().Log("msg", "Got raw transaction by hash")
				htmlcoinTx = &htmlcoin.GetTransactionResponse{
					BlockHash:  rawTx.BlockHash,
					BlockIndex: 1, // TODO: Possible to get this somewhere?
					Hex:        rawTx.Hex,
				}
			}
		} else {
			p.GetDebugLogger().Log("msg", "Got reward transaction by hash")
			htmlcoinTx = &htmlcoin.GetTransactionResponse{
				Hex:       rawHtmlcoinTx.Hex,
				BlockHash: rawHtmlcoinTx.BlockHash,
			}
		}

		// return ethTx, nil
	}
	htmlcoinDecodedRawTx, err := p.DecodeRawTransaction(htmlcoinTx.Hex)
	if err != nil {
		return nil, eth.NewCallbackError("couldn't get raw transaction")
	}

	if ethTx == nil {
		ethTx = &eth.GetTransactionByHashResponse{
			Hash:  utils.AddHexPrefix(htmlcoinDecodedRawTx.ID),
			Nonce: "0x0",

			// TODO: researching
			// ? Do we need those values
			//! Added for go-ethereum client support
			V: "0x0",
			R: "0x0",
			S: "0x0",

			Gas:      "0x0",
			GasPrice: "0x0",
		}
	}

	if !htmlcoinTx.IsPending() { // otherwise, the following values must be nulls
		blockNumber, err := getBlockNumberByHash(p, htmlcoinTx.BlockHash)
		if err != nil {
			return nil, eth.NewCallbackError("couldn't get block number by hash")
		}
		ethTx.BlockNumber = hexutil.EncodeUint64(blockNumber)
		ethTx.BlockHash = utils.AddHexPrefix(htmlcoinTx.BlockHash)
		if ethTx.TransactionIndex == "" {
			ethTx.TransactionIndex = hexutil.EncodeUint64(uint64(htmlcoinTx.BlockIndex))
		} else {
			// Already set in getRewardTransactionByHash
		}
	}

	if ethTx.Value == "" {
		// TODO: This CalcAmount() func needs improvement
		ethAmount, err := formatHtmlcoinAmount(htmlcoinDecodedRawTx.CalcAmount())
		if err != nil {
			// TODO: Correct error code?
			return nil, eth.NewInvalidParamsError("couldn't format amount")
		}
		ethTx.Value = ethAmount
	}

	htmlcoinTxContractInfo, isContractTx, err := htmlcoinDecodedRawTx.ExtractContractInfo()
	if err != nil {
		return nil, eth.NewCallbackError(htmlcoinTx.Hex /*"couldn't extract contract info"*/)
	}
	if isContractTx {
		// TODO: research is this allowed? ethTx.Input = utils.AddHexPrefix(htmlcoinTxContractInfo.UserInput)
		if htmlcoinTxContractInfo.UserInput == "" {
			ethTx.Input = "0x0"
		} else {
			ethTx.Input = utils.AddHexPrefix(htmlcoinTxContractInfo.UserInput)
		}
		if htmlcoinTxContractInfo.From != "" {
			ethTx.From = utils.AddHexPrefix(htmlcoinTxContractInfo.From)
		}
		//TODO: research if 'To' adress could be other than zero address when 'isContractTx == TRUE'
		if len(htmlcoinTxContractInfo.To) == 0 {
			ethTx.To = utils.AddHexPrefix(htmlcoin.ZeroAddress)
		} else {
			ethTx.To = utils.AddHexPrefix(htmlcoinTxContractInfo.To)
		}
		ethTx.Gas = hexutil.Encode([]byte(htmlcoinTxContractInfo.GasLimit))
		ethTx.GasPrice = hexutil.Encode([]byte(htmlcoinTxContractInfo.GasPrice))

		return ethTx, nil
	}

	if htmlcoinTx.Generated {
		ethTx.From = utils.AddHexPrefix(htmlcoin.ZeroAddress)
	} else {
		// TODO: Figure out proper way to do this
		// There is a problem with this function, sometimes it returns errors on regtest, empty block?
		// commenting it out as its being overwritten below anyway
		/*
			ethTx.From, err = getNonContractTxSenderAddress(p, htmlcoinDecodedRawTx.Vins)
			if err != nil {
				return nil, eth.NewCallbackError("couldn't get non contract transaction sender address")
			}
		*/
		// TODO: discuss
		// ? Does func above return incorrect address for graph-node (len is < 40)
		// ! Temporary solution
		if ethTx.From == "" {
			ethTx.From = utils.AddHexPrefix(htmlcoin.ZeroAddress)
		}
	}
	if ethTx.To == "" {
		ethTx.To, err = findNonContractTxReceiverAddress(htmlcoinDecodedRawTx.Vouts)
		if err != nil {
			// TODO: discuss, research
			// ? Some vouts doesn't have `receive` category at all
			ethTx.To = utils.AddHexPrefix(htmlcoin.ZeroAddress)

			// TODO: uncomment, after todo above will be resolved
			// return nil, errors.WithMessage(err, "couldn't get non contract transaction receiver address")
		}
	}
	// TODO: discuss
	// ? Does func above return incorrect address for graph-node (len is < 40)
	// ! Temporary solution
	if ethTx.To == "" {
		ethTx.To = utils.AddHexPrefix(htmlcoin.ZeroAddress)
	}

	// TODO: researching
	// ! Temporary solution
	//	if len(htmlcoinTx.Hex) == 0 {
	//		ethTx.Input = "0x0"
	//	} else {
	//		ethTx.Input = utils.AddHexPrefix(htmlcoinTx.Hex)
	//	}
	ethTx.Input = utils.AddHexPrefix(htmlcoinTx.Hex)

	return ethTx, nil
}

// TODO: Does this need to return eth.JSONRPCError
// TODO: discuss
// ? There are `witness` transactions, that is not acquireable nither via `gettransaction`, nor `getrawtransaction`
func getRewardTransactionByHash(p *htmlcoin.Htmlcoin, hash string) (*eth.GetTransactionByHashResponse, *htmlcoin.GetRawTransactionResponse, error) {
	rawHtmlcoinTx, err := p.GetRawTransaction(hash, false)
	if err != nil {
		return nil, nil, errors.WithMessage(err, "couldn't get raw reward transaction")
	}

	ethTx := &eth.GetTransactionByHashResponse{
		Hash:  utils.AddHexPrefix(hash),
		Nonce: "0x0",

		// TODO: discuss
		// ? Expect this value to be always zero
		// Geth returns 0x if there is no input data for a transaction
		Input: "0x",

		// TODO: discuss
		// ? Are zero values applicable
		Gas:      "0x0",
		GasPrice: "0x0",

		// TODO: researching
		// ? Do we need those values
		//! Added for go-ethereum client support
		V: "0x0",
		R: "0x0",
		S: "0x0",
	}

	if !rawHtmlcoinTx.IsPending() {
		blockIndex, err := getTransactionIndexInBlock(p, hash, rawHtmlcoinTx.BlockHash)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "couldn't get transaction index in block")
		}
		ethTx.TransactionIndex = hexutil.EncodeUint64(uint64(blockIndex))

		blockNumber, err := getBlockNumberByHash(p, rawHtmlcoinTx.BlockHash)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "couldn't get block number by hash")
		}
		ethTx.BlockNumber = hexutil.EncodeUint64(blockNumber)

		ethTx.BlockHash = utils.AddHexPrefix(rawHtmlcoinTx.BlockHash)
	}

	for i := range rawHtmlcoinTx.Vouts {
		// TODO: discuss
		// ! The response may be null, even if txout is presented
		_, err := p.GetTransactionOut(hash, i, rawHtmlcoinTx.IsPending())
		if err != nil {
			return nil, nil, errors.WithMessage(err, "couldn't get transaction out")
		}
		// TODO: discuss, researching
		// ? Where is a reward amount
		ethTx.Value = "0x0"
	}

	// TODO: discuss
	// ? Do we have to set `from` == `0x00..00`
	ethTx.From = utils.AddHexPrefix(htmlcoin.ZeroAddress)

	// I used Base58AddressToHex at the moment
	// because convertHtmlcoinAddress functions causes error for
	// P2Sh address(such as MUrenj2sPqEVTiNbHQ2RARiZYyTAAeKiDX) and BECH32 address (such as qc1qkt33x6hkrrlwlr6v59wptwy6zskyrjfe40y0lx)
	if rawHtmlcoinTx.OP_SENDER != "" {
		// addr, err := convertHtmlcoinAddress(rawHtmlcoinTx.OP_SENDER)
		addr, err := p.Base58AddressToHex(rawHtmlcoinTx.OP_SENDER)
		if err == nil {
			ethTx.From = utils.AddHexPrefix(addr)
		}
	} else if len(rawHtmlcoinTx.Vins) > 0 && rawHtmlcoinTx.Vins[0].Address != "" {
		// addr, err := convertHtmlcoinAddress(rawHtmlcoinTx.Vins[0].Address)
		addr, err := p.Base58AddressToHex(rawHtmlcoinTx.Vins[0].Address)
		if err == nil {
			ethTx.From = utils.AddHexPrefix(addr)
		}
	}
	// TODO: discuss
	// ? Where is a `to`
	ethTx.To = utils.AddHexPrefix(htmlcoin.ZeroAddress)

	// when sending HTMLCOIN, the first vout will be the target
	// the second will be change from the vin, it will be returned to the same account
	if len(rawHtmlcoinTx.Vouts) >= 2 {
		from := ""
		if len(rawHtmlcoinTx.Vins) > 0 {
			from = rawHtmlcoinTx.Vins[0].Address
		}

		var valueIn int64
		var valueOut int64
		var refund int64
		var sent int64
		var sentTo int64

		for _, vin := range rawHtmlcoinTx.Vins {
			valueIn += vin.AmountSatoshi
		}

		var to string

		for _, vout := range rawHtmlcoinTx.Vouts {
			valueOut += vout.AmountSatoshi
			addressesCount := len(vout.Details.Addresses)
			if addressesCount > 0 && vout.Details.Addresses[0] == from {
				refund += vout.AmountSatoshi
			} else {
				if addressesCount > 0 && vout.Details.Addresses[0] != "" {
					if to == "" {
						to = vout.Details.Addresses[0]
					}
					if to == vout.Details.Addresses[0] {
						sentTo += vout.AmountSatoshi
					}
				}
				sent += vout.AmountSatoshi
			}
		}
		fee := valueIn - valueOut
		if fee < 0 {
			return nil, nil, errors.New("Detected negative fee - shouldn't happen")
		}

		if refund == 0 && sent == 0 {
			// entire tx was burnt
		} else if refund == 0 {
			// no refund, entire vin was consumed
			// subtract fee from sent coins
			sent -= fee
			sentTo -= fee
		} else {
			// no coins sent to anybody
			// subtract fee from refund
			refund -= fee
		}
		sentToInWei := convertFromSatoshiToWei(big.NewInt(sentTo))
		ethTx.Value = hexutil.EncodeUint64(sentToInWei.Uint64())

		toAddress, err := p.Base58AddressToHex(to)
		if err == nil {
			ethTx.To = utils.AddHexPrefix(toAddress)
		}

		// TODO: compute gasPrice based on fee, guess a gas amount based on vin/vout
	}

	return ethTx, rawHtmlcoinTx, nil
}
