// Copyright 2017 Aleksey Blinov. All rights reserved.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/baobabus/go-apnsmock/apns2mock"
	"golang.org/x/net/http2"
)

const usageStr = `
APNS Mock -- Emulator of Apple Push Notification HTTP/2 service.

Usage:

  go-apnsmock <flags>

Flags:
`

func main() {

	lb := apns2mock.LoopbackAddr()

	fs := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	addr := fs.String("addr", lb+":8443", "network `address` to serve on")
	certFile := fs.String("cert", "certs/server.crt", "`path` to server TLS certificate")
	keyFile := fs.String("key", "certs/server.key", "`path` to TLS certificate key")
	allOk := fs.Bool("allok", false, "if allok is true, server will respond with 200 status to all requests")
	verbose := fs.Bool("verbose", false, "if true, verbose enables http2 verbose logging")
	streams := fs.Uint("streams", 500, "`number` of concurrent HTTP/2 streams")
	conns := fs.Uint("conns", 5, "maximum `number` of concurrent HTTP/2 connections")
	cdelay := fs.Duration("conn-delay", 100*time.Millisecond, "amount of `time` by which client connect attempts should be delayed")
	rdelay := fs.Duration("resp-delay", 5*time.Millisecond, "amount of `time` by which responses should be delayed")
	usage := func() {
		fmt.Fprintln(os.Stderr, usageStr)
		fs.PrintDefaults()
	}
	fs.Usage = usage

	fs.Parse(os.Args[1:])
	if args := fs.Args(); len(args) > 0 {
		// TODO Add command for printing useful info.
		fmt.Fprintf(os.Stderr, "argument provided but not defined: %v\n", args[0])
		usage()
		os.Exit(3)
	}

	flag.Set("httptest.serve", *addr)
	commsCfg := apns2mock.CommsCfg{
		MaxConcurrentStreams: uint32(*streams),
		MaxConns:             uint32(*conns),
		ConnectionDelay:      *cdelay,
		ResponseTime:         *rdelay,
	}
	http2.VerboseLogs = *verbose
	var handler http.Handler = apns2mock.DefaultHandler
	if *allOk {
		handler = apns2mock.AllOkayHandler
	}

	fmt.Fprintf(os.Stderr, "Using certificate %#v with key %#v\n", *certFile, *keyFile)

	srv, err := apns2mock.NewServer(commsCfg, handler, *certFile, *keyFile)
	if err != nil {
		log.Fatal(err)
	}
	defer srv.Close()

	fmt.Fprintln(os.Stderr, "Serving on ", *addr)
	fmt.Fprintln(os.Stderr, "Press Ctrl+C to stop...")

	select {}
}
