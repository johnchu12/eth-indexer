package ethindexa

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"hw/internal/service"
	"hw/pkg/logger"
	"hw/pkg/pg"

	"hw/pkg/ethindexa/ethclient"
	"hw/pkg/ethindexa/utils"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"golang.org/x/sync/errgroup"
)

// Config defines the structure of the configuration file.
type Config struct {
	Networks  map[string]NetworkConfig  `json:"networks"`
	Contracts map[string]ContractConfig `json:"contracts"`
}

// NetworkConfig defines the configuration for a network.
type NetworkConfig struct {
	ChainID            int    `json:"chainId"`
	RPCURL             string `json:"rpc_url"`
	Address            string `json:"address"`
	StartBlock         int64  `json:"startBlock"`
	FinalityBlockCount int64  `json:"finalityBlockCount"`
}

// ContractConfig defines the configuration for each contract.
type ContractConfig struct {
	ABI      string                           `json:"abi"`
	Networks map[string]ContractNetworkConfig `json:"network"`
	Events   []string                         `json:"events"`
}

// ContractNetworkConfig defines the contract configuration on a specific network.
type ContractNetworkConfig struct {
	Address    string `json:"address"`
	StartBlock int64  `json:"startBlock"`
}

// EventConfig defines the structure of event configuration.
type EventConfig struct {
	ContractName       string
	ContractAddress    common.Address
	ContractABI        abi.ABI
	StartBlock         *big.Int
	FinalityBlockCount *big.Int
	EventName          string
	Handler            EventHandler
}

// BlockTask defines the structure for block data.
type BlockTask struct {
	Network         string
	BlockNumber     uint64
	BlockData       *ethclient.GetBlockResponse
	TransactionLogs []types.Log
}

// EventsTask defines the structure for events tasks.
type EventsTask struct {
	Network string
	Blocks  map[string]*ethclient.GetBlockResponse
	Logs    []types.Log
	mutex   sync.RWMutex
}

// HandlerTask defines the structure for handling tasks.
type HandlerTask struct {
	Network        string
	BlockNumber    int64
	EventHandler   EventHandler
	IndexerService *IndexerService
	Event          Event
}

// IndexerImpl implements the Indexer interface.
type IndexerImpl struct {
	Clients       map[string]*ethclient.Client
	Events        map[string]map[common.Hash][]*EventConfig // map[network][topic0][]*EventConfig
	Service       service.Service
	MainCtx       context.Context
	CancelFunc    context.CancelFunc
	Wg            sync.WaitGroup
	HandlerQueues map[string]chan HandlerTask
	EventQueues   map[string]chan *EventsTask
}

var (
	MaxBatchEventSize   = 10
	MaxBatchHandlerSize = 200
)

