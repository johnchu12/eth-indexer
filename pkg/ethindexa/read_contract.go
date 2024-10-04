package ethindexa

import (
	"context"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// ReadContract reads the contract and returns the result of the function call.
func ReadContract(client *ethclient.Client, contractAddress common.Address, contractABI abi.ABI, startBlock *big.Int, functionName string, functionParams ...interface{}) (interface{}, error) {
	method, exists := contractABI.Methods[functionName]
	if !exists {
		return nil, fmt.Errorf("function %s not found in ABI", functionName)
	}

	var data []byte
	var err error

	if len(method.Inputs) == 0 {
		data, err = contractABI.Pack(functionName)
	} else {
		if len(functionParams) != len(method.Inputs) {
			return nil, fmt.Errorf("parameter count mismatch: expected %d, got %d", len(method.Inputs), len(functionParams))
		}
		data, err = contractABI.Pack(functionName, functionParams...)
	}

	if err != nil {
		return nil, fmt.Errorf("error packing function call data: %w", err)
	}

	msg := ethereum.CallMsg{
		To:   &contractAddress,
		Data: data,
	}

	result, err := client.CallContract(context.Background(), msg, startBlock)
	if err != nil {
		return nil, fmt.Errorf("error calling contract: %w", err)
	}

	unpackedResult, err := method.Outputs.Unpack(result)
	if err != nil {
		return nil, fmt.Errorf("error unpacking result: %w", err)
	}

	return unpackedResult, nil
}
