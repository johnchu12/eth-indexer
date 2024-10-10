package ethclient

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// GetLogsByBlockNumber retrieves logs from the Ethereum blockchain within a specified block range.
// It filters logs based on the provided addresses.
func (c *Client) GetLogsByBlockNumber(ctx context.Context, fromNumber, endNumber *big.Int, addresses []common.Address) ([]types.Log, error) {
	query := ethereum.FilterQuery{
		Addresses: addresses,
		Topics:    [][]common.Hash{},
		FromBlock: fromNumber,
		ToBlock:   endNumber,
	}

	// Use the provided context instead of creating a new one
	logs, err := c.Client.FilterLogs(ctx, query)
	if err != nil {
		return nil, err
	}

	return logs, nil
}
