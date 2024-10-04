package utils

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"hw/internal/indexer/abis"
	"hw/internal/model"
	"hw/pkg/bigrat"
	"hw/pkg/logger"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
	"golang.org/x/sync/errgroup"
)

// LoadABI loads and parses the specified ABI file.
//
//	abiName: the name of the ABI file without the .json extension.
func LoadABI(abiName string) (abi.ABI, error) {
	abiPath := abiName + ".json"

	abiBytes, err := abis.AbisFs.ReadFile(abiPath)
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to read ABI file %s: %w", abiName, err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to parse ABI: %w", err)
	}

	return parsedABI, nil
}

func GetTokenInfo(ctx context.Context, client *ethclient.Client, tokenId string, blockNumber int64) (*model.Token, error) {
	token := &model.Token{ID: tokenId}
	g, _ := errgroup.WithContext(ctx)
	abi, err := LoadABI("erc20_usdc")
	if err != nil {
		logger.Errorw("Error loading USDC ABI:", err)
		return nil, err
	}
	g.Go(func() error {
		result, err := ReadContract(
			client,
			common.HexToAddress(tokenId),
			abi,
			big.NewInt(blockNumber),
			"decimals",
		)
		if err != nil {
			return fmt.Errorf("error reading token contract(decimals): %+v", err)
		}
		token.Decimals = bigrat.NewBigN(result.([]interface{})[0]).ToTruncateInt64(0)
		return nil
	})
	g.Go(func() error {
		result, err := ReadContract(
			client,
			common.HexToAddress(tokenId),
			abi,
			big.NewInt(blockNumber),
			"name",
		)
		if err != nil {
			return fmt.Errorf("error reading USDC contract: %+v", err)
		}
		token.Name = result.([]interface{})[0].(string)
		return nil
	})
	g.Go(func() error {
		result, err := ReadContract(
			client,
			common.HexToAddress(tokenId),
			abi,
			big.NewInt(blockNumber),
			"symbol",
		)
		if err != nil {
			return fmt.Errorf("error reading USDC contract: %+v", err)
		}
		token.Symbol = result.([]interface{})[0].(string)
		return nil
	})
	if err := g.Wait(); err != nil {
		logger.Errorw("Error reading USDC contract:", err)
		return nil, err
	}
	return token, nil
}

// ReadContract reads data from the specified contract.
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
