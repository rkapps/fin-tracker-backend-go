package kraken

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string // "https://api.coinbase.com"
	version    string
	limit      int
}

// compile-time check this satisfies the interface — also self-documenting
var _ API = (*Client)(nil)

func NewHTTPClient() *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL:    "https://api.kraken.com",
		version:    "0",
	}
}

func (c *Client) GetAssets(ctx context.Context) (map[string]AssetInfo, error) {
	assets := map[string]AssetInfo{}

	_, err := doRequestPublic(c.baseURL, c.version, "Assets", &assets)
	// log.Println(body)
	if err != nil {
		return nil, err
	}
	return assets, nil
}

func (c *Client) GetAssetPairs(ctx context.Context) (map[string]AssetPairInfo, error) {
	pairs := map[string]AssetPairInfo{}

	_, err := doRequestPublic(c.baseURL, c.version, "AssetPairs", &pairs)
	// log.Println(body)
	if err != nil {
		return nil, err
	}
	return pairs, nil
}

func (c *Client) GetTradeHistory(ctx context.Context, cfg Config, start, offset string) (*TradesHistoryResponse, error) {

	values := url.Values{}
	if len(start) > 0 {
		values.Add("start", start)
	} else {
		if len(offset) > 0 {
			values.Add("ofs", offset)
		}
	}

	thr := TradesHistoryResponse{}
	if _, err := doRequestPrivate(cfg, c.baseURL, c.version, "TradesHistory", values, &thr); err != nil {
		return nil, err
	}

	return &thr, nil
}

func (c *Client) GetLedgers(ctx context.Context, cfg Config, start string, offset string) (*LedgersResponse, error) {

	values := url.Values{}
	values.Add("type", "all")
	if len(start) > 0 {
		values.Add("start", start)
	} else {
		if len(offset) > 0 {
			values.Add("ofs", offset)
		}
	}

	var ldgs LedgersResponse
	if _, err := doRequestPrivate(cfg, c.baseURL, c.version, "Ledgers", values, &ldgs); err != nil {
		return nil, err
	}
	return &ldgs, nil

}
