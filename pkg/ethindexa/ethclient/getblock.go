package ethclient

import (
	"context"
	"fmt"
	"math/big"
	"reflect"
	"time"

	"hw/pkg/logger"
	"hw/pkg/request"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

// This package is used to fix "transaction type not supported"

// GetBlockResponse represents the response structure for retrieving a block.
type GetBlockResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      int    `json:"id"`
	Result  struct {
		Number          string                   `json:"number"`
		Hash            string                   `json:"hash"`
		Transactions    []GetTransactionResponse `json:"transactions"`
		LogsBloom       string                   `json:"logsBloom"`
		TotalDifficulty string                   `json:"totalDifficulty"`
		ReceiptsRoot    string                   `json:"receiptsRoot"`
		ExtraData       string                   `json:"extraData"`
		WithdrawalsRoot string                   `json:"withdrawalsRoot"`
		BaseFeePerGas   string                   `json:"baseFeePerGas"`
		Nonce           string                   `json:"nonce"`
		Miner           string                   `json:"miner"`
		Withdrawals     []struct {
			Amount         string `json:"amount"`
			Address        string `json:"address"`
			Index          string `json:"index"`
			ValidatorIndex string `json:"validatorIndex"`
		} `json:"withdrawals"`
		ExcessBlobGas         string        `json:"excessBlobGas"`
		Difficulty            string        `json:"difficulty"`
		GasLimit              string        `json:"gasLimit"`
		GasUsed               string        `json:"gasUsed"`
		Uncles                []interface{} `json:"uncles"`
		ParentBeaconBlockRoot string        `json:"parentBeaconBlockRoot"`
		Size                  string        `json:"size"`
		Sha3Uncles            string        `json:"sha3Uncles"`
		TransactionsRoot      string        `json:"transactionsRoot"`
		StateRoot             string        `json:"stateRoot"`
		MixHash               string        `json:"mixHash"`
		ParentHash            string        `json:"parentHash"`
		BlobGasUsed           string        `json:"blobGasUsed"`
		Timestamp             string        `json:"timestamp"`
	} `json:"result"`
}

// Number returns the block number as a big.Int.
func (block GetBlockResponse) Number() *big.Int {
	return new(big.Int).SetBytes(common.FromHex(block.Result.Number))
}

// Time returns the block timestamp as an int64.
func (block GetBlockResponse) Time() int64 {
	return new(big.Int).SetBytes(common.FromHex(block.Result.Timestamp)).Int64()
}

// GetBlockByNumber retrieves a block by its number.
func (c *Client) GetBlockByNumber(ctx context.Context, number *big.Int) (*GetBlockResponse, error) {
	var res GetBlockResponse
	var errResp ErrorResponse

	// Format the request parameters by converting the block number to a hexadecimal string.
	reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByNumber","params":["0x%s", true],"id":1}`, number.Text(16))

	response, reqErr := request.NewClient(
		request.Timeout("5s"),
		request.SetRetryCount(0),
		request.Header(map[string]string{
			"Content-Type": "application/json",
		}),
	).
		SetResult(&res).
		SetError(&errResp).
		SetBody(reqBody).
		Do("POST", c.RPCURL)

	// Check for API response errors.
	if !reflect.DeepEqual(errResp, ErrorResponse{}) {
		return nil, fmt.Errorf("API error code %d: %s", errResp.Error.Code, errResp.Error.Message)
	}

	// Check for request execution errors.
	if reqErr != nil {
		logger.Warnf("request response data: %+v", response)
		return nil, fmt.Errorf("request failed: %w", reqErr)
	}

	return &res, nil
}

// GetBlockByHash retrieves a block by its hash.
func (c *Client) GetBlockByHash(ctx context.Context, hash string) (*GetBlockResponse, error) {
	var res GetBlockResponse
	var errResp ErrorResponse

	// Look for the user in the cache, if not found, query from the database
	err := c.localCache.GetFunc(ctx, c.localCache.FormatKey(c.Name, "eth_getBlockByHash", hash), &res, time.Second*5, func(ctx context.Context) (interface{}, error) {
		var res GetBlockResponse
		reqBody := fmt.Sprintf(`{"jsonrpc":"2.0","method":"eth_getBlockByHash","params":["%s", true],"id":1}`, hash)

		_, reqErr := request.NewClient(
			request.Timeout("12s"),
			request.SetRetryCount(2),
			request.Header(map[string]string{
				"Content-Type": "application/json",
			}),
		).
			SetResult(&res).
			SetError(&errResp).
			SetBody(reqBody).
			Do("POST", c.RPCURL)

		// Check for API response errors.
		if !reflect.DeepEqual(errResp, ErrorResponse{}) {
			return nil, fmt.Errorf("API error code %d: %s", errResp.Error.Code, errResp.Error.Message)
		}

		// Check for request execution errors.
		if reqErr != nil {
			return nil, fmt.Errorf("request failed: %w", reqErr)
		}

		return res, nil
	})

	return &res, err
}

// HeaderByNumber retrieves a block header by its number.
func (c *Client) HeaderByNumber(ctx context.Context, number *big.Int) (*types.Header, error) {
	return c.Client.HeaderByNumber(ctx, number)
}
