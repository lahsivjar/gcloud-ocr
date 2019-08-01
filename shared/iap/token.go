package iap

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
)

type tokenClient struct {
	HTTPClient *http.Client
}

func (t *tokenClient) refresh(jwsassertion string) (string, error) {

	params := url.Values{}
	params.Set("grant_type", jwtBearerType)
	params.Set("assertion", jwsassertion)

	resp, err := t.HTTPClient.PostForm(tokenURI, params)
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
