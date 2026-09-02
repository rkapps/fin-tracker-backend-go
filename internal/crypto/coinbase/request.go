package coinbase

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math"
	"math/big"
	"net/url"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v4"
	"github.com/go-jose/go-jose/v4/jwt"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
)

type APIKeyClaims struct {
	*jwt.Claims
	URI string `json:"uri"`
}

func buildJWT(uri string, keyName string, keySecret string) (string, error) {

	keyStr := keySecret
	keyStr = strings.ReplaceAll(keySecret, "\\n", "\n")
	privatePEM := []byte(keyStr)
	block, _ := pem.Decode(privatePEM)
	if block == nil {
		return "", fmt.Errorf("jwt: Could not decode private key")
	}

	key, err := x509.ParseECPrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("jwt: %w", err)
	}

	sig, err := jose.NewSigner(
		jose.SigningKey{Algorithm: jose.ES256, Key: key},
		(&jose.SignerOptions{NonceSource: nonceSource{}}).WithType("JWT").WithHeader("kid", keyName),
	)
	if err != nil {
		return "", fmt.Errorf("jwt: %w", err)
	}

	cl := &APIKeyClaims{
		Claims: &jwt.Claims{
			Subject:   keyName,
			Issuer:    "cdp",
			NotBefore: jwt.NewNumericDate(time.Now()),
			Expiry:    jwt.NewNumericDate(time.Now().Add(2 * time.Minute)),
		},
		URI: uri,
	}
	jwtString, err := jwt.Signed(sig).Claims(cl).Serialize()
	if err != nil {
		return "", fmt.Errorf("jwt: %w", err)
	}
	return jwtString, nil
}

var max = big.NewInt(math.MaxInt64)

type nonceSource struct{}

func (n nonceSource) Nonce() (string, error) {
	r, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return r.String(), nil
}

func doNewRequest(config Config, method string, host string, baseUrl string, requestPath string, requestFullPath string, dst interface{}) (string, error) {
	uri := fmt.Sprintf("%s %s%s", method, host, requestPath)
	jwtstring, err := buildJWT(uri, config.KeyName, config.PrivateKey)
	if err != nil {
		return "", err
	}

	fullURL := baseUrl + requestFullPath
	headers := url.Values{"Authorization": {"Bearer " + jwtstring}}

	return core.DoHttpRequest(fullURL, method, headers, nil, nil, dst)

}
