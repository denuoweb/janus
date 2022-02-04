package transformer

import (
	"github.com/labstack/echo"
	"github.com/htmlcoin/janus/pkg/eth"
)

// Web3ClientVersion implements web3_clientVersion
type Web3ClientVersion struct {
	// *htmlcoin.Htmlcoin
}

func (p *Web3ClientVersion) Method() string {
	return "web3_clientVersion"
}

func (p *Web3ClientVersion) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	return "HTMLCOIN ETHTestRPC/ethereum-js", nil
}

// func (p *Web3ClientVersion) ToResponse(ethresp *htmlcoin.CallContractResponse) *eth.CallResponse {
// 	data := utils.AddHexPrefix(ethresp.ExecutionResult.Output)
// 	htmlcoinresp := eth.CallResponse(data)
// 	return &htmlcoinresp
// }