// NewIndexer creates a new instance of IndexerImpl and injects necessary dependencies.
func NewIndexer(db *pg.PostgresDB, service service.Service, handlers map[string]EventHandler) (*IndexerImpl, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current working directory: %w", err)
	}

	configPath := filepath.Join(workingDir, "internal", "indexer", "config.json")
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(configFile, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Initialize main context and cancel function.
	mainContext, cancel := context.WithCancel(context.Background())

	indexer := &IndexerImpl{
		Clients:       make(map[string]*ethclient.Client),
		Events:        make(map[string]map[common.Hash][]*EventConfig),
		Service:       service,
		MainCtx:       mainContext,
		CancelFunc:    cancel,
		HandlerQueues: make(map[string]chan HandlerTask),
		EventQueues:   make(map[string]chan *EventsTask),
	}

	// Initialize configuration as map[network][topic0][]*EventConfig
	for contractName, contractConfig := range config.Contracts {
		for networkName, networkConfig := range contractConfig.Networks {
			// Check network configuration
			netConfig, exists := config.Networks[networkName]
			if !exists {
				return nil, fmt.Errorf("network configuration not found: %s", networkName)
			}

			// If the client for the network is not yet created, create and store it.
			if _, exists := indexer.Clients[networkName]; !exists {
				client, err := ethclient.NewClient(networkName, netConfig.RPCURL)
				if err != nil {
					return nil, fmt.Errorf("failed to connect to network %s: %w", networkName, err)
				}
				indexer.Clients[networkName] = client
			}

			contractAddress := common.HexToAddress(networkConfig.Address)
			startBlockNumber := networkConfig.StartBlock

			if _, exists := indexer.Events[networkName]; !exists {
				indexer.Events[networkName] = make(map[common.Hash][]*EventConfig)
			}

			for _, eventName := range contractConfig.Events {
				handlerKey := fmt.Sprintf("%s:%s:%s", contractName, networkName, eventName)
				eventHandler, hasHandler := handlers[handlerKey]
				if !hasHandler {
					eventHandler = nil
				}

				parsedABI, err := utils.LoadABI(contractConfig.ABI)
				if err != nil {
					return nil, fmt.Errorf("failed to load ABI for contract %s: %w", contractName, err)
				}

				topic0, err := GetEventTopic0(parsedABI, eventName)
				if err != nil {
					return nil, fmt.Errorf("failed to get Topic0 for event %s: %w", eventName, err)
				}

				eventConfig := &EventConfig{
					ContractName:       contractName,
					ContractAddress:    contractAddress,
					ContractABI:        parsedABI,
					StartBlock:         big.NewInt(startBlockNumber),
					FinalityBlockCount: big.NewInt(netConfig.FinalityBlockCount),
					EventName:          eventName,
					Handler:            eventHandler,
				}

				indexer.Events[networkName][topic0] = append(indexer.Events[networkName][topic0], eventConfig)
			}
		}
	}

	// Initialize handlerQueue and eventQueue for each network
	for networkName := range indexer.Events {
		indexer.HandlerQueues[networkName] = make(chan HandlerTask, MaxBatchHandlerSize)
		indexer.EventQueues[networkName] = make(chan *EventsTask, MaxBatchEventSize)
	}

	// Start event consumers for each network
	for networkName, eventConfigs := range indexer.Events {
		client, exists := indexer.Clients[networkName]
		if !exists {
			log.Printf("No client found for network %s", networkName)
			continue
		}
		indexer.Wg.Add(3)
		logger.Infof("Starting event consumers for network %s with configurations %+v", networkName, eventConfigs)
		go indexer.startBlockFetcher(networkName, client, eventConfigs)
		go indexer.startLogProcessor(networkName)
		go indexer.startTaskHandler(networkName)
	}

	return indexer, nil
}

// GetEventTopic0 calculates the topic[0] signature for the specified event.
func GetEventTopic0(contractABI abi.ABI, eventName string) (common.Hash, error) {
	event, exists := contractABI.Events[eventName]
	if !exists {
		return common.Hash{}, fmt.Errorf("event %s not found in ABI", eventName)
	}

	return event.ID, nil
}

// loadABI loads and parses the ABI.
func loadABI(abiName string) (abi.ABI, error) {
	// Assume ABI files are located in the internal/indexer/abis/ directory with the filename {name}.json
	workingDir, err := os.Getwd()
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to get current working directory: %w", err)
	}

	abiFilePath := filepath.Join(workingDir, "internal", "indexer", "abis", fmt.Sprintf("%s.json", abiName))
	abiBytes, err := os.ReadFile(abiFilePath)
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to read ABI file %s: %w", abiFilePath, err)
	}

	parsedABI, err := abi.JSON(strings.NewReader(string(abiBytes)))
	if err != nil {
		return abi.ABI{}, fmt.Errorf("failed to parse ABI for %s: %w", abiName, err)
	}

	return parsedABI, nil
}

