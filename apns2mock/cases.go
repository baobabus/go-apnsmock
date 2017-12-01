// Copyright 2017 Aleksey Blinov. All rights reserved.

package apns2mock

// TokenHandler only handles provider token-based requests.
var TokenAuthHandler = &CaseHandler{
	CaseHandlers: JoinHandlers(HeaderHandlers, DeviceTokenHandlers, AuthTokenHandlers),
}

// CertAuthHandler only handles client certificate-based requests.
// TODO Implement and add CertHandlers.
var CertAuthHandler = &CaseHandler{
	CaseHandlers: JoinHandlers(HeaderHandlers, DeviceTokenHandlers),
}

// DefaultHandler handles provider token-based and
// client certificate-based requests.
// TODO Implement and add CertHandlers and combination handlers.
var DefaultHandler = &CaseHandler{
	CaseHandlers: JoinHandlers(HeaderHandlers, DeviceTokenHandlers, AuthTokenHandlers),
}

// JoinHandlers is a convenience function that joins all supplied handlers
// and returns the combine slice.
func JoinHandlers(hs ...[]func(req *APNSRequest) (statusCode int, rejectionReason string)) []func(req *APNSRequest) (statusCode int, rejectionReason string) {
	res := []func(req *APNSRequest) (statusCode int, rejectionReason string){}
	for _, h := range hs {
		res = append(res, h...)
	}
	return res
}
