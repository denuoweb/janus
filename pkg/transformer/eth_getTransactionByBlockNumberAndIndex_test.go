package transformer

import (
	"encoding/json"
	"testing"

	"github.com/htmlcoin/janus/pkg/internal"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

func initializeProxyETHGetTransactionByBlockNumberAndIndex(htmlcoinClient *htmlcoin.Htmlcoin) ETHProxy {
	return &ProxyETHGetTransactionByBlockNumberAndIndex{htmlcoinClient}
}

func TestGetTransactionByBlockNumberAndIndex(t *testing.T) {
	testETHProxyRequest(
		t,
		initializeProxyETHGetTransactionByBlockNumberAndIndex,
		[]json.RawMessage{[]byte(`"` + internal.GetTransactionByHashBlockNumberHex + `"`), []byte(`"0x0"`)},
		internal.GetTransactionByHashResponseData,
	)
}
