// Copyright 2017 Aleksey Blinov. All rights reserved.

package apns2mock

// TokenHandler only handles provider token-based requests.
var TokenAuthHandler *CaseHandler

// CertAuthHandler only handles client certificate-based requests.
var CertAuthHandler *CaseHandler

// DefaultHandler handles provider token-based and
// client certificate-based requests.
var DefaultHandler *CaseHandler

// JoinHandlers is a convenience function that joins all supplied handlers
// and returns the combine slice.
func JoinHandlers(hs ...[]HadlerFunc) []HadlerFunc {
	res := []HadlerFunc{}
	for _, h := range hs {
		res = append(res, h...)
	}
	return res
}

func init() {
	TokenAuthHandler = &CaseHandler{
		CaseHandlers: JoinHandlers(HeaderHandlers, DeviceTokenHandlers, AuthTokenHandlers),
	}
	// TODO Implement and add CertHandlers.
	CertAuthHandler = &CaseHandler{
		CaseHandlers: JoinHandlers(HeaderHandlers, DeviceTokenHandlers),
	}
	// TODO Implement and add CertHandlers and combination handlers.
	DefaultHandler = &CaseHandler{
		CaseHandlers: JoinHandlers(HeaderHandlers, DeviceTokenHandlers, AuthTokenHandlers),
	}
}
