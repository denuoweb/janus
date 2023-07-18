package transformer

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/denuoweb/janus/pkg/eth"
	"github.com/denuoweb/janus/pkg/internal"
	"github.com/denuoweb/janus/pkg/htmlcoin"
)

func TestHashrateRequest(t *testing.T) {
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

	exampleResponse := `{"enabled": true, "staking": false, "errors": "", "currentblocktx": 0, "pooledtx": 0, "difficulty": 4.656542373906925e-010, "search-interval": 0, "weight": 0, "netstakeweight": 0, "expectedtime": 0}`
	getHashrateResponse := htmlcoin.GetHashrateResponse{}
	unmarshalRequest([]byte(exampleResponse), &getHashrateResponse)

	err = mockedClientDoer.AddResponse(htmlcoin.MethodGetStakingInfo, getHashrateResponse)
	if err != nil {
		t.Fatal(err)
	}

	proxyEth := ProxyETHHashrate{htmlcoinClient}
	got, jsonErr := proxyEth.Request(request, internal.NewEchoContext())
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	expected := hexutil.EncodeUint64(math.Float64bits(4.656542373906925e-010))
	want := eth.HashrateResponse(expected)

	internal.CheckTestResultEthRequestRPC(*request, &want, got, t, false)
}
