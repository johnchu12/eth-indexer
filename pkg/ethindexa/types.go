package ethindexa

import (
	"context"
	"fmt"
	"math/big"

	"hw/internal/service"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Event defines the structure of an event.
type Event struct {
	EventName       string
	Args            map[string]interface{}
	TransactionHash common.Hash
	BlockHash       common.Hash
	ContractAddress common.Address
	ContractName    string
	NetworkName     string
	Ctx             context.Context
}

// IndexerService provides access to the Ethereum client and the PostgreSQL database.
type IndexerService struct {
	Client  *ethclient.Client
	Service service.Service
}

// ReadContract is a method of IndexerService used to read contract data.
func (s *IndexerService) ReadContract(contractAddress common.Address, contractABI abi.ABI, startBlock *big.Int, functionName string, functionParams ...interface{}) (interface{}, error) {
	return ReadContract(s.Client, contractAddress, contractABI, startBlock, functionName, functionParams...)
}

// GetBlockByHash retrieves a block by its hash.
func (s *IndexerService) GetBlockByHash(blockHash common.Hash) (*types.Block, error) {
	return s.Client.BlockByHash(context.Background(), blockHash)
}

// GetTransactionByHash retrieves transaction details based on the transaction hash.
func (s *IndexerService) GetTransactionByHash(txHash common.Hash) (txInfo TransactionInfo, err error) {
	tx, _, err := s.Client.TransactionByHash(context.Background(), txHash)
	if err != nil {
		return
	}

	chainID, err := s.Client.NetworkID(context.Background())
	if err != nil {
		return
	}

	var signer types.Signer
	switch tx.Type() {
	case types.LegacyTxType:
		signer = types.NewEIP155Signer(chainID)
	case types.AccessListTxType, types.DynamicFeeTxType:
		signer = types.NewLondonSigner(chainID)
	default:
		err = fmt.Errorf("unsupported transaction type: %d", tx.Type())
		return
	}

	from, err := types.Sender(signer, tx)
	if err != nil {
		return
	}
	var to common.Address
	if tx.To() != nil {
		to = *tx.To()
	}
	value := tx.Value()

	txInfo = TransactionInfo{
		TxHash:    txHash,
		From:      from,
		To:        to,
		Value:     value,
		Timestamp: tx.Time().Unix(),
	}
	return
}

// EventHandler is a function type for handling events.
type EventHandler func(idx *IndexerService, event Event)

// TransactionInfo represents transaction information.
type TransactionInfo struct {
	TxHash    common.Hash
	From      common.Address
	To        common.Address
	Value     *big.Int
	Timestamp int64
}

// BlockInfo represents block information.
type BlockInfo struct {
	BlockNumber int64
	Timestamp   int64
}
