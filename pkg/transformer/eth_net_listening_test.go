package transformer

import (
	"encoding/json"
	"testing"

	"github.com/htmlcoin/janus/pkg/internal"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

func TestNetListeningInactive(t *testing.T) {
	testNetListeningRequest(t, false)
}

func TestNetListeningActive(t *testing.T) {
	testNetListeningRequest(t, true)
}

func testNetListeningRequest(t *testing.T, active bool) {
	//preparing the request
	requestParams := []json.RawMessage{} //net_listening has no params
	request, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := internal.NewDoerMappedMock()
	htmlcoinClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	networkInfoResponse := htmlcoin.NetworkInfoResponse{NetworkActive: active}
	err = mockedClientDoer.AddResponseWithRequestID(2, htmlcoin.MethodGetNetworkInfo, networkInfoResponse)
	if err != nil {
		t.Fatal(err)
	}

	proxyEth := ProxyNetListening{htmlcoinClient}
	got, jsonErr := proxyEth.Request(request, nil)
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := active
	if want != got {
		t.Errorf(
			"error\nwant: %t\ngot: %t",
			want,
			got,
		)
	}
}
