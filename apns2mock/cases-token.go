// Copyright 2017 Aleksey Blinov. All rights reserved.

package apns2mock

import (
	"time"
)

// AuthTokenHandlers deal with provider JWT tokens.
//
// Expired tokens are 403, "ExpiredProviderToken".
//
// Tokens with incorrct signing algorithm are 403, "InvalidProviderToken".
//
// Team ID (JWT "iss" claim) starting with '1' are 403, "InvalidProviderToken".
var AuthTokenHandlers []HadlerFunc

func init() {
	AuthTokenHandlers = []HadlerFunc{
		func(req *APNSRequest) (int, string) {
			if v, ok := req.TokenClaims["iat"]; !ok || int64(v.(float64)) < time.Now().Add(-1*time.Hour).Unix() {
				return 403, "ExpiredProviderToken"
			}
			if v, ok := req.TokenHeader["alg"]; !ok || v.(string) != "ES256" {
				return 403, "InvalidProviderToken"
			}
			if v, ok := req.TokenClaims["iss"]; !ok || v.(string)[0] == '1' {
				return 403, "InvalidProviderToken"
			}
			return 0, ""
		},
	}
}
