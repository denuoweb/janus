package transformer

import (
	"context"
	"fmt"

	"github.com/labstack/echo"
	"github.com/denuoweb/janus/pkg/eth"
	"github.com/denuoweb/janus/pkg/htmlcoin"
	"github.com/denuoweb/janus/pkg/utils"
)

// ProxyETHGetStorageAt implements ETHProxy
type ProxyETHGetStorageAt struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHGetStorageAt) Method() string {
	return "eth_getStorageAt"
}

func (p *ProxyETHGetStorageAt) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var req eth.GetStorageRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	htmlcoinAddress := utils.RemoveHexPrefix(req.Address)
	blockNumber, err := getBlockNumberByParam(c.Request().Context(), p.Htmlcoin, req.BlockNumber, false)
	if err != nil {
		p.GetDebugLogger().Log("msg", fmt.Sprintf("Failed to get block number by param for '%s'", req.BlockNumber), "err", err)
		return nil, err
	}

	return p.request(
		c.Request().Context(),
		&htmlcoin.GetStorageRequest{
			Address:     htmlcoinAddress,
			BlockNumber: blockNumber,
		},
		utils.RemoveHexPrefix(req.Index),
	)
}

func (p *ProxyETHGetStorageAt) request(ctx context.Context, ethreq *htmlcoin.GetStorageRequest, index string) (*eth.GetStorageResponse, eth.JSONRPCError) {
	htmlcoinresp, err := p.Htmlcoin.GetStorage(ctx, ethreq)
	if err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	// htmlcoin res -> eth res
	return p.ToResponse(htmlcoinresp, index), nil
}

func (p *ProxyETHGetStorageAt) ToResponse(htmlcoinresp *htmlcoin.GetStorageResponse, slot string) *eth.GetStorageResponse {
	// the value for unknown anything
	storageData := eth.GetStorageResponse("0x0000000000000000000000000000000000000000000000000000000000000000")
	if len(slot) != 64 {
		slot = leftPadStringWithZerosTo64Bytes(slot)
	}
	for _, outerValue := range *htmlcoinresp {
		htmlcoinStorageData, ok := outerValue[slot]
		if ok {
			storageData = eth.GetStorageResponse(utils.AddHexPrefix(htmlcoinStorageData))
			return &storageData
		}
	}

	return &storageData
}

// left pad a string with leading zeros to fit 64 bytes
func leftPadStringWithZerosTo64Bytes(hex string) string {
	return fmt.Sprintf("%064v", hex)
}
