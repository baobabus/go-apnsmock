// Copyright 2017 Aleksey Blinov. All rights reserved.

package apns2mock

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"time"

	"golang.org/x/net/http2"
)

// APNS default root URL path.
const RequestRoot = "/3/device/"

// CommsCfg can be used to configure communications aspects of the mock server.
type CommsCfg struct {

	// MaxConcurrentStreams is the maximum allowed number of concurrent streams
	// per HTTP/2 connection. This will be communicated to clients in HTTP/2
	// SETTINGS frame.
	MaxConcurrentStreams uint32

	// MaxConns is maximum allowed number of concurrent connections. Attempted
	// connections over this limit will be dropped by the server.
	MaxConns uint32

	// ConnectionDelay delays connection accept by the specified time.
	ConnectionDelay time.Duration

	// ResponseTime is the time to be taken to respond to a request other than
	// 404 BadPath response.
	ResponseTime time.Duration
}

// TypicalCommsCfg contains settings that emulate typical latency and
// connection handling behavior of actual APNS servers.
var TypicalCommsCfg = CommsCfg{
	MaxConcurrentStreams: 500,
	MaxConns:             1000,
	ConnectionDelay:      1 * time.Second,
	ResponseTime:         20 * time.Millisecond,
}

// NoDelayCommsCfg contains settings that do not introduce any additional delay
// in the mock server handling of the requests and incoming client connections.
var NoDelayCommsCfg = CommsCfg{
	MaxConcurrentStreams: 500,
	MaxConns:             1000,
	ConnectionDelay:      0,
	ResponseTime:         0,
}

// AutoCert can be supplied to NewServer as certFile argument instead of ""
// for improved semantics indicating that server certificate and key
// should be auto-generated.
const AutoCert = ""

// AutoKey can be supplied to NewServer as keyFile argument instead of ""
// for improved semantics indicating that server certificate and key
// should be auto-generated.
const AutoKey = ""

// Server represents a mock APNS service. See NewServer for information
// on creating servers.
type Server struct {

	// We are wrapping httptest.Server in order to extend functionality
	// and to backfill features missing in go 1.7.
	*httptest.Server

	// RootCertificate to use when setting up client TLS's RootCAs.
	RootCertificate *tls.Certificate

	// Preconfigured client. This supercedes httptest.Server's client
	// in go 1.8 and later.
	client *http.Client

	interceptor *atomic.Value
}

// NewServer creates and starts a new Server instance with handler servicing
// requests on RequestRoot and with all other paths returning 404 status.
// If certFile and keyFile are not empty, the server TLS certicicate will be
// loaded from the specified files.
//
// If either certFile or keyFile is empty a new self-signed certificate
// will be created. This is usually all that is needed for integrating
// mock server in automated tests. Simply use server's pre-configured client
// for your testing or retrieve server's URL and root certificate to configure
// your custom client.
func NewServer(commsCfg CommsCfg, handler http.Handler, certFile string, keyFile string) (*Server, error) {
	if handler == nil {
		return nil, errors.New("apns2mock: no handler supplied.")
	}
	mux := http.NewServeMux()
	itcpr := &atomic.Value{}
	tryIntercept := func(w http.ResponseWriter) bool {
		if ihi := itcpr.Load(); ihi != nil {
			ih := ihi.(func(w http.ResponseWriter) bool)
			return ih(w)
		}
		return false
	}
	mux.HandleFunc(RequestRoot, func(w http.ResponseWriter, r *http.Request) {
		if tryIntercept(w) {
			return
		}
		if commsCfg.ResponseTime > 0 {
			time.Sleep(commsCfg.ResponseTime)
		}
		handler.ServeHTTP(w, r)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if tryIntercept(w) {
			return
		}
		writeApnsId(w, r)
		respErr(w, 404, "BadPath")
	})
	srv := httptest.NewUnstartedServer(mux)
	srv.Listener = &cappedConnListener{
		Listener: srv.Listener,
		Cap:      commsCfg.MaxConns,
		Delay:    commsCfg.ConnectionDelay,
	}
	http2Conf := &http2.Server{
		MaxConcurrentStreams: commsCfg.MaxConcurrentStreams,
	}
	if err := http2.ConfigureServer(srv.Config, http2Conf); err != nil {
		return nil, err
	}
	srv.TLS = &tls.Config{
		CipherSuites: []uint16{tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
		NextProtos:   []string{http2.NextProtoTLS},
	}
	if certFile != "" && keyFile != "" {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return nil, err
		}
		srv.TLS.Certificates = []tls.Certificate{cert}
	}
	srv.StartTLS()
	res := &Server{
		Server:          srv,
		RootCertificate: &srv.TLS.Certificates[0],
		client:          makeClient(&srv.TLS.Certificates[0]),
		interceptor:     itcpr,
	}
	return res, nil
}

// Client returns an HTTP client configured for making requests to the server.
// It is configured to trust the server's TLS test certificate and will close
// its idle connections on Server.Close.
func (s *Server) Client() *http.Client {
	return s.client
}

// BecomeUnavailable makes server begin responding with specified status code
// and reason to any future requests. This is typically used to test handling
// of 5XX status codes by clients.
func (s *Server) BecomeUnavailable(statusCode int, reason string) {
	s.interceptor.Store(func(w http.ResponseWriter) bool {
		respErr(w, statusCode, reason)
		return true
	})
}

// BecomeAvailable restores normal request handling flow.
func (s *Server) BecomeAvailable() {
	s.interceptor.Store(func(w http.ResponseWriter) bool {
		return false
	})
}

func makeClient(cert *tls.Certificate) *http.Client {
	// httptest.Server.Certificate() is not available in go 1.7,
	// so we must to it the hard way.
	// This will not error out as the same cert was just parsed
	// while creating the server.
	rCert, _ := x509.ParseCertificate(cert.Certificate[0])
	certpool := x509.NewCertPool()
	certpool.AddCert(rCert)
	res := &http.Client{
		Transport: &http2.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: certpool,
			},
		},
	}
	return res
}
