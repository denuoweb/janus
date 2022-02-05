package transformer

import (
	"encoding/json"
	"testing"

	"github.com/htmlcoin/janus/pkg/internal"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
	"github.com/htmlcoin/janus/pkg/utils"
)

func initializeProxyETHGetBlockByHash(htmlcoinClient *htmlcoin.Htmlcoin) ETHProxy {
	return &ProxyETHGetBlockByHash{htmlcoinClient}
}

func TestGetBlockByHashRequestNonceLength(t *testing.T) {
	if len(utils.RemoveHexPrefix(internal.GetTransactionByHashResponse.Nonce)) != 16 {
		t.Errorf("Nonce test data should be zero left padded length 16")
	}
}

func TestGetBlockByHashRequest(t *testing.T) {
	testETHProxyRequest(
		t,
		initializeProxyETHGetBlockByHash,
		[]json.RawMessage{[]byte(`"` + internal.GetTransactionByHashBlockHexHash + `"`), []byte(`false`)},
		&internal.GetTransactionByHashResponse,
	)
}

func TestGetBlockByHashTransactionsRequest(t *testing.T) {
	testETHProxyRequest(
		t,
		initializeProxyETHGetBlockByHash,
		[]json.RawMessage{[]byte(`"` + internal.GetTransactionByHashBlockHexHash + `"`), []byte(`true`)},
		&internal.GetTransactionByHashResponseWithTransactions,
	)
}
