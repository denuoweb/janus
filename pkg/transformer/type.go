package transformer

import (
	"errors"

	"github.com/labstack/echo"
	"github.com/htmlcoin/janus/pkg/eth"
)

var UnmarshalRequestErr = errors.New("Input is invalid")

type Option func(*Transformer) error

type ETHProxy interface {
	Request(*eth.JSONRPCRequest, echo.Context) (interface{}, eth.JSONRPCError)
	Method() string
}
