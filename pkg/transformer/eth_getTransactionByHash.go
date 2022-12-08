package transformer

import (
	"context"
	"encoding/json"
	"math/big"
	"strings"

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
	return p.request(c.Request().Context(), htmlcoinReq)
}

func (p *ProxyETHGetTransactionByHash) request(ctx context.Context, req *htmlcoin.GetTransactionRequest) (*eth.GetTransactionByHashResponse, eth.JSONRPCError) {
	ethTx, err := getTransactionByHash(ctx, p.Htmlcoin, req.TxID)
	if err != nil {
		return nil, err
	}
	return ethTx, nil
}

// TODO: think of returning flag if it's a reward transaction for miner
//
// FUTURE WORK: It might be possible to simplify this (and other?) translation by using a single verbose getblock htmlcoin RPC command,
// since it returns a lot of data including the equivalent of calling GetRawTransaction on every transaction in block.
// The last point is of particular interest because GetRawTransaction doesn't by default work for every transaction.
// This would mean fetching a lot of probably unnecessary data, but in this setup query response delay is reasonably the biggest bottleneck anyway
func getTransactionByHash(ctx context.Context, p *htmlcoin.Htmlcoin, hash string) (*eth.GetTransactionByHashResponse, eth.JSONRPCError) {
	htmlcoinTx, err := p.GetTransaction(ctx, hash)
	var ethTx *eth.GetTransactionByHashResponse
	if err != nil {
		if errors.Cause(err) != htmlcoin.ErrInvalidAddress {
			p.GetDebugLogger().Log("msg", "Failed to GetTransaction", "hash", hash, "err", err)
			return nil, eth.NewCallbackError(err.Error())
		}
		var rawHtmlcoinTx *htmlcoin.GetRawTransactionResponse
		ethTx, rawHtmlcoinTx, err = getRewardTransactionByHash(ctx, p, hash)
		if err != nil {
			if errors.Cause(err) == htmlcoin.ErrInvalidAddress {
				return nil, nil
			}
			rawTx, err := p.GetRawTransaction(ctx, hash, false)
			if err != nil {
				if errors.Cause(err) == htmlcoin.ErrInvalidAddress {
					return nil, nil
				}
				p.GetDebugLogger().Log("msg", "Failed to GetRawTransaction", "hash", hash, "err", err)
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

	}
	htmlcoinDecodedRawTx, err := p.DecodeRawTransaction(ctx, htmlcoinTx.Hex)
	if err != nil {
		p.GetDebugLogger().Log("msg", "Failed to DecodeRawTransaction", "hex", htmlcoinTx.Hex, "err", err)
		return nil, eth.NewCallbackError("couldn't get raw transaction")
	}

	if ethTx == nil {
		ethTx = &eth.GetTransactionByHashResponse{
			Hash:  utils.AddHexPrefix(htmlcoinDecodedRawTx.ID),
			Nonce: "0x0",

			// Added for go-ethereum client and graph-node support
			R: "0xf000000000000000000000000000000000000000000000000000000000000000",
			S: "0xf000000000000000000000000000000000000000000000000000000000000000",
			V: "0x25",

			Gas:      "0x0",
			GasPrice: "0x0",
		}
	}

	if !htmlcoinTx.IsPending() { // otherwise, the following values must be nulls
		blockNumber, err := getBlockNumberByHash(ctx, p, htmlcoinTx.BlockHash)
		if err != nil {
			p.GetDebugLogger().Log("msg", "Failed to get block number by hash", "hash", htmlcoinTx.BlockHash, "err", err)
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
			p.GetDebugLogger().Log("msg", "Couldn't format htmlcoin amount", "htmlcoin", htmlcoinDecodedRawTx.CalcAmount().String(), "err", err)
			return nil, eth.NewInvalidParamsError("couldn't format amount")
		}
		ethTx.Value = ethAmount
	}

	htmlcoinTxContractInfo, isContractTx, _ := htmlcoinDecodedRawTx.ExtractContractInfo()
	// parsing err is discarded because it's not an error if the transaction is not a valid contract call
	// https://testnet.htmlcoin.info/tx/24ed3749022ed21e53d8924764bb0303a4b6fa469f26922bfa64ba44507c4c4a
	// if err != nil {
	// 	p.GetDebugLogger().Log("msg", "Couldn't extract contract info", "err", err)
	// 	return nil, eth.NewCallbackError(htmlcoinTx.Hex /*"couldn't extract contract info"*/)
	// }
	if isContractTx {
		// TODO: research is this allowed? ethTx.Input = utils.AddHexPrefix(htmlcoinTxContractInfo.UserInput)
		if htmlcoinTxContractInfo.UserInput == "" {
			ethTx.Input = "0x"
		} else {
			ethTx.Input = utils.AddHexPrefix(htmlcoinTxContractInfo.UserInput)
		}
		if htmlcoinTxContractInfo.From != "" {
			ethTx.From = utils.AddHexPrefix(htmlcoinTxContractInfo.From)
		} else {
			// It seems that ExtractContractInfo only looks for OP_SENDER address when assigning From field, so if none is present we handle it like for a non-contract TX
			ethTx.From, err = getNonContractTxSenderAddress(ctx, p, htmlcoinDecodedRawTx)
			if err != nil {
				p.GetDebugLogger().Log("msg", "Contract tx parsing found no sender address", "tx", htmlcoinDecodedRawTx, "err", err)
				return nil, eth.NewCallbackError("Contract tx parsing found no sender address, and the fallback function also failed: " + err.Error())
			}
		}
		//TODO: research if 'To' adress could be other than zero address when 'isContractTx == TRUE'
		if len(htmlcoinTxContractInfo.To) == 0 {
			ethTx.To = utils.AddHexPrefix(htmlcoin.ZeroAddress)
		} else {
			ethTx.To = utils.AddHexPrefix(htmlcoinTxContractInfo.To)
		}

		// gasLimit
		if len(htmlcoinTxContractInfo.GasLimit) == 0 {
			htmlcoinTxContractInfo.GasLimit = "0"
		}
		ethTx.Gas = utils.AddHexPrefix(htmlcoinTxContractInfo.GasLimit)

		// trim leading zeros from gasPrice
		htmlcoinTxContractInfo.GasPrice = strings.TrimLeft(htmlcoinTxContractInfo.GasPrice, "0")
		if len(htmlcoinTxContractInfo.GasPrice) == 0 {
			htmlcoinTxContractInfo.GasPrice = "0"
		}
		// Gas price is in hex satoshis, convert to wei
		gasPriceInSatoshis, err := utils.DecodeBig(htmlcoinTxContractInfo.GasPrice)
		if err != nil {
			p.GetErrorLogger().Log("msg", "Failed to parse gasPrice: "+htmlcoinTxContractInfo.GasPrice, "error", err.Error())
			return ethTx, eth.NewCallbackError("Failed to parse gasPrice")
		}

		gasPriceInWei := convertFromSatoshiToWei(gasPriceInSatoshis)
		ethTx.GasPrice = hexutil.EncodeBig(gasPriceInWei)

		return ethTx, nil
	}

	if htmlcoinTx.Generated {
		ethTx.From = utils.AddHexPrefix(htmlcoin.ZeroAddress)
	} else {
		// TODO: Figure out if following code still cause issues in some cases, see next comment

		// causes issues on coinbase txs, coinbase will not have a sender and so this should be able to fail
		ethTx.From, _ = getNonContractTxSenderAddress(ctx, p, htmlcoinDecodedRawTx)

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
func getRewardTransactionByHash(ctx context.Context, p *htmlcoin.Htmlcoin, hash string) (*eth.GetTransactionByHashResponse, *htmlcoin.GetRawTransactionResponse, error) {
	rawHtmlcoinTx, err := p.GetRawTransaction(ctx, hash, false)
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

		R: "0xf000000000000000000000000000000000000000000000000000000000000000",
		S: "0xf000000000000000000000000000000000000000000000000000000000000000",
		V: "0x25",
	}

	if rawHtmlcoinTx.IsPending() {
		// geth returns null if the tx is pending
		return nil, rawHtmlcoinTx, nil
	} else {
		blockIndex, err := getTransactionIndexInBlock(ctx, p, hash, rawHtmlcoinTx.BlockHash)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "couldn't get transaction index in block")
		}
		ethTx.TransactionIndex = hexutil.EncodeUint64(uint64(blockIndex))

		blockNumber, err := getBlockNumberByHash(ctx, p, rawHtmlcoinTx.BlockHash)
		if err != nil {
			return nil, nil, errors.WithMessage(err, "couldn't get block number by hash")
		}
		ethTx.BlockNumber = hexutil.EncodeUint64(blockNumber)

		ethTx.BlockHash = utils.AddHexPrefix(rawHtmlcoinTx.BlockHash)
	}

	for i := range rawHtmlcoinTx.Vouts {
		// TODO: discuss
		// ! The response may be null, even if txout is presented
		_, err := p.GetTransactionOut(ctx, hash, i, rawHtmlcoinTx.IsPending())
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
			addresses := vout.Details.GetAddresses()
			addressesCount := len(addresses)
			if addressesCount > 0 && addresses[0] == from {
				refund += vout.AmountSatoshi
			} else {
				if addressesCount > 0 && addresses[0] != "" {
					if to == "" {
						to = addresses[0]
					}
					if to == addresses[0] {
						sentTo += vout.AmountSatoshi
					}
				}
				sent += vout.AmountSatoshi
			}
		}
		fee := valueIn - valueOut
		if fee < 0 {
			// coinbase/coinstake txs have no fees since they are a part of making a block
			fee = 0
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

		if to != "" {
			toAddress, err := p.Base58AddressToHex(to)
			if err == nil {
				ethTx.To = utils.AddHexPrefix(toAddress)
			}
		}

		// TODO: compute gasPrice based on fee, guess a gas amount based on vin/vout
		// gas price is set in the OP_CALL/OP_CREATE script
	}

	return ethTx, rawHtmlcoinTx, nil
}
