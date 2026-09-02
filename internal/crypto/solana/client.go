package solana

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/rkapps/fin-tracker-backend-go/internal/core"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
	API_KEY    string
}

// compile-time check this satisfies the interface — also self-documenting
var _ API = (*Client)(nil)

func NewAlchemyHTTPClient(apiKey string) *Client {
	return &Client{
		// httpClient: &http.Client{Timeout: 10 * time.Second},
		baseURL: "https://solana-mainnet.g.alchemy.com/v2",
		API_KEY: apiKey,
	}
}

// GetEpoch implements [cardano.API].
func (c *Client) GetSolanaTokenAccounts(address string) (SolanaTokenAccountResult, error) {

	programId := "TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA"

	si := c.createSolanaInput()
	si.Method = "getTokenAccountsByOwner"
	si.Params = append(si.Params, address)

	pId := ProgramId{ProgramId: programId}
	si.Params = append(si.Params, pId)

	ed := MyEncoding{Encoding: "jsonParsed"}
	si.Params = append(si.Params, ed)

	body, err := json.Marshal(si)
	if err != nil {
		log.Println(err)
	}

	url := fmt.Sprintf("%s/%s", c.baseURL, c.API_KEY)
	var result SolanaTokenAccountResult
	_, err = core.DoHttpRequest(url, http.MethodPost, nil, nil, body, &result)
	// log.Println(body)
	return result, err
}

func (c *Client) GetSolanaSignaturesForAddress(addr string, untilSig string) (SolanaSignatureResult, error) {

	si := c.createSolanaInput()
	si.Method = "getSignaturesForAddress"
	si.Params = append(si.Params, addr)

	if len(untilSig) == 0 {
		si.Params = append(si.Params, solanaInputConfig{Limit: 500})
	} else {
		si.Params = append(si.Params, solanaInputConfig{Limit: 500, Until: untilSig})
	}

	result := SolanaSignatureResult{}
	body, err := json.Marshal(si)
	if err != nil {
		return result, err
	}
	// log.Println(string(body))

	url := fmt.Sprintf("%s/%s", c.baseURL, c.API_KEY)
	_, err = core.DoHttpRequest(url, http.MethodPost, nil, nil, body, &result)

	return result, err
}

func (c *Client) GetSolanaTransaction(sig string) (SolanaParsedTransactionResult, error) {

	si := c.createSolanaInput()
	si.Method = "getTransaction"
	si.Params = append(si.Params, sig)

	ed := MyEncoding{Encoding: "jsonParsed", MaxSupportedTransactionVersion: 2}
	si.Params = append(si.Params, ed)

	result := SolanaParsedTransactionResult{}
	body, err := json.Marshal(si)
	if err != nil {
		return result, err
	}
	url := fmt.Sprintf("%s/%s", c.baseURL, c.API_KEY)
	core.DoHttpRequest(url, http.MethodPost, nil, nil, body, &result)
	return result, nil
}

func (c *Client) createSolanaInput() solanaInput {
	return solanaInput{Id: 1, Jsonrpc: "2.0"}
}
