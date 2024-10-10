package ethclient

import (
	"os"

	"hw/pkg/cache"

	"github.com/ethereum/go-ethereum/ethclient"
)

// Client represents an Ethereum client with caching capabilities.
type Client struct {
	Name       string
	RPCURL     string
	Client     *ethclient.Client
	localCache cache.Cache
}

// ErrorResponse represents the structure of an error response from the Ethereum JSON-RPC.
type ErrorResponse struct {
	Jsonrpc string      `json:"jsonrpc"`
	ID      interface{} `json:"id"`
	Error   struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// NewClient creates a new Ethereum client with the given network and RPC URL.
func NewClient(network, rpcURL string) (*Client, error) {
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return nil, err
	}

	os.Setenv("CACHE_DEFAULT_TTL", "10s")
	cache := cache.NewLocalCache()

	return &Client{
		Name:       network,
		Client:     client,
		RPCURL:     rpcURL,
		localCache: cache,
	}, nil
}
