package transformer

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/internal"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

func TestChainIdMainnet(t *testing.T) {
	testChainIdsImpl(t, "main", "0x115C")
}

func TestChainIdTestnet(t *testing.T) {
	testChainIdsImpl(t, "test", "0x115D")
}

func TestChainIdRegtest(t *testing.T) {
	testChainIdsImpl(t, "regtest", "0x115E")
}

func TestChainIdUnknown(t *testing.T) {
	testChainIdsImpl(t, "???", "0x115E")
}

func testChainIdsImpl(t *testing.T, chain string, expected string) {
	//preparing request
	requestParams := []json.RawMessage{}
	request, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := internal.NewDoerMappedMock()
	htmlcoinClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//preparing client response
	getBlockCountResponse := htmlcoin.GetBlockChainInfoResponse{Chain: chain}
	err = mockedClientDoer.AddResponseWithRequestID(2, htmlcoin.MethodGetBlockChainInfo, getBlockCountResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	proxyEth := ProxyETHChainId{htmlcoinClient}
	got, jsonErr := proxyEth.Request(request, nil)
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := eth.ChainIdResponse(expected)
	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			request,
			want,
			got,
		)
	}
}
