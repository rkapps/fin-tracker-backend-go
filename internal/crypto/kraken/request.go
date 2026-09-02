package kraken

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/internal/core"
)

// Execute a public method query
func doRequestPublic(baseURL, version, method string, dst interface{}) (string, error) {
	reqURL := fmt.Sprintf("%s/%s/public/%s", baseURL, version, method)

	var env struct {
		Error  []string    `json:"error"`
		Result interface{} `json:"result"`
	}
	env.Result = dst
	var body string
	if body, err := core.DoHttpRequest(reqURL, http.MethodPost, nil, nil, nil, &env); err != nil {
		return body, err
	}
	if len(env.Error) > 0 {
		return "", fmt.Errorf("kraken error: %v", env.Error)
	}
	return body, nil
}

// doRequestPrivate executes a private, signed method query
func doRequestPrivate(cfg Config, baseURL, version, method string, values url.Values, dst interface{}) (string, error) {
	urlPath := fmt.Sprintf("/%s/private/%s", version, method)
	reqURL := fmt.Sprintf("%s%s", baseURL, urlPath)

	secret, _ := base64.StdEncoding.DecodeString(cfg.Api_Secret)
	values.Set("nonce", fmt.Sprintf("%d", time.Now().UnixNano()))

	encoded := values.Encode() // computed once — used for BOTH the signature and the body
	signature := getKrakenSignature(urlPath, values, encoded, secret)

	headers := url.Values{
		"API-Key":      {cfg.Api_Key},
		"API-Sign":     {signature},
		"Content-Type": {"application/x-www-form-urlencoded"},
	}

	var env struct {
		Error  []string    `json:"error"`
		Result interface{} `json:"result"`
	}
	env.Result = dst
	var body string
	if body, err := core.DoHttpRequest(reqURL, http.MethodPost, headers, nil, []byte(encoded), &env); err != nil {
		return body, err
	}
	if len(env.Error) > 0 {
		return body, fmt.Errorf("kraken error: %v", env.Error)
	}
	return body, nil
}

func getKrakenSignature(urlPath string, values url.Values, encoded string, secret []byte) string {
	sha := sha256.New()
	sha.Write([]byte(values.Get("nonce") + encoded))
	shasum := sha.Sum(nil)

	mac := hmac.New(sha512.New, secret)
	mac.Write(append([]byte(urlPath), shasum...))
	macsum := mac.Sum(nil)
	return base64.StdEncoding.EncodeToString(macsum)
}
