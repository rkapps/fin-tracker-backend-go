package coinbase

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string // "https://api.coinbase.com"
	host       string
	limit      int
}

// compile-time check this satisfies the interface — also self-documenting
var _ API = (*Client)(nil)

func NewHTTPClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://api.coinbase.com",
		host:       "api.coinbase.com",
		limit:      10,
	}

}

// ListPaymentMethods implements [coinbase.API].
func (c *Client) ListPaymentMethods(ctx context.Context, cfg Config, pageToken string) (CnbV3PaymentData, error) {

	var data CnbV3PaymentData
	pnexturi := "/api/v3/brokerage/payment_methods"
	_, err := doNewRequest(cfg, "GET", c.host, c.baseURL, pnexturi, pnexturi, &data)
	// log.Println(body)
	return data, err
}

// ListFills implements [coinbase.API].
func (c *Client) ListFills(ctx context.Context, cfg Config, pageToken string) error {

	pnexturi := "/api/v3/brokerage/orders/historical/fills"
	_, err := doNewRequest(cfg, "GET", c.host, c.baseURL, pnexturi, pnexturi, nil)
	// log.Println(body)
	return err
}

// ListOrders implements [coinbase.API].
func (c *Client) ListOrders(ctx context.Context, cfg Config, pageToken string) error {

	pnexturi := "/api/v3/brokerage/orders/historical/batch"
	_, err := doNewRequest(cfg, "GET", c.host, c.baseURL, pnexturi, pnexturi, nil)
	// log.Println(body)
	return err
}

// ListAccounts implements [coinbase.API].
func (c *Client) ListAccounts(ctx context.Context, cfg Config, pageToken string, limit int) (CnbPaginatedData, error) {

	var data CnbPaginatedData
	requestPath := "/v2/accounts"
	pnexturi := fmt.Sprintf("/v2/accounts?limit=%d", limit)
	if len(pageToken) > 0 {
		pnexturi = pageToken
	}
	_, err := doNewRequest(cfg, "GET", c.host, c.baseURL, requestPath, pnexturi, &data)
	// log.Println(body)
	return data, err
}

// ListTransactions implements [coinbase.API].
func (c *Client) ListTransactions(ctx context.Context, cfg Config, accountId, endpoint, pageToken string) (CnbPaginatedData, error) {

	var data CnbPaginatedData
	requestPath := fmt.Sprintf("/v2/accounts/%s/%s", accountId, endpoint)
	pnexturi := fmt.Sprintf("%s?expand=all&limit=%d&order=asc", requestPath, c.limit)
	if len(pageToken) > 0 {
		pnexturi = pageToken
	}
	_, err := doNewRequest(cfg, "GET", c.host, c.baseURL, requestPath, pnexturi, &data)
	return data, err
}
