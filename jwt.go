package main

import (
	"encoding/json"
	"errors"
	"fmt"
	uuid2 "github.com/google/uuid"
	"gopkg.in/square/go-jose.v2"
	"strings"
)

type OAuth2ServiceJwt struct {
	//TODO: Awaiting actual shape of this from Enrique
	Sub            string
	Name           string
	SubscriptionId *uuid2.UUID `json:"subscription_id"`
	AndroidId      *string     `json:"android_id"`
}

func getOAuth2ServiceJwt(authorizationHeaders []string) (*OAuth2ServiceJwt, error) {
	if len(authorizationHeaders) == 0 {
		return nil, errors.New("no authorization header provided")
	}

	jwt, err := jose.ParseSigned(strings.Replace(authorizationHeaders[0], "Bearer ", "", 1))
	if err != nil {
		return nil, fmt.Errorf("unable to parse JWT from Autorization header: %w", err)
	}

	//Caution: This assumes that the authentication middleware has verified the JWT's signature against the
	//		   OAuth2 server's public key already.

	var result OAuth2ServiceJwt
	err = json.Unmarshal(jwt.UnsafePayloadWithoutVerification(), &result)
	if err != nil {
		return nil, fmt.Errorf("unable to deserialize JWT payload: %w", err)
	}

	return &result, nil
}
