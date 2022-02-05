package transformer

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/htmlcoin/janus/pkg/eth"
	"github.com/htmlcoin/janus/pkg/internal"
	"github.com/htmlcoin/janus/pkg/htmlcoin"
)

func TestEstimateGasRequest(t *testing.T) {
	request := eth.CallRequest{
		From: "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		To:   "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		Data: "0x0",
	}
	requestRaw, err := json.Marshal(&request)
	if err != nil {
		t.Fatal(err)
	}
	requestParamsArray := []json.RawMessage{requestRaw}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParamsArray)

	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := internal.NewDoerMappedMock()
	htmlcoinClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//preparing responses
	fromHexAddressResponse := htmlcoin.FromHexAddressResponse("0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960")
	err = mockedClientDoer.AddResponseWithRequestID(2, htmlcoin.MethodFromHexAddress, fromHexAddressResponse)
	if err != nil {
		t.Fatal(err)
	}

	callContractResponse := htmlcoin.CallContractResponse{
		Address: "1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		ExecutionResult: struct {
			GasUsed         int    `json:"gasUsed"`
			Excepted        string `json:"excepted"`
			ExceptedMessage string `json:"exceptedMessage"`
			NewAddress      string `json:"newAddress"`
			Output          string `json:"output"`
			CodeDeposit     int    `json:"codeDeposit"`
			GasRefunded     int    `json:"gasRefunded"`
			DepositSize     int    `json:"depositSize"`
			GasForDeposit   int    `json:"gasForDeposit"`
		}{
			GasUsed:  21678,
			Excepted: "None",
		},
	}
	err = mockedClientDoer.AddResponseWithRequestID(1, htmlcoin.MethodCallContract, callContractResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	proxyEth := ProxyETHCall{htmlcoinClient}
	proxyEthEstimateGas := ProxyETHEstimateGas{&proxyEth}
	got, jsonErr := proxyEthEstimateGas.Request(requestRPC, nil)
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := eth.EstimateGasResponse("0x5d25")
	if !reflect.DeepEqual(got, &want) {
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			request,
			string(internal.MustMarshalIndent(want, "", "  ")),
			string(internal.MustMarshalIndent(got, "", "  ")),
		)
	}
}

func TestEstimateGasRequestExecutionReverted(t *testing.T) {
	request := eth.CallRequest{
		From: "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		To:   "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		Data: "0x0",
	}
	requestRaw, err := json.Marshal(&request)
	if err != nil {
		t.Fatal(err)
	}
	requestParamsArray := []json.RawMessage{requestRaw}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParamsArray)

	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := internal.NewDoerMappedMock()
	htmlcoinClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//preparing responses
	fromHexAddressResponse := htmlcoin.FromHexAddressResponse("0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960")
	err = mockedClientDoer.AddResponseWithRequestID(2, htmlcoin.MethodFromHexAddress, fromHexAddressResponse)
	if err != nil {
		t.Fatal(err)
	}

	callContractResponse := htmlcoin.CallContractResponse{
		Address: "1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		ExecutionResult: struct {
			GasUsed         int    `json:"gasUsed"`
			Excepted        string `json:"excepted"`
			ExceptedMessage string `json:"exceptedMessage"`
			NewAddress      string `json:"newAddress"`
			Output          string `json:"output"`
			CodeDeposit     int    `json:"codeDeposit"`
			GasRefunded     int    `json:"gasRefunded"`
			DepositSize     int    `json:"depositSize"`
			GasForDeposit   int    `json:"gasForDeposit"`
		}{
			GasUsed:  21678,
			Excepted: "OutOfGas",
		},
	}
	err = mockedClientDoer.AddResponseWithRequestID(1, htmlcoin.MethodCallContract, callContractResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	proxyEth := ProxyETHCall{htmlcoinClient}
	proxyEthEstimateGas := ProxyETHEstimateGas{&proxyEth}
	_, got := proxyEthEstimateGas.Request(requestRPC, nil)
	if got == nil {
		t.Fatal("Expected error")
	}

	want := eth.NewCallbackError(ErrExecutionReverted.Error())
	if !reflect.DeepEqual(want, got) {
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			request,
			string(internal.MustMarshalIndent(want, "", "  ")),
			string(internal.MustMarshalIndent(got, "", "  ")),
		)
	}
}

func TestEstimateGasNonVMRequest(t *testing.T) {
	request := eth.CallRequest{
		From: "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		To:   "0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
	}
	requestRaw, err := json.Marshal(&request)
	if err != nil {
		t.Fatal(err)
	}
	requestParamsArray := []json.RawMessage{requestRaw}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParamsArray)

	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := internal.NewDoerMappedMock()
	htmlcoinClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//preparing responses
	fromHexAddressResponse := htmlcoin.FromHexAddressResponse("0x1e6f89d7399081b4f8f8aa1ae2805a5efff2f960")
	err = mockedClientDoer.AddResponseWithRequestID(2, htmlcoin.MethodFromHexAddress, fromHexAddressResponse)
	if err != nil {
		t.Fatal(err)
	}

	callContractResponse := htmlcoin.CallContractResponse{
		Address: "1e6f89d7399081b4f8f8aa1ae2805a5efff2f960",
		ExecutionResult: struct {
			GasUsed         int    `json:"gasUsed"`
			Excepted        string `json:"excepted"`
			ExceptedMessage string `json:"exceptedMessage"`
			NewAddress      string `json:"newAddress"`
			Output          string `json:"output"`
			CodeDeposit     int    `json:"codeDeposit"`
			GasRefunded     int    `json:"gasRefunded"`
			DepositSize     int    `json:"depositSize"`
			GasForDeposit   int    `json:"gasForDeposit"`
		}{
			GasUsed:  21678,
			Excepted: "None",
		},
	}
	err = mockedClientDoer.AddResponseWithRequestID(1, htmlcoin.MethodCallContract, callContractResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	proxyEth := ProxyETHCall{htmlcoinClient}
	proxyEthEstimateGas := ProxyETHEstimateGas{&proxyEth}
	got, jsonErr := proxyEthEstimateGas.Request(requestRPC, nil)
	if jsonErr != nil {
		t.Fatal(jsonErr)
	}

	want := eth.EstimateGasResponse(NonContractVMGasLimit)
	if !reflect.DeepEqual(got, &want) {
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			request,
			string(internal.MustMarshalIndent(want, "", "  ")),
			string(internal.MustMarshalIndent(got, "", "  ")),
		)
	}
}
