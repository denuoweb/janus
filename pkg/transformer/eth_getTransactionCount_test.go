package transformer

import (
	"encoding/json"
	"testing"

	"github.com/denuoweb/janus/pkg/internal"
)

func TestGetTransactionCountRequest(t *testing.T) {
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

	//preparing proxy & executing request
	proxyEth := ProxyETHTxCount{htmlcoinClient}
	got, jsonErr := proxyEth.Request(request, internal.NewEchoContext())
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := string("0x1") //tx count is hardcoded inside the implement

	internal.CheckTestResultEthRequestRPC(*request, want, got, t, false)
}
