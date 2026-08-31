package ethereum

import (
	"fmt"
	"strings"
	"time"

	"github.com/nanmu42/etherscan-api"
)

const (
	SLEEP_TIME               = 500 * time.Millisecond
	ETHEREUM_MAINNET_CHAINID = 1
	OPTIMISM_MAINNET_CHAINID = 10
	POLYGON_MAINNET_CHAINID  = 137
	MAX_BLOCKNUMBER          = 999_999_999
)

type Client struct {
	client *etherscan.Client
}

// compile-time check this satisfies the interface — also self-documenting
var _ API = (*Client)(nil)

func NewEtherscanClient(apiKey string) *Client {

	client := etherscan.NewCustomized(etherscan.Customization{
		Timeout: SLEEP_TIME,
		Key:     apiKey,
		BaseURL: fmt.Sprintf("https://api.etherscan.io/v2/api?chainid=%d&", ETHEREUM_MAINNET_CHAINID),
		Verbose: false,
	})
	return &Client{client: client}
}

func NewOptimismClient(apiKey string) *Client {

	client := etherscan.NewCustomized(etherscan.Customization{
		Timeout: SLEEP_TIME,
		Key:     apiKey,
		BaseURL: fmt.Sprintf("https://api.etherscan.io/v2/api?chainid=%d&", OPTIMISM_MAINNET_CHAINID),
		Verbose: false,
	})
	return &Client{client: client}
}

func NewPolygonClient(apiKey string) *Client {

	client := etherscan.NewCustomized(etherscan.Customization{
		Timeout: SLEEP_TIME,
		Key:     apiKey,
		BaseURL: fmt.Sprintf("https://api.etherscan.io/v2/api?chainid=%d&", POLYGON_MAINNET_CHAINID),
		Verbose: false,
	})
	return &Client{client: client}
}

func (c *Client) GetNormalTransactions(addr string, blockNumber *int, page int, offset int) ([]etherscan.NormalTx, error) {

	endBlock := MAX_BLOCKNUMBER
	txns, err := c.client.NormalTxByAddress(addr, blockNumber, &endBlock, page, offset, false)
	if err != nil {
		return nil, err
	}

	return txns, nil
}

func (c *Client) GetInternalTransactions(addr string, blockNumber *int, page int, offset int) ([]etherscan.InternalTx, error) {

	endBlock := MAX_BLOCKNUMBER
	txns, err := c.client.InternalTxByAddress(addr, blockNumber, &endBlock, page, offset, false)
	if err != nil {
		return nil, err
	}

	return txns, nil
}
func (c *Client) GetERC20Transfers(addr string, blockNumber *int, page int, offset int) ([]etherscan.ERC20Transfer, error) {

	endBlock := MAX_BLOCKNUMBER
	txns, err := c.client.ERC20Transfers(nil, &addr, blockNumber, &endBlock, page, offset, false)
	ntxns := []etherscan.ERC20Transfer{}
	for _, txn := range txns {

		from := strings.ToLower(txn.From)
		to := strings.ToLower(txn.To)
		laddr := strings.ToLower(addr)
		if strings.Compare(from, laddr) != 0 &&
			strings.Compare(to, laddr) != 0 {
			continue
		}
		ntxns = append(ntxns, txn)
	}
	if err != nil {
		return nil, err
	}

	return ntxns, nil
}

func (c *Client) GetERC721Transfers(addr string, blockNumber *int, page int, offset int) ([]etherscan.ERC721Transfer, error) {
	endBlock := MAX_BLOCKNUMBER

	txns, err := c.client.ERC721Transfers(nil, &addr, blockNumber, &endBlock, page, offset, false)
	if err != nil {
		return nil, err
	}
	return txns, nil
}
