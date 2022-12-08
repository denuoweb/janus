package transformer

import (
	"fmt"
	"testing"

	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/internal"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
	"github.com/htmlcoin/janus/pkg/utils"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestEthValueToHtmlcoinAmount(t *testing.T) {
	cases := []map[string]interface{}{
		{
			"in":   "0xde0b6b3a7640000",
			"want": decimal.NewFromFloat(1),
		},
		{

			"in":   "0x6f05b59d3b20000",
			"want": decimal.NewFromFloat(0.5),
		},
		{
			"in":   "0x2540be400",
			"want": decimal.NewFromFloat(0.00000001),
		},
		{
			"in":   "0x1",
			"want": decimal.NewFromInt(0),
		},
	}
	for _, c := range cases {
		in := c["in"].(string)
		want := c["want"].(decimal.Decimal)
		got, err := EthValueToHtmlcoinAmount(in, MinimumGas)
		if err != nil {
			t.Error(err)
		}

		// TODO: Refactor to use new testing utilities?
		if !got.Equal(want) {
			t.Errorf("in: %s, want: %v, got: %v", in, want, got)
		}
	}
}

func TestHtmlcoinValueToEthAmount(t *testing.T) {
	cases := []decimal.Decimal{
		decimal.NewFromFloat(1),
		decimal.NewFromFloat(0.5),
		decimal.NewFromFloat(0.00000001),
		MinimumGas,
	}
	for _, c := range cases {
		in := c
		eth := HtmlcoinDecimalValueToETHAmount(in)
		out := EthDecimalValueToHtmlcoinAmount(eth)

		// TODO: Refactor to use new testing utilities?
		if !in.Equals(out) {
			t.Errorf("in: %s, eth: %v, htmlcoin: %v", in, eth, out)
		}
	}
}

func TestHtmlcoinAmountToEthValue(t *testing.T) {
	in, want := decimal.NewFromFloat(0.1), "0x16345785d8a0000"
	got, err := formatHtmlcoinAmount(in)
	if err != nil {
		t.Error(err)
	}

	internal.CheckTestResultUnspecifiedInputMarshal(in, want, got, t, false)
}

func TestLowestHtmlcoinAmountToEthValue(t *testing.T) {
	in, want := decimal.NewFromFloat(0.00000001), "0x2540be400"
	got, err := formatHtmlcoinAmount(in)
	if err != nil {
		t.Error(err)
	}

	internal.CheckTestResultUnspecifiedInputMarshal(in, want, got, t, false)
}

func TestAddressesConversion(t *testing.T) {
	t.Parallel()

	inputs := []struct {
		htmlcoinChain   string
		ethAddress  string
		htmlcoinAddress string
	}{
		{
			htmlcoinChain:   htmlcoin.ChainTest,
			ethAddress:  "6c89a1a6ca2ae7c00b248bb2832d6f480f27da68",
			htmlcoinAddress: "qTTH1Yr2eKCuDLqfxUyBLCAjmomQ8pyrBt",
		},

		// Test cases for addresses defined here:
		// 	- https://github.com/hayeah/openzeppelin-solidity/blob/htmlcoin/HTMLCOIN-NOTES.md#create-test-accounts
		//
		// NOTE: Ethereum addresses are without `0x` prefix, as it expects by conversion functions
		{
			htmlcoinChain:   htmlcoin.ChainTest,
			ethAddress:  "7926223070547d2d15b2ef5e7383e541c338ffe9",
			htmlcoinAddress: "qUbxboqjBRp96j3La8D1RYkyqx5uQbJPoW",
		},
		{
			htmlcoinChain:   htmlcoin.ChainTest,
			ethAddress:  "2352be3db3177f0a07efbe6da5857615b8c9901d",
			htmlcoinAddress: "qLn9vqbr2Gx3TsVR9QyTVB5mrMoh4x43Uf",
		},
		{
			htmlcoinChain:   htmlcoin.ChainTest,
			ethAddress:  "69b004ac2b3993bf2fdf56b02746a1f57997420d",
			htmlcoinAddress: "qTCCy8qy7pW94EApdoBjYc1vQ2w68UnXPi",
		},
		{
			htmlcoinChain:   htmlcoin.ChainTest,
			ethAddress:  "8c647515f03daeefd09872d7530fa8d8450f069a",
			htmlcoinAddress: "qWMi6ne9mDQFatRGejxdDYVUV9rQVkAFGp",
		},
		{
			htmlcoinChain:   htmlcoin.ChainTest,
			ethAddress:  "2191744eb5ebeac90e523a817b77a83a0058003b",
			htmlcoinAddress: "qLcshhsRS6HKeTKRYFdpXnGVZxw96QQcfm",
		},
		{
			htmlcoinChain:   htmlcoin.ChainTest,
			ethAddress:  "88b0bf4b301c21f8a47be2188bad6467ad556dcf",
			htmlcoinAddress: "qW28njWueNpBXYWj2KDmtFG2gbLeALeHfV",
		},
	}

	for i, in := range inputs {
		var (
			in       = in
			testDesc = fmt.Sprintf("#%d", i)
		)
		// TODO: Investigate why this testing setup is so different
		t.Run(testDesc, func(t *testing.T) {
			htmlcoinAddress, err := convertETHAddress(in.ethAddress, in.htmlcoinChain)
			require.NoError(t, err, "couldn't convert Ethereum address to Htmlcoin address")
			require.Equal(t, in.htmlcoinAddress, htmlcoinAddress, "unexpected converted Htmlcoin address value")

			ethAddress, err := utils.ConvertHtmlcoinAddress(in.htmlcoinAddress)
			require.NoError(t, err, "couldn't convert Htmlcoin address to Ethereum address")
			require.Equal(t, in.ethAddress, ethAddress, "unexpected converted Ethereum address value")
		})
	}
}

func TestSendTransactionRequestHasDefaultGasPriceAndAmount(t *testing.T) {
	var req eth.SendTransactionRequest
	err := unmarshalRequest([]byte(`[{}]`), &req)
	if err != nil {
		t.Fatal(err)
	}
	defaultGasPriceInWei := req.GasPrice.Int
	defaultGasPriceInHTMLCOIN := EthDecimalValueToHtmlcoinAmount(decimal.NewFromBigInt(defaultGasPriceInWei, 1))

	// TODO: Refactor to use new testing utilities?
	if !defaultGasPriceInHTMLCOIN.Equals(MinimumGas) {
		t.Fatalf("Default gas price does not convert to HTMLCOIN minimum gas price, got: %s want: %s", defaultGasPriceInHTMLCOIN.String(), MinimumGas.String())
	}
	if eth.DefaultGasAmountForHtmlcoin.String() != req.Gas.Int.String() {
		t.Fatalf("Default gas amount does not match expected default, got: %s want: %s", req.Gas.Int.String(), eth.DefaultGasAmountForHtmlcoin.String())
	}
}
