package transformer

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/htmlcoin/janus/pkg/internal"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

type ETHProxyInitializer = func(*htmlcoin.Htmlcoin) ETHProxy

func testETHProxyRequest(t *testing.T, initializer ETHProxyInitializer, requestParams []json.RawMessage, want interface{}) {
	request, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := internal.NewDoerMappedMock()
	htmlcoinClient, err := internal.CreateMockedClient(mockedClientDoer)

	internal.SetupGetBlockByHashResponses(t, mockedClientDoer)

	//preparing proxy & executing request
	proxyEth := initializer(htmlcoinClient)
	got, jsonErr := proxyEth.Request(request, nil)
	if jsonErr != nil {
		t.Fatalf("Failed to process request on %T.Request(%s): %s", proxyEth, requestParams, jsonErr)
	}

	if !reflect.DeepEqual(got, want) {
		wantString := string(internal.MustMarshalIndent(want, "", "  "))
		gotString := string(internal.MustMarshalIndent(got, "", "  "))
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			request,
			wantString,
			gotString,
		)
		if wantString == gotString {
			t.Errorf("Want and Got are equal strings but !DeepEqual, probably differ in types (%T ?= %T)", want, got)
		}
	}
}
