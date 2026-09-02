package core

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
)

// DoHttpRequest executes an HTTP request and optionally decodes the JSON response
// body into dst. Auth is entirely the caller's responsibility — pass whatever
// headers (Bearer JWT, API-Key/API-Sign, etc.) the target API needs; this
// function has no knowledge of any specific provider's signing scheme.
//
// Returns the raw response body (even on decode success) in case the caller
// wants it for logging/storage alongside the decoded value.
func DoHttpRequest(rawURL, method string, headers, query url.Values, body []byte, dst interface{}) (string, error) {

	var reqBody io.Reader
	if body != nil {
		reqBody = bytes.NewReader(body)
	}

	req, err := http.NewRequest(method, rawURL, reqBody)
	if err != nil {
		return "", fmt.Errorf("building request: %w", err)
	}

	req.Header.Add("Accept", "application/json")
	for k, values := range headers {
		for _, v := range values {
			req.Header.Add(k, v)
		}
	}

	if len(query) > 0 {
		req.URL.RawQuery = query.Encode()
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("http request error: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		slog.Error("DoRequest", "method", method, "url", rawURL, "status", resp.StatusCode, "body", string(respBody))
		return string(respBody), fmt.Errorf("http %s %s: status %d: %s", method, rawURL, resp.StatusCode, respBody)
	}

	if dst != nil {
		if err := json.Unmarshal(respBody, dst); err != nil {
			return string(respBody), fmt.Errorf("response body: %s: unmarshal error: %w", respBody, err)
		}
	}

	return string(respBody), nil
}
