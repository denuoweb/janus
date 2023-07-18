package transformer

import (
	"encoding/json"
	"testing"

	"github.com/denuoweb/janus/pkg/internal"
	"github.com/denuoweb/janus/pkg/htmlcoin"
)

func initializeProxyETHGetTransactionByBlockHashAndIndex(htmlcoinClient *htmlcoin.Htmlcoin) ETHProxy {
	return &ProxyETHGetTransactionByBlockHashAndIndex{htmlcoinClient}
}

func TestGetTransactionByBlockHashAndIndex(t *testing.T) {
	testETHProxyRequest(
		t,
		initializeProxyETHGetTransactionByBlockHashAndIndex,
		[]json.RawMessage{[]byte(`"` + internal.GetTransactionByHashBlockHash + `"`), []byte(`"0x0"`)},
		internal.GetTransactionByHashResponseData,
	)
}
