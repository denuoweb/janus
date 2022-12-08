package transformer

import (
	"github.com/go-kit/kit/log"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/notifier"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

type Transformer struct {
	htmlcoinClient   *htmlcoin.Htmlcoin
	debugMode    bool
	logger       log.Logger
	transformers map[string]ETHProxy
}

// New creates a new Transformer
func New(htmlcoinClient *htmlcoin.Htmlcoin, proxies []ETHProxy, opts ...Option) (*Transformer, error) {
	if htmlcoinClient == nil {
		return nil, errors.New("htmlcoinClient cannot be nil")
	}

	t := &Transformer{
		htmlcoinClient: htmlcoinClient,
		logger:     log.NewNopLogger(),
	}

	var err error
	for _, p := range proxies {
		if err = t.Register(p); err != nil {
			return nil, err
		}
	}

	for _, opt := range opts {
		if err := opt(t); err != nil {
			return nil, err
		}
	}

	return t, nil
}

// Register registers an ETHProxy to a Transformer
func (t *Transformer) Register(p ETHProxy) error {
	if t.transformers == nil {
		t.transformers = make(map[string]ETHProxy)
	}

	m := p.Method()
	if _, ok := t.transformers[m]; ok {
		return errors.Errorf("method already exist: %s ", m)
	}

	t.transformers[m] = p

	return nil
}

// Transform takes a Transformer and transforms the request from ETH request and returns the proxy request
func (t *Transformer) Transform(req *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	proxy, err := t.getProxy(req.Method)
	if err != nil {
		return nil, err
	}
	resp, err := proxy.Request(req, c)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (t *Transformer) getProxy(method string) (ETHProxy, eth.JSONRPCError) {
	proxy, ok := t.transformers[method]
	if !ok {
		return nil, eth.NewMethodNotFoundError(method)
	}
	return proxy, nil
}

func (t *Transformer) IsDebugEnabled() bool {
	return t.debugMode
}

// DefaultProxies are the default proxy methods made available
func DefaultProxies(htmlcoinRPCClient *htmlcoin.Htmlcoin, agent *notifier.Agent) []ETHProxy {
	filter := eth.NewFilterSimulator()
	getFilterChanges := &ProxyETHGetFilterChanges{Htmlcoin: htmlcoinRPCClient, filter: filter}
	ethCall := &ProxyETHCall{Htmlcoin: htmlcoinRPCClient}

	ethProxies := []ETHProxy{
		ethCall,
		&ProxyNetListening{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHPersonalUnlockAccount{},
		&ProxyETHChainId{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHBlockNumber{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHHashrate{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHMining{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHNetVersion{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHGetTransactionByHash{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHGetTransactionByBlockNumberAndIndex{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHGetLogs{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHGetTransactionReceipt{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHSendTransaction{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHAccounts{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHGetCode{Htmlcoin: htmlcoinRPCClient},

		&ProxyETHNewFilter{Htmlcoin: htmlcoinRPCClient, filter: filter},
		&ProxyETHNewBlockFilter{Htmlcoin: htmlcoinRPCClient, filter: filter},
		getFilterChanges,
		&ProxyETHGetFilterLogs{ProxyETHGetFilterChanges: getFilterChanges},
		&ProxyETHUninstallFilter{Htmlcoin: htmlcoinRPCClient, filter: filter},

		&ProxyETHEstimateGas{ProxyETHCall: ethCall},
		&ProxyETHGetBlockByNumber{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHGetBlockByHash{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHGetBalance{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHGetStorageAt{Htmlcoin: htmlcoinRPCClient},
		&ETHGetCompilers{},
		&ETHProtocolVersion{},
		&ETHGetUncleByBlockHashAndIndex{},
		&ETHGetUncleCountByBlockHash{},
		&ETHGetUncleCountByBlockNumber{},
		&Web3ClientVersion{},
		&Web3Sha3{},
		&ProxyETHSign{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHGasPrice{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHTxCount{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHSignTransaction{Htmlcoin: htmlcoinRPCClient},
		&ProxyETHSendRawTransaction{Htmlcoin: htmlcoinRPCClient},

		&ETHSubscribe{Htmlcoin: htmlcoinRPCClient, Agent: agent},
		&ETHUnsubscribe{Htmlcoin: htmlcoinRPCClient, Agent: agent},

		&ProxyHTMLCOINGetUTXOs{Htmlcoin: htmlcoinRPCClient},
		&ProxyHTMLCOINGenerateToAddress{Htmlcoin: htmlcoinRPCClient},

		&ProxyNetPeerCount{Htmlcoin: htmlcoinRPCClient},
	}

	permittedHtmlcoinCalls := []string{
		htmlcoin.MethodGetHexAddress,
		htmlcoin.MethodFromHexAddress,
	}

	for _, htmlcoinMethod := range permittedHtmlcoinCalls {
		ethProxies = append(
			ethProxies,
			&ProxyHTMLCOINGenericStringArguments{
				Htmlcoin:   htmlcoinRPCClient,
				prefix: "dev",
				method: htmlcoinMethod,
			},
		)
	}

	return ethProxies
}

func SetDebug(debug bool) func(*Transformer) error {
	return func(t *Transformer) error {
		t.debugMode = debug
		return nil
	}
}

func SetLogger(l log.Logger) func(*Transformer) error {
	return func(t *Transformer) error {
		t.logger = log.WithPrefix(l, "component", "transformer")
		return nil
	}
}