// startBlockFetcher starts the block fetching consumer.
func (indexer *IndexerImpl) startBlockFetcher(networkName string, client *ethclient.Client, eventConfigs map[common.Hash][]*EventConfig) {
	defer indexer.Wg.Done()

	// Get the minimum start block from the configuration
	minStartBlock := big.NewInt(0)
	finalityBlockCount := big.NewInt(0)
	for _, eventConfigList := range eventConfigs {
		for _, config := range eventConfigList {
			if minStartBlock.Cmp(config.StartBlock) < 0 {
				minStartBlock.Set(config.StartBlock)
			}
			if finalityBlockCount.Cmp(config.FinalityBlockCount) < 0 {
				finalityBlockCount.Set(config.FinalityBlockCount)
			}
		}
	}

	// Main block fetching loop
	for {
		select {
		case <-indexer.MainCtx.Done():
			return
		default:
			latestBlockHeader, err := client.HeaderByNumber(context.Background(), nil)
			if err != nil {
				log.Printf("Failed to get latest block for network %s: %v", networkName, err)
				continue
			}
			latestBlockNumber := big.NewInt(0).Sub(latestBlockHeader.Number, finalityBlockCount).Uint64()

			startBlock := minStartBlock.Uint64()
			endBlock := latestBlockNumber

			currentBlock := startBlock

			// Process 37 blocks at a time
			for currentBlock <= endBlock {
				eg, ctx := errgroup.WithContext(indexer.MainCtx)

				startTime := time.Now()

				processingEndBlock := currentBlock + 37
				if processingEndBlock >= endBlock {
					processingEndBlock = endBlock
				}

				logEntries, err := client.GetLogsByBlockNumber(context.Background(), big.NewInt(int64(currentBlock)), big.NewInt(int64(processingEndBlock)), getUniqueAddresses(eventConfigs))
				if err != nil {
					log.Printf("Failed to get logs for network %s from #%d to #%d: %v", networkName, currentBlock, processingEndBlock, err)
					break
				}

				eventsTask := EventsTask{
					Network: networkName,
					Blocks:  make(map[string]*ethclient.GetBlockResponse),
					Logs:    logEntries,
				}

				for _, logEntry := range logEntries {
					blockNumberKey := fmt.Sprintf("%d", logEntry.BlockNumber)
					_, exists := eventsTask.Blocks[blockNumberKey]
					if exists {
						continue
					}

					eventsTask.mutex.Lock()
					eventsTask.Blocks[blockNumberKey] = nil
					eventsTask.mutex.Unlock()

					eg.Go(func() error {
						ctxLog, cancel := context.WithCancel(ctx)
						defer cancel()

						blockResponse, err := client.GetBlockByHash(ctxLog, logEntry.BlockHash.Hex())
						if err != nil {
							log.Printf("Failed to get block by hash %s: %v", logEntry.BlockHash.Hex(), err)
							return fmt.Errorf("failed to get block by hash %s: %w", logEntry.BlockHash.Hex(), err)
						}
						eventsTask.mutex.Lock()
						eventsTask.Blocks[blockNumberKey] = blockResponse
						eventsTask.mutex.Unlock()
						return nil
					})
				}

				// Wait for all goroutines to finish
				if err := eg.Wait(); err != nil {
					logger.Errorf("Error fetching blocks for network %s: %v", networkName, err)
					break
				}

				logger.Infof("Fetched %s blocks %d to %d (%s)", networkName, currentBlock, processingEndBlock, time.Since(startTime))

				indexer.EventQueues[networkName] <- &eventsTask
				currentBlock = processingEndBlock + 1
			}

			logger.Infof("Processed blocks %d to %d... waiting for new blocks", startBlock, endBlock)

			// Update the minimum start block to the last processed block
			minStartBlock.SetUint64(endBlock + 1)

			// Wait before checking for new blocks again
			time.Sleep(20 * time.Second)
		}
	}
}

