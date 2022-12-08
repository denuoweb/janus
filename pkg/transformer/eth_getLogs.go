package transformer

import (
	"context"
	"encoding/json"

	"github.com/labstack/echo"
	"github.com/htmlcoin/janus/pkg/conversion"
	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
	"github.com/htmlcoin/janus/pkg/utils"
)

// ProxyETHGetLogs implements ETHProxy
type ProxyETHGetLogs struct {
	*htmlcoin.Htmlcoin
}

func (p *ProxyETHGetLogs) Method() string {
	return "eth_getLogs"
}

func (p *ProxyETHGetLogs) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var req eth.GetLogsRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	// TODO: Graph Node is sending the topic
	// if len(req.Topics) != 0 {
	// 	return nil, errors.New("topics is not supported yet")
	// }

	// Calls ToRequest in order transform ETH-Request to a Htmlcoin-Request
	htmlcoinreq, err := p.ToRequest(c.Request().Context(), &req)
	if err != nil {
		return nil, err
	}

	return p.request(c.Request().Context(), htmlcoinreq)
}

func (p *ProxyETHGetLogs) request(ctx context.Context, req *htmlcoin.SearchLogsRequest) (*eth.GetLogsResponse, eth.JSONRPCError) {
	receipts, err := conversion.SearchLogsAndFilterExtraTopics(ctx, p.Htmlcoin, req)
	if err != nil {
		return nil, err
	}

	logs := make([]eth.Log, 0)
	for _, receipt := range receipts {
		r := htmlcoin.TransactionReceipt(receipt)
		logs = append(logs, conversion.ExtractETHLogsFromTransactionReceipt(r, r.Log)...)
	}

	resp := eth.GetLogsResponse(logs)
	return &resp, nil
}

func (p *ProxyETHGetLogs) ToRequest(ctx context.Context, ethreq *eth.GetLogsRequest) (*htmlcoin.SearchLogsRequest, eth.JSONRPCError) {
	//transform EthRequest fromBlock to HtmlcoinReq fromBlock:
	from, err := getBlockNumberByRawParam(ctx, p.Htmlcoin, ethreq.FromBlock, true)
	if err != nil {
		return nil, err
	}

	//transform EthRequest toBlock to HtmlcoinReq toBlock:
	to, err := getBlockNumberByRawParam(ctx, p.Htmlcoin, ethreq.ToBlock, true)
	if err != nil {
		return nil, err
	}

	//transform EthReq address to HtmlcoinReq address:
	var addresses []string
	if ethreq.Address != nil {
		if isBytesOfString(ethreq.Address) {
			var addr string
			if jsonErr := json.Unmarshal(ethreq.Address, &addr); jsonErr != nil {
				return nil, eth.NewInvalidParamsError(jsonErr.Error())
			}
			addresses = append(addresses, addr)
		} else {
			if jsonErr := json.Unmarshal(ethreq.Address, &addresses); jsonErr != nil {
				return nil, eth.NewInvalidParamsError(jsonErr.Error())
			}
		}
		for i := range addresses {
			addresses[i] = utils.RemoveHexPrefix(addresses[i])
		}
	}

	//transform EthReq topics to HtmlcoinReq topics:
	topics, topicsErr := eth.TranslateTopics(ethreq.Topics)
	if topicsErr != nil {
		return nil, eth.NewCallbackError(topicsErr.Error())
	}

	return &htmlcoin.SearchLogsRequest{
		Addresses: addresses,
		FromBlock: from,
		ToBlock:   to,
		Topics:    htmlcoin.NewSearchLogsTopics(topics),
	}, nil
}
