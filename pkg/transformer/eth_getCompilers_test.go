package transformer

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/htmlcoin/janus/pkg/internal"
)

func TestGetCompilersReturnsEmptyArray(t *testing.T) {
	//preparing the request
	requestParams := []json.RawMessage{} //eth_getCompilers has no params
	request, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}

	proxyEth := ETHGetCompilers{}
	got, jsonErr := proxyEth.Request(request, nil)
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	if fmt.Sprintf("%v", got) != "[]" {
		t.Errorf(
			"error\nwant: '[]'\ngot: '%v'",
			got,
		)
	}
}
