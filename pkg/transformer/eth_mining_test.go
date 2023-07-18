package transformer

import (
	"encoding/json"
	"testing"

	"github.com/denuoweb/janus/pkg/eth"
	"github.com/denuoweb/janus/pkg/internal"
	"github.com/denuoweb/janus/pkg/htmlcoin"
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
	got, jsonErr := proxyEth.Request(request, internal.NewEchoContext())
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := eth.MiningResponse(true)

	internal.CheckTestResultEthRequestRPC(*request, &want, got, t, false)
}
