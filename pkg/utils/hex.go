package utils

import (
	"encoding/hex"
	"math/big"
	"strings"

	"github.com/decred/base58"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
)

func RemoveHexPrefix(hex string) string {
	if strings.HasPrefix(hex, "0x") {
		return hex[2:]
	}
	return hex
}

func IsEthHexAddress(str string) bool {
	return strings.HasPrefix(str, "0x") || common.IsHexAddress("0x"+str)
}

func AddHexPrefix(hex string) string {
	if strings.HasPrefix(hex, "0x") {
		return hex
	}
	return "0x" + hex
}

func AddHexPrefixIfNotEmpty(hex string) string {
	if hex == "" {
		return hex
	}
	return AddHexPrefix(hex)
}

// DecodeBig decodes a hex string whether input is with 0x prefix or not.
func DecodeBig(input string) (*big.Int, error) {
	input = AddHexPrefix(input)
	return hexutil.DecodeBig(input)
}

// Converts Htmlcoin address to an Ethereum address
func ConvertHtmlcoinAddress(address string) (ethAddress string, _ error) {
	if n := len(address); n < 22 {
		return "", errors.Errorf("invalid address: length is less than 22 bytes - %d", n)
	}

	// Drop Htmlcoin chain prefix and checksum suffix
	ethAddrBytes := base58.Decode(address)[1:21]

	return hex.EncodeToString(ethAddrBytes), nil
}
