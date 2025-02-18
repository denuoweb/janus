package htmlcoin

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"github.com/denuoweb/btcd/txscript"
)

type (
	ContractInvokeInfo struct {
		// VMVersion string
		From     string
		GasLimit string
		GasPrice string
		CallData string
		To       string
	}
)

func ParseCallASM(parts []string) (*ContractInvokeInfo, error) {

	// "4 25548 40 8588b2c50000000000000000000000000000000000000000000000000000000000000000 57946bb437560b13275c32a468c6fd1e0c2cdd48 OP_CAL"

	// 4                     // EVM version
	// 100000                // gas limit
	// 10                    // gas price
	// 1234                  // data to be sent by the contract
	// Contract Address      // contract address
	// OP_CALL

	if len(parts) != 6 {
		return nil, errors.New(fmt.Sprintf("invalid OP_CALL script for parts 6: %v", parts))
	}

	gasLimit, err := convertToBigEndian(parts[1])
	if err != nil {
		return nil, err
	}

	return &ContractInvokeInfo{
		GasPrice: parts[2],
		GasLimit: gasLimit,
		CallData: parts[3],
		To:       parts[4],
	}, nil

}

func ParseCallSenderASM(parts []string) (*ContractInvokeInfo, error) {
	// See: https://github.com/denuoweb/qips/issues/6

	// "1 7926223070547d2d15b2ef5e7383e541c338ffe9 69463043021f3ba540f52e0bae0c608c3d7135424fb683c77ee03217fcfe0af175c586aadc02200222e460a42268f02f130bc46f3ef62f228dd8051756dc13693332423515fcd401210299d391f528b9edd07284c7e23df8415232a8ce41531cf460a390ce32b4efd112 OP_SENDER 4 40000000 40 60fe47b10000000000000000000000000000000000000000000000000000000000000319 9e11fba86ee5d0ba4996b0d1973de6b694f4fc95 OP_CALL"

	if len(parts) != 10 {
		return nil, errors.New(fmt.Sprintf("invalid create_sender script for parts 10: %v", parts))
	}

	// 1    // address type of the pubkeyhash (public key hash)
	// Address               // sender's pubkeyhash address
	// {signature, pubkey}   // serialized scriptSig
	// OP_SENDER
	// 4                     // EVM version
	// 100000                // gas limit
	// 10                    // gas price
	// 1234                  // data to be sent by the contract
	// Contract Address      // contract address
	// OP_CALL

	gasLimit, err := convertToBigEndian(parts[5])
	if err != nil {
		return nil, err
	}

	return &ContractInvokeInfo{
		From:     parts[1],
		GasPrice: parts[6],
		GasLimit: gasLimit,
		CallData: parts[7],
		To:       parts[8],
	}, nil

}

func ParseCreateASM(parts []string) (*ContractInvokeInfo, error) {

	// 4                     // EVM version
	// 100000                // gas limit
	// 10                    // gas price
	// 1234                  // data to be sent by the contract
	// OP_CREATE

	if len(parts) != 5 {
		return nil, errors.New(fmt.Sprintf("invalid OP_CREATE script for parts 5: %v", len(parts)))
	}

	gasLimit, err := convertToBigEndian(parts[1])
	if err != nil {
		return nil, err
	}

	info := &ContractInvokeInfo{
		GasPrice: parts[2],
		GasLimit: gasLimit,
		CallData: parts[3],
	}
	return info, nil

}

func ParseCreateSenderASM(parts []string) (*ContractInvokeInfo, error) {
	// See: https://github.com/denuoweb/qips/issues/6
	// https://blog.htmlcoin.org/qip-5-add-op-sender-opcode-571511802938

	// "1 7926223070547d2d15b2ef5e7383e541c338ffe9 6a473044022067ca66b0308ae16aeca7a205ce0490b44a61feebe5632710b52aabde197f9e4802200e8beec61a58dbe1279a9cdb68983080052ae7b9997bc863b7c5623e4cb55fd
	// b01210299d391f528b9edd07284c7e23df8415232a8ce41531cf460a390ce32b4efd112 OP_SENDER 4 6721975 100 6060604052341561000f57600080fd5b60008054600160a060020a033316600160a060020a03199091161790556101de8061003b6000
	// 396000f300606060405263ffffffff7c01000000000000000000000000000000000000000000000000000000006000350416630900f010811461005d578063445df0ac1461007e5780638da5cb5b146100a3578063fdacd576146100d257600080fd5b341561
	// 006857600080fd5b61007c600160a060020a03600435166100e8565b005b341561008957600080fd5b61009161017d565b60405190815260200160405180910390f35b34156100ae57600080fd5b6100b6610183565b604051600160a060020a039091168152
	// 60200160405180910390f35b34156100dd57600080fd5b61007c600435610192565b6000805433600160a060020a03908116911614156101795781905080600160a060020a031663fdacd5766001546040517c01000000000000000000000000000000000000
	// 0000000000000000000063ffffffff84160281526004810191909152602401600060405180830381600087803b151561016457600080fd5b6102c65a03f1151561017557600080fd5b5050505b5050565b60015481565b600054600160a060020a031681565b
	// 60005433600160a060020a03908116911614156101af5760018190555b505600a165627a7a72305820b6a912c5b5115d1a5412235282372dc4314f325bac71ee6c8bd18f658d7ed1ad0029 OP_CREATE"

	// 1    // address type of the pubkeyhash (public key hash)
	// Address               // sender's pubkeyhash address
	// {signature, pubkey}   //serialized scriptSig
	// OP_SENDER
	// 4                     // EVM version
	// 100000                //gas limit
	// 10                    //gas price
	// 1234                  // data to be sent by the contract
	// OP_CREATE

	if len(parts) != 9 {
		return nil, errors.New(fmt.Sprintf("invalid create_sender script for parts 9: %v", len(parts)))
	}

	gasLimit, err := convertToBigEndian(parts[5])
	if err != nil {
		return nil, err
	}

	info := &ContractInvokeInfo{
		From:     parts[1],
		GasPrice: parts[6],
		GasLimit: gasLimit,
		CallData: parts[7],
	}
	return info, nil
}

// function disasm converts the hex string (from the pubkey) to an asm string
func DisasmScript(scriptHex string) (string, error) {
	scriptBytes, err := hex.DecodeString(scriptHex)
	if err != nil {
		return "", err
	}
	disasm, err := txscript.DisasmString(scriptBytes)
	if err != nil {
		return "", err
	}
	return disasm, nil
}

// convertToBigEndian is a helper function to convert 'gasLimit' hex to big endian
func convertToBigEndian(hex string) (string, error) {
	if len(hex)%2 != 0 {
		return "", fmt.Errorf("invalid hex string")
	}
	var result string
	for i := len(hex); i > 0; i -= 2 {
		result += hex[i-2 : i]
	}
	// trim leading zeros before returning
	return strings.TrimLeft(result, "0"), nil
}