// startLogProcessor starts the log processing consumer.
func (indexer *IndexerImpl) startLogProcessor(networkName string) {
	defer indexer.Wg.Done()
	for {
		select {
		case <-indexer.MainCtx.Done():
			// Close the eventQueue when the main context is canceled
			close(indexer.EventQueues[networkName])
			return
		default:
			for {
				select {
				case <-indexer.MainCtx.Done():
					return
				case eventTask, ok := <-indexer.EventQueues[networkName]:
					if !ok {
						logger.Errorf("EventQueue for network %s is closed", networkName)
						return
					}
					// Parse and filter events
					for _, logEntry := range eventTask.Logs {
						if len(logEntry.Topics) == 0 {
							logger.Warnf("No topics found")
							continue
						}
						topic0 := logEntry.Topics[0]
						eventConfigs, exists := indexer.Events[networkName][topic0]
						if !exists {
							continue
						}

						for _, eventConfig := range eventConfigs {
							// Skip if handler is not set
							if eventConfig.Handler == nil {
								continue
							}

							// Compare contract address
							if logEntry.Address != eventConfig.ContractAddress {
								continue
							}

							// Compare block number
							if logEntry.Address == eventConfig.ContractAddress && logEntry.BlockNumber < eventConfig.StartBlock.Uint64() {
								continue
							}

							// Decode event
							eventArgs, err := eventConfig.extractEventArgs(logEntry)
							if err != nil {
								logger.Warnf("Failed to extract event args for log %s: %v", logEntry.TxHash.Hex(), err)
								continue
							}

							blockResponse, exists := eventTask.Blocks[fmt.Sprintf("%d", logEntry.BlockNumber)]
							if !exists {
								logger.Errorf("Block %d not found", logEntry.BlockNumber)
								continue
							}

							getTransaction := func(transactions []ethclient.GetTransactionResponse, txHash string) ethclient.GetTransactionResponse {
								for _, tx := range transactions {
									if tx.Hash == txHash {
										return tx
									}
								}
								logger.Warnf("Transaction %s not found in block %d for network %s", logEntry.TxHash.Hex(), logEntry.BlockNumber, networkName)
								return ethclient.GetTransactionResponse{}
							}

							// Create event context
							eventContext, cancel := context.WithCancel(indexer.MainCtx)
							event := Event{
								Block:           *blockResponse,
								Transaction:     getTransaction(blockResponse.Result.Transactions, logEntry.TxHash.Hex()),
								NetworkName:     eventTask.Network,
								ContractName:    eventConfig.ContractName,
								EventName:       eventConfig.EventName,
								ContractAddress: eventConfig.ContractAddress,
								Args:            eventArgs,
								TransactionHash: logEntry.TxHash,
								BlockHash:       logEntry.BlockHash,
								Ctx:             eventContext,
								Cancel:          cancel,
							}

							indexerService := &IndexerService{
								Client:  indexer.Clients[eventTask.Network].Client,
								Service: indexer.Service,
							}

							// Add handling task to handlerQueue
							indexer.HandlerQueues[networkName] <- HandlerTask{
								Network:        eventTask.Network,
								BlockNumber:    int64(logEntry.BlockNumber),
								EventHandler:   eventConfig.Handler,
								IndexerService: indexerService,
								Event:          event,
							}
						}
					}
				}
			}
		}
	}
}

// startTaskHandler starts the task handling consumer.
func (indexer *IndexerImpl) startTaskHandler(networkName string) {
	defer indexer.Wg.Done()
	for {
		select {
		case <-indexer.MainCtx.Done():
			return
		case task, ok := <-indexer.HandlerQueues[networkName]:
			if !ok {
				return
			}
			task.EventHandler(task.IndexerService, task.Event)
		}
	}
}

// getUniqueAddresses extracts unique contract addresses from event configurations.
func getUniqueAddresses(eventConfigs map[common.Hash][]*EventConfig) []common.Address {
	addressMap := make(map[common.Address]struct{})
	for _, configList := range eventConfigs {
		for _, config := range configList {
			addressMap[config.ContractAddress] = struct{}{}
		}
	}
	addresses := make([]common.Address, 0, len(addressMap))
	for addr := range addressMap {
		addresses = append(addresses, addr)
	}

	return addresses
}

// extractEventArgs extracts event arguments from the log entry.
func (eventConfig *EventConfig) extractEventArgs(logEntry types.Log) (map[string]interface{}, error) {
	eventArgs := make(map[string]interface{})

	err := eventConfig.ContractABI.UnpackIntoMap(eventArgs, eventConfig.EventName, logEntry.Data)
	if err != nil {
		return nil, fmt.Errorf("error unpacking event data: %w", err)
	}

	event, exists := eventConfig.ContractABI.Events[eventConfig.EventName]
	if !exists {
		return nil, fmt.Errorf("event %s not found in ABI", eventConfig.EventName)
	}

	indexedParamIndex := 1
	for _, input := range event.Inputs {
		if input.Indexed {
			if indexedParamIndex >= len(logEntry.Topics) {
				return nil, fmt.Errorf("not enough topics to unpack indexed parameters: %d", indexedParamIndex)
			}
			switch input.Type.String() {
			case "address":
				eventArgs[input.Name] = common.HexToAddress(logEntry.Topics[indexedParamIndex].Hex())
			case "uint256":
				eventArgs[input.Name] = new(big.Int).SetBytes(logEntry.Topics[indexedParamIndex].Bytes())
			default:
				eventArgs[input.Name] = logEntry.Topics[indexedParamIndex].Hex()
			}
			indexedParamIndex++
		}
	}

	return eventArgs, nil
}

// Stop stops all event consumers and cancels the main context.
func (indexer *IndexerImpl) Stop() {
	indexer.CancelFunc()
	indexer.Wg.Wait()
	logger.Infow("All event consumers have been stopped.")
}
