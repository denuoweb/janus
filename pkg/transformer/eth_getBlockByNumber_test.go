package transformer

import (
	"encoding/json"
	"testing"

	"github.com/denuoweb/janus/pkg/eth"
	"github.com/denuoweb/janus/pkg/internal"
	"github.com/denuoweb/janus/pkg/htmlcoin"
)

func initializeProxyETHGetBlockByNumber(htmlcoinClient *htmlcoin.Htmlcoin) ETHProxy {
	return &ProxyETHGetBlockByNumber{htmlcoinClient}
}

func TestGetBlockByNumberRequest(t *testing.T) {
	testETHProxyRequest(
		t,
		initializeProxyETHGetBlockByNumber,
		[]json.RawMessage{[]byte(`"` + internal.GetTransactionByHashBlockNumberHex + `"`), []byte(`false`)},
		&internal.GetTransactionByHashResponse,
	)
}

func TestGetBlockByNumberWithTransactionsRequest(t *testing.T) {
	testETHProxyRequest(
		t,
		initializeProxyETHGetBlockByNumber,
		[]json.RawMessage{[]byte(`"` + internal.GetTransactionByHashBlockNumberHex + `"`), []byte(`true`)},
		&internal.GetTransactionByHashResponseWithTransactions,
	)
}

func TestGetBlockByNumberUnknownBlockRequest(t *testing.T) {
	requestParams := []json.RawMessage{[]byte(`"` + internal.GetTransactionByHashBlockNumberHex + `"`), []byte(`true`)}
	request, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := internal.NewDoerMappedMock()
	htmlcoinClient, err := internal.CreateMockedClient(mockedClientDoer)

	unknownBlockResponse := htmlcoin.GetErrorResponse(htmlcoin.ErrInvalidParameter)
	err = mockedClientDoer.AddError(htmlcoin.MethodGetBlockHash, unknownBlockResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	proxyEth := ProxyETHGetBlockByNumber{htmlcoinClient}
	got, jsonErr := proxyEth.Request(request, internal.NewEchoContext())
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := (*eth.GetBlockByNumberResponse)(nil)

	internal.CheckTestResultDefault(want, got, t, false)
}
