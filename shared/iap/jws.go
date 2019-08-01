package iap

import (
	"crypto/rsa"
	"time"

	"golang.org/x/oauth2/jws"
)

type jwsClaim struct {
	IssuerEmail string
	Audience    string
	PrivateKey  *rsa.PrivateKey
	ClientID    string
}

func (j *jwsClaim) getJWS() (string, error) {
	iat := time.Now()
	exp := iat.Add(time.Minute * 10)
	jwt := &jws.ClaimSet{
		Iss: j.IssuerEmail,
		Aud: j.Audience,
		Iat: iat.Unix(),
		Exp: exp.Unix(),
		PrivateClaims: map[string]interface{}{
			"target_audience": j.ClientID,
		},
	}
	jwsHeader := &jws.Header{
		Algorithm: "RS256",
		Typ:       "JWT",
	}

	return jws.Encode(jwsHeader, jwt, j.PrivateKey)
}
