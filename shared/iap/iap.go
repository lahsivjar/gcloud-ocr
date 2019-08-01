package iap

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"golang.org/x/oauth2/google"
)

const (
	tokenURI      = "https://www.googleapis.com/oauth2/v4/token"
	jwtBearerType = "urn:ietf:params:oauth:grant-type:jwt-bearer"
)

// IAP represents the information needed to access IAP-protected app
type IAP struct {
	ServiceAccount string
	ClientID       string
	JWSClaim       jwsClaim
	HTTPClient     *http.Client
}

// Returns a new instance of IAP
func New(hc *http.Client, sa, id string) (*IAP, error) {
	conf, err := google.JWTConfigFromJSON([]byte(sa))
	if err != nil {
		return nil, err
	}

	rsaKey, err := parsePrivateKey(conf.PrivateKey)
	if err != nil {
		return nil, err
	}

	return &IAP{
		ServiceAccount: sa,
		ClientID:       id,
		HTTPClient:     hc,
		JWSClaim: jwsClaim{
			IssuerEmail: conf.Email,
			Audience:    tokenURI,
			PrivateKey:  rsaKey,
			ClientID:    id,
		},
	}, nil
}

// Token returns a bearer token to be used with IAP protected endpoint
func (c *IAP) Token() (string, error) {
	assertionMsg, err := c.JWSClaim.getJWS()
	if err != nil {
		return "", err
	}

	return c.refreshToken(assertionMsg)
}

func (c *IAP) refreshToken(jwsassertion string) (string, error) {

	params := url.Values{}
	params.Set("grant_type", jwtBearerType)
	params.Set("assertion", jwsassertion)

	resp, err := c.HTTPClient.PostForm(tokenURI, params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	var tokenRes struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		IDToken     string `json:"id_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}

	if err := json.Unmarshal(body, &tokenRes); err != nil {
		return "", err
	}

	return tokenRes.IDToken, nil
}

func parsePrivateKey(bytes []byte) (*rsa.PrivateKey, error) {
	var key *rsa.PrivateKey
	var err error
	block, _ := pem.Decode(bytes)
	if block == nil {
		return nil, fmt.Errorf("invalid private key data")
	}

	if block.Type == "RSA PRIVATE KEY" {
		key, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
	} else if block.Type == "PRIVATE KEY" {
		keyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		var ok bool
		key, ok = keyInterface.(*rsa.PrivateKey)
		if !ok {
			return nil, errors.New("not RSA private key")
		}
	} else {
		return nil, fmt.Errorf("invalid private key type: %s", block.Type)
	}

	key.Precompute()

	if err := key.Validate(); err != nil {
		return nil, err
	}
	return key, nil
}
