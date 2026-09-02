package cardano

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/internal/core"
)

type Client struct {
	httpClient            *http.Client
	baseURL               string
	blockfrost_project_id string
}

// compile-time check this satisfies the interface — also self-documenting
var _ API = (*Client)(nil)

func NewBlockFrostHTTPClient(blockfrost_project_id string) *Client {
	return &Client{
		httpClient:            &http.Client{Timeout: 10 * time.Second},
		baseURL:               "https://cardano-mainnet.blockfrost.io/api/v0",
		blockfrost_project_id: blockfrost_project_id,
	}
}

// GetEpoch implements [cardano.API].
func (c *Client) GetEpochInformation(ctx context.Context, epoch int64, page int64) ([]EpochInformation, error) {
	url := fmt.Sprintf("%s/epochs/%d/next?page=%d", c.baseURL, epoch, page)
	var epochs []EpochInformation
	_, err := core.DoHttpRequest(url, http.MethodGet, c.getCardanoHeaders(), nil, nil, &epochs)
	// log.Println(body)
	return epochs, err
}

func (c *Client) GetAccountRewards(ctx context.Context, saddress string, page int, count int) ([]AccountReward, error) {
	url := fmt.Sprintf("%s/accounts/%s/rewards?page=%d&count=%d", c.baseURL, saddress, page, count)
	var rewards []AccountReward
	_, err := core.DoHttpRequest(url, http.MethodGet, c.getCardanoHeaders(), nil, nil, &rewards)
	// log.Println(body)
	return rewards, err
}

func (c *Client) GetAccountAddresses(ctx context.Context, saddress string) ([]AccountAddress, error) {
	url := fmt.Sprintf("%s/accounts/%s/addresses", c.baseURL, saddress)
	var addresses []AccountAddress
	_, err := core.DoHttpRequest(url, http.MethodGet, c.getCardanoHeaders(), nil, nil, &addresses)
	// log.Println(body)
	return addresses, err
}

func (c *Client) GetAccountTransactions(ctx context.Context, address string, bheight int64, page int) ([]AddressTransaction, error) {
	url := fmt.Sprintf("%s/addresses/%s/transactions?from=%d&page=%d", c.baseURL, address, bheight, page)
	var txns []AddressTransaction
	_, err := core.DoHttpRequest(url, http.MethodGet, c.getCardanoHeaders(), nil, nil, &txns)
	// log.Println(body)
	return txns, err
}

func (c *Client) GetTransactionUTXOs(txHash string) (TransactionUTXO, error) {

	txutxo := TransactionUTXO{}
	url := fmt.Sprintf("%s/txs/%s/utxos", c.baseURL, txHash)
	_, err := core.DoHttpRequest(url, "GET", c.getCardanoHeaders(), nil, nil, &txutxo)
	return txutxo, err
}

func (c *Client) GetTransactionMetadata(txHash string) ([]TransactionMetadata, error) {
	tmds := []TransactionMetadata{}
	url := fmt.Sprintf("%s/txs/%s/metadata", c.baseURL, txHash)
	_, err := core.DoHttpRequest(url, "GET", c.getCardanoHeaders(), nil, nil, &tmds)
	return tmds, err
}

func (c *Client) GetTransactionDelegations(txHash string) ([]TransactionDelegation, error) {
	var delegations []TransactionDelegation
	url := fmt.Sprintf("%s/txs/%s/delegations", c.baseURL, txHash)
	_, err := core.DoHttpRequest(url, "GET", c.getCardanoHeaders(), nil, nil, &delegations)
	// log.Printf("body: %v", body)
	return delegations, err
}

func (c *Client) GetTransactionStakeCerticates(txHash string) ([]TransasctionStakeCertificate, error) {

	tstakes := []TransasctionStakeCertificate{}
	url := fmt.Sprintf("%s/txs/%s/stakes", c.baseURL, txHash)
	_, err := core.DoHttpRequest(url, "GET", c.getCardanoHeaders(), nil, nil, &tstakes)
	return tstakes, err

}

func (c *Client) GetTransactionInfo(txHash string) (TransactionInfo, error) {
	var tinfo TransactionInfo
	url := fmt.Sprintf("%s/txs/%s", c.baseURL, txHash)
	_, err := core.DoHttpRequest(url, "GET", c.getCardanoHeaders(), nil, nil, &tinfo)
	// log.Println(body)
	return tinfo, err
}

func (c *Client) GetTransactionWithdrawals(stakeaddress string) ([]TransactionWithdrawal, error) {
	var tws []TransactionWithdrawal
	url := fmt.Sprintf("%s/accounts/%s/withdrawals", c.baseURL, stakeaddress)
	_, err := core.DoHttpRequest(url, "GET", c.getCardanoHeaders(), nil, nil, &tws)
	return tws, err
}

func (c *Client) getCardanoHeaders() url.Values {
	headers := url.Values{}
	headers.Add("project_id", c.blockfrost_project_id)
	return headers
}
