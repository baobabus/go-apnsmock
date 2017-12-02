// Copyright 2017 Aleksey Blinov. All rights reserved.

package apns2mock

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/satori/go.uuid"
)

// APNSRequest represents parsed and validated request to APN service.
type APNSRequest struct {

	// DeviceToken is device token from the original request path.
	DeviceToken string

	// Header is http.Header of the original request.
	Header http.Header

	// TokenHeader is parsed headers of JWT provider token.
	TokenHeader map[string]interface{}

	// TokenClaims is parsed claims of JWT provider token.
	TokenClaims map[string]interface{}

	// Payload is parsed request payload.
	Payload map[string]interface{}
}

type HadlerFunc func(req *APNSRequest) (statusCode int, rejectionReason string)

// AllOkayHandler always respons with status 200 and a valid APN ID.
var AllOkayHandler = allOkayHandler{}

type allOkayHandler struct{}

func (h allOkayHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeApnsId(w, r)
	respSucc(w)
}

var (
	regEx_DeviceToken = regexp.MustCompile("[[:xdigit:]]+")
)

// CaseHandler handles all requests rooted at /3/device/ path.
// It validates all request attributes, including method,
// required headers, request path and payload body.
// JSON encoded values are unmashalled and related errors
// are handled. If all of the above checks are successful,
// parsed APNSRequest is evaluated by hander's case handlers.
//
// You can choose from a set of predefined case handlers or
// define some of your own. Case handler slices can be joined
// together to form larger combined lists.
type CaseHandler struct {
	// CaseHandlers are asked to evaluate APNSRequest in order of
	// their appearance. The first case handler returning non-zero status code
	// is selected and its rejection reason is sent in the response.
	//
	// If none of the case handlers return non-zero status code, as 200
	// successful response is sent back to the client.
	CaseHandlers []HadlerFunc
}

// ServeHTTP serves all incoming HTTP requests. It performs initial
// request validation and parsing before delegating to CaseHandlers.
func (h *CaseHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	writeApnsId(w, r)
	if strings.ToUpper(r.Method) != "POST" {
		h.respErr(w, 405, "MethodNotAllowed")
		return
	}
	dt := r.URL.Path[len(RequestRoot):]
	if !regEx_DeviceToken.MatchString(dt) {
		h.respErr(w, 400, "BadDeviceToken")
		return
	}
	ah := r.Header.Get("authorization")
	if !strings.HasPrefix(ah, "bearer ") {
		h.respErr(w, 403, "MissingProviderToken")
		return
	}
	pt := strings.TrimSpace(ah[len("bearer "):])
	if len(pt) == 0 {
		h.respErr(w, 403, "InvalidProviderToken")
		return
	}
	ts := strings.Split(pt, ".")
	if len(ts) != 3 {
		h.respErr(w, 403, "InvalidProviderToken")
		return
	}
	thb, derr := jwt.DecodeSegment(ts[0])
	var th map[string]interface{}
	uerr := json.Unmarshal(thb, &th)
	if derr != nil || uerr != nil {
		h.respErr(w, 403, "InvalidProviderToken")
		return
	}
	tcb, derr := jwt.DecodeSegment(ts[1])
	var tc map[string]interface{}
	uerr = json.Unmarshal(tcb, &tc)
	if derr != nil || uerr != nil {
		h.respErr(w, 403, "InvalidProviderToken")
		return
	}
	bb, err := ioutil.ReadAll(r.Body)
	if err != nil || len(bb) == 0 {
		h.respErr(w, 400, "PayloadEmpty")
		return
	}
	if len(bb) > 4096 {
		h.respErr(w, 400, "PayloadEmpty") // Need a different reason?
		return
	}
	req := &APNSRequest{DeviceToken: dt, Header: r.Header, TokenHeader: th, TokenClaims: tc, Payload: nil}
	for _, ch := range h.CaseHandlers {
		if status, reason := ch(req); status > 0 {
			h.respErr(w, status, reason)
			return
		}
	}
	h.respSucc(w)
	return
}

func writeApnsId(w http.ResponseWriter, r *http.Request) {
	id := r.Header.Get("apns-id")
	if id == "" {
		id = uuid.NewV4().String()
	}
	w.Header().Set("apns-id", id)
}

func (h *CaseHandler) respErr(w http.ResponseWriter, status int, reason string) {
	respErr(w, status, reason)
}

func respErr(w http.ResponseWriter, status int, reason string) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, "{\"reason\": \"%v\"}", reason)
}

func (h *CaseHandler) respSucc(w http.ResponseWriter) {
	respSucc(w)
}

func respSucc(w http.ResponseWriter) {
	w.WriteHeader(200)
}
