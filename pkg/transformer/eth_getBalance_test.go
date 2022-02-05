package transformer

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/btcsuite/btcutil"
	"github.com/htmlcoin/janus/pkg/internal"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

func TestGetBalanceRequestAccount(t *testing.T) {
	//prepare request
	requestParams := []json.RawMessage{[]byte(`"0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960"`), []byte(`"123"`)}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}
	//prepare client
	mockedClientDoer := internal.NewDoerMappedMock()
	htmlcoinClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//prepare account
	account, err := btcutil.DecodeWIF("5JK4Gu9nxCvsCxiq9Zf3KdmA9ACza6dUn5BRLVWAYEtQabdnJ89")
	if err != nil {
		t.Fatal(err)
	}
	htmlcoinClient.Accounts = append(htmlcoinClient.Accounts, account)

	//prepare responses
	fromHexAddressResponse := htmlcoin.FromHexAddressResponse("5JK4Gu9nxCvsCxiq9Zf3KdmA9ACza6dUn5BRLVWAYEtQabdnJ89")
	err = mockedClientDoer.AddResponseWithRequestID(2, htmlcoin.MethodFromHexAddress, fromHexAddressResponse)
	if err != nil {
		t.Fatal(err)
	}

	getAddressBalanceResponse := htmlcoin.GetAddressBalanceResponse{Balance: uint64(100000000), Received: uint64(100000000), Immature: int64(0)}
	err = mockedClientDoer.AddResponseWithRequestID(3, htmlcoin.MethodGetAddressBalance, getAddressBalanceResponse)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: Need getaccountinfo to return an account for unit test
	// if getaccountinfo returns nil
	// then address is contract, else account

	//preparing proxy & executing request
	proxyEth := ProxyETHGetBalance{htmlcoinClient}
	got, jsonErr := proxyEth.Request(requestRPC, nil)
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := string("0xde0b6b3a7640000") //1 Htmlcoin represented in Wei
	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			requestRPC,
			string(internal.MustMarshalIndent(want, "", "  ")),
			string(internal.MustMarshalIndent(got, "", "  ")),
		)
	}
}

func TestGetBalanceRequestContract(t *testing.T) {
	//prepare request
	requestParams := []json.RawMessage{[]byte(`"0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960"`), []byte(`"123"`)}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}
	//prepare client
	mockedClientDoer := internal.NewDoerMappedMock()
	htmlcoinClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//prepare account
	account, err := btcutil.DecodeWIF("5JK4Gu9nxCvsCxiq9Zf3KdmA9ACza6dUn5BRLVWAYEtQabdnJ89")
	if err != nil {
		t.Fatal(err)
	}
	htmlcoinClient.Accounts = append(htmlcoinClient.Accounts, account)

	//prepare responses
	getAccountInfoResponse := htmlcoin.GetAccountInfoResponse{
		Address: "1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		Balance: 12431243,
		// Storage json.RawMessage `json:"storage"`,
		// Code    string          `json:"code"`,
	}
	err = mockedClientDoer.AddResponseWithRequestID(3, htmlcoin.MethodGetAccountInfo, getAccountInfoResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	proxyEth := ProxyETHGetBalance{htmlcoinClient}
	got, jsonErr := proxyEth.Request(requestRPC, nil)
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := string("0xbdaf8b")
	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			requestRPC,
			string(internal.MustMarshalIndent(want, "", "  ")),
			string(internal.MustMarshalIndent(got, "", "  ")),
		)
	}
}
