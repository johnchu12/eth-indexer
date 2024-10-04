package ethindexa

import (
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"hw/internal/service"
	"hw/pkg/ethindexa/utils"
	"hw/pkg/logger"
	"hw/pkg/pg"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Config defines the structure of the configuration file.
type Config struct {
	Networks  map[string]NetworkConfig  `json:"networks"`
	Contracts map[string]ContractConfig `json:"contracts"`
}

// NetworkConfig defines the configuration for a network.
type NetworkConfig struct {
	ChainID int    `json:"chainId"`
	RPCURL  string `json:"rpc_url"`
}

// ContractConfig defines the configuration for each contract.
type ContractConfig struct {
	ABI      string                           `json:"abi"`
	Networks map[string]ContractNetworkConfig `json:"network"`
	Events   []string                         `json:"events"`
}

// ContractNetworkConfig defines the configuration of a contract on a specific network.
type ContractNetworkConfig struct {
	Address    string `json:"address"`
	StartBlock int64  `json:"startBlock"`
}

// EventConfig defines the structure of event configuration.
type EventConfig struct {
	ContractName    string // Name of the contract
	ContractAddress common.Address
	ContractABI     abi.ABI
	StartBlock      *big.Int
	EventName       string
}

// IndexerImpl implements the Indexer interface.
type IndexerImpl struct {
	clients    map[string]*ethclient.Client
	contracts  map[string]*EventConfig
	handlers   map[string]EventHandler
	mainCtx    context.Context
	cancelFunc context.CancelFunc
	networkWG  sync.WaitGroup
	service    service.Service
}

// NewIndexer creates a new IndexerImpl instance and injects the necessary dependencies.
func NewIndexer(db *pg.PostgresDB, service service.Service, handlers map[string]EventHandler) (*IndexerImpl, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	configPath := filepath.Join(wd, "internal", "indexer", "config.json")
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Initialize the main context and cancel function.
	mainCtx, cancel := context.WithCancel(context.Background())

	indexer := &IndexerImpl{
		clients:    make(map[string]*ethclient.Client),
		contracts:  make(map[string]*EventConfig),
		handlers:   handlers,
		mainCtx:    mainCtx,
		cancelFunc: cancel,
		service:    service,
	}

	// Initialize contract event configurations, including only events with registered handlers.
	for contractName, contract := range config.Contracts {
		parsedABI, err := utils.LoadABI(contract.ABI)
		if err != nil {
			return nil, fmt.Errorf("failed to load and parse ABI: %w", err)
		}

		for networkName, networkConfig := range contract.Networks {
			netConfig, ok := config.Networks[networkName]
			if !ok {
				return nil, fmt.Errorf("network configuration not found: %s", networkName)
			}

			// If the client for the network is not created yet, create and store it.
			if _, exists := indexer.clients[networkName]; !exists {
				client, err := ethclient.Dial(netConfig.RPCURL)
				if err != nil {
					return nil, fmt.Errorf("failed to connect to network %s: %w", networkName, err)
				}
				indexer.clients[networkName] = client
			}

			address := common.HexToAddress(networkConfig.Address)
			startBlock := networkConfig.StartBlock

			for _, eventName := range contract.Events {
				key := fmt.Sprintf("%s:%s:%s", contractName, networkName, eventName)
				if _, handlerExists := handlers[key]; !handlerExists {
					// Skip the event if no handler is registered.
					continue
				}

				eventConfig, err := InitializeEventConfig(contractName, parsedABI, address, startBlock, eventName)
				if err != nil {
					return nil, fmt.Errorf("failed to create event configuration: %w", err)
				}
				indexer.contracts[key] = eventConfig
			}
		}
	}

	return indexer, nil
}

// InitializeEventConfig initializes and returns an EventConfig.
func InitializeEventConfig(contractName string, parsedABI abi.ABI, contractAddr common.Address, startBlock int64, eventName string) (*EventConfig, error) {
	return &EventConfig{
		ContractName:    contractName,
		ContractAddress: contractAddr,
		ContractABI:     parsedABI,
		StartBlock:      big.NewInt(startBlock),
		EventName:       eventName,
	}, nil
}

// RegisterHandler registers an event handler for the specified contract and event.
//
//	contractKey: The contract key specified in the configuration file in the format {ContractName}:{NetworkName}:{EventName}.
func (i *IndexerImpl) RegisterHandler(contractKey string, handler EventHandler) {
	i.handlers[contractKey] = handler
	logger.Infof("Registered handler for key: %s", contractKey)
}

// ListContracts returns all event configurations for the contracts.
func (i *IndexerImpl) ListContracts() map[string]*EventConfig {
	return i.contracts
}

// StartAllEventListeners starts listening to all contract events.
func (i *IndexerImpl) StartAllEventListeners() {
	logger.Infow("StartAllEventListeners", len(i.contracts))

	// Start a listener goroutine for each contract.
	for key, eventConfig := range i.contracts {
		i.networkWG.Add(1)
		go i.runContractListener(key, eventConfig)
	}
	i.networkWG.Wait()
}

// runContractListener runs an event listener for a specific contract.
//
//	contractKey: The contract key in the format "{ContractName}:{NetworkName}:{EventName}".
func (i *IndexerImpl) runContractListener(contractKey string, eventConfig *EventConfig) {
	defer i.networkWG.Done()

	// Extract network name and contract name from the contractKey.
	parts := strings.Split(contractKey, ":")
	if len(parts) != 3 {
		logger.Errorf("Invalid contract key format: %s", contractKey)
		return
	}
	contractName := parts[0]
	networkName := parts[1]
	eventName := parts[2]

	// Combine in the format {Network}:{Contract}.
	networkContract := fmt.Sprintf("%s:%s", networkName, contractName)

	logger.Infof("%s - Starting listener at %d with address: %s for event: %s",
		networkContract, eventConfig.StartBlock.Int64(), eventConfig.ContractAddress.Hex(), eventName)

	// Create a contract-level context.
	contractCtx, contractCancel := context.WithCancel(i.mainCtx)
	defer contractCancel()

	// Get the client associated with the contract.
	client := i.getClientByKey(contractKey)
	if client == nil {
		logger.Errorf("%s - No client available", networkContract)
		return
	}

	// Define the filter query.
	query := ethereum.FilterQuery{
		Addresses: []common.Address{eventConfig.ContractAddress},
		Topics:    [][]common.Hash{{eventConfig.ContractABI.Events[eventName].ID}},
	}

	// Initialize the current block number.
	currentBlock := eventConfig.StartBlock

	for {
		select {
		case <-contractCtx.Done():
			logger.Warnf("%s - Listener is shutting down", networkContract)
			return
		default:
			// Get the latest block number.
			latestBlock, err := client.BlockNumber(contractCtx)
			if err != nil {
				logger.Errorf("%s - Error getting latest block number: %v", networkContract, err)
				time.Sleep(10 * time.Second)
				continue
			}

			// Define the number of blocks to process in each batch.
			batchSize := int64(50)
			endBlock := new(big.Int).Add(currentBlock, big.NewInt(batchSize))
			if endBlock.Uint64() > latestBlock {
				endBlock = big.NewInt(int64(latestBlock))
			}

			// Update the filter query parameters.
			query.FromBlock = currentBlock
			query.ToBlock = endBlock

			// print blocks progress
			// logger.Infof("%s - Filtering logs from block %s to %s",
			// networkContract, currentBlock.String(), endBlock.String())

			logs, err := client.FilterLogs(contractCtx, query)
			if err != nil {
				logger.Errorf("%s - Error filtering logs: %v", networkContract, err)
				time.Sleep(10 * time.Second)
				continue
			}

			for _, vLog := range logs {
				// Process each event.
				go func(ec *EventConfig, tl types.Log) {
					eventData, err := ec.extractEventArgs(tl)
					if err != nil {
						logger.Errorf("%s - Error processing event arguments for event %s: %v",
							networkContract, ec.EventName, err)
						return
					}

					idx := &IndexerService{
						Client:  client,
						Service: i.service,
					}

					eventStruct := Event{
						EventName:       ec.EventName,
						Args:            eventData,
						TransactionHash: tl.TxHash,
						BlockHash:       tl.BlockHash,
						ContractAddress: ec.ContractAddress,
						ContractName:    ec.ContractName,
						NetworkName:     networkName,
						Ctx:             contractCtx,
					}

					// Retrieve the handler using the "ContractName:NetworkName:EventName" format.
					handler, exists := i.handlers[contractKey]
					if !exists {
						logger.Warnf("%s - Handler not found for contract key: %s",
							networkContract, contractKey)
						return
					}

					handler(idx, eventStruct)
				}(eventConfig, vLog)
			}

			// Update the current block number.
			currentBlock = new(big.Int).Add(endBlock, big.NewInt(1))

			// If the latest block has been processed, wait before continuing.
			if endBlock.Uint64() >= latestBlock {
				time.Sleep(15 * time.Second)
				logger.Infof("%s - Block history has been updated", networkContract)
			}
		}
	}
}

// getClientByKey retrieves the corresponding ethclient based on the contract key.
func (i *IndexerImpl) getClientByKey(key string) *ethclient.Client {
	parts := strings.Split(key, ":")
	if len(parts) != 3 {
		logger.Warnf("invalid contract key format: %s", key)
		return nil
	}
	networkName := parts[1]
	client, exists := i.clients[networkName]
	if !exists {
		logger.Errorf("client not found for network: %s", networkName)
		return nil
	}
	return client
}

// Stop stops all listeners and cancels the main context.
func (i *IndexerImpl) Stop() {
	i.cancelFunc()
	i.networkWG.Wait()
	logger.Infow("All listeners have been stopped.")
}

// extractEventArgs extracts event arguments and returns eventData.
func (ec *EventConfig) extractEventArgs(vLog types.Log) (map[string]interface{}, error) {
	eventData := make(map[string]interface{})

	err := ec.ContractABI.UnpackIntoMap(eventData, ec.EventName, vLog.Data)
	if err != nil {
		return nil, fmt.Errorf("error unpacking event data: %w", err)
	}

	event, exists := ec.ContractABI.Events[ec.EventName]
	if !exists {
		return nil, fmt.Errorf("event %s not found in ABI", ec.EventName)
	}

	indexedIndex := 1
	for _, input := range event.Inputs {
		if input.Indexed {
			if indexedIndex >= len(vLog.Topics) {
				return nil, fmt.Errorf("not enough Topics to unpack indexed parameters: %d", indexedIndex)
			}
			switch input.Type.String() {
			case "address":
				eventData[input.Name] = common.HexToAddress(vLog.Topics[indexedIndex].Hex())
			case "uint256":
				eventData[input.Name] = new(big.Int).SetBytes(vLog.Topics[indexedIndex].Bytes())
			default:
				eventData[input.Name] = vLog.Topics[indexedIndex].Hex()
			}
			indexedIndex++
		}
	}

	return eventData, nil
}
