package transformer

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/internal"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

func TestMiningRequest(t *testing.T) {
	//preparing the request
	requestParams := []json.RawMessage{} //eth_hashrate has no params
	request, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := internal.NewDoerMappedMock()
	htmlcoinClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	getMiningResponse := htmlcoin.GetMiningResponse{Staking: true}
	err = mockedClientDoer.AddResponse(htmlcoin.MethodGetStakingInfo, getMiningResponse)
	if err != nil {
		t.Fatal(err)
	}

	proxyEth := ProxyETHMining{htmlcoinClient}
	got, jsonErr := proxyEth.Request(request, nil)
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := eth.MiningResponse(true)
	if !reflect.DeepEqual(got, &want) {
		t.Errorf(
			"error\ninput: %s\nwant: %t\ngot: %t",
			request,
			want,
			got,
		)
	}

}
