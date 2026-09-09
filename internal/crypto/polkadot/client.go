package polkadot

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/internal/core"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	apiKey     string
}

// compile-time check this satisfies the interface — also self-documenting
var _ API = (*Client)(nil)

func NewPolkadotHttpClient(apiKey string) *Client {
	return &Client{
		httpClient: &http.Client{Timeout: 45 * time.Second},
		baseURL:    "https://api.pubfi.ai/v1/gateway/subscan/polkadot/api/v2/scan",
		apiKey:     apiKey,
	}
}

// GetEpoch implements [cardano.API].
func (c *Client) GetRewards(address string, page int, row int) (*PolkadotRewardData, error) {

	dotinput := PolkadotRewardInput{}
	dotinput.Address = address
	dotinput.Category = "Reward"
	dotinput.Is_Stash = true
	dotinput.Page = page
	dotinput.Row = row

	body, err := json.Marshal(dotinput)
	if err != nil {
		return nil, err
	}

	url := fmt.Sprintf("%s/account/reward_slash:free", c.baseURL)
	dotRewardData := &PolkadotRewardData{}
	core.DoHttpRequest(url, http.MethodPost, c.getPolkadotHeaders(), nil, body, &dotRewardData)
	return dotRewardData, err
}

func (c *Client) GetTransfers(address string, page int, row int) (*PolkadotTransferData, error) {

	tsfrinput := PolkadotTransferInput{}
	tsfrinput.Address = address
	tsfrinput.After_Id = []int{0}
	tsfrinput.Order = "asc"
	tsfrinput.Page = page
	tsfrinput.Row = row

	body, err := json.Marshal(tsfrinput)
	if err != nil {
		return nil, err
	}

	dotData := PolkadotTransferData{}
	url := fmt.Sprintf("%s/transfers:free", c.baseURL)
	_, err = core.DoHttpRequest(url, http.MethodPost, c.getPolkadotHeaders(), nil, body, &dotData)
	// log.Println(body1)
	return &dotData, err
}

func (c *Client) getPolkadotHeaders() url.Values {
	headers := url.Values{}
	headers.Add("Content-Type", "application/json")
	headers.Add("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	return headers
}
