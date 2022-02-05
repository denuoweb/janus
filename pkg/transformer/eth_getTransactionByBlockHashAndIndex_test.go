package transformer

import (
	"encoding/json"
	"testing"

	"github.com/htmlcoin/janus/pkg/internal"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
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
