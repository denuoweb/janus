package transformer

import (
	"github.com/labstack/echo"
	"github.com/denuoweb/janus/pkg/eth"
	"github.com/denuoweb/janus/pkg/htmlcoin"
	"github.com/denuoweb/janus/pkg/utils"
)

// ProxyETHAccounts implements ETHProxy
type ProxyETHAccounts struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHAccounts) Method() string {
	return "eth_accounts"
}

func (p *ProxyETHAccounts) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	return p.request()
}

func (p *ProxyETHAccounts) request() (eth.AccountsResponse, eth.JSONRPCError) {
	var accounts eth.AccountsResponse

	for _, acc := range p.Accounts {
		acc := htmlcoin.Account{acc}
		addr := acc.ToHexAddress()

		accounts = append(accounts, utils.AddHexPrefix(addr))
	}

	return accounts, nil
}

func (p *ProxyETHAccounts) ToResponse(ethresp *htmlcoin.CallContractResponse) *eth.CallResponse {
	data := utils.AddHexPrefix(ethresp.ExecutionResult.Output)
	htmlcoinresp := eth.CallResponse(data)
	return &htmlcoinresp
}
