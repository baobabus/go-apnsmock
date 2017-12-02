// Copyright 2017 Aleksey Blinov. All rights reserved.

// go-apnsmock is a command line tool that can be used to run a standalone
// APNS emulator.
//
// Usage:
// 
//   go-apnsmock <flags>
// 
// Flags:
// 
//   -addr address
//     	network address to serve on (default "127.0.0.1:8443")
//   -allok
//     	if allok is true, server will respond with 200 status to all requests
//   -cert path
//     	path to server TLS certificate (default "certs/server.crt")
//   -conn-delay time
//     	amount of time by which client connect attempts should be delayed (default 100ms)
//   -conns number
//     	maximum number of concurrent HTTP/2 connections (default 5)
//   -key path
//     	path to TLS certificate key (default "certs/server.key")
//   -resp-delay time
//     	amount of time by which responses should be delayed (default 5ms)
//   -streams number
//     	number of concurrent HTTP/2 streams (default 500)
//   -verbose
//     	if true, verbose enables http2 verbose logging
package main

import (
	"flag"
	"fmt"
	"log"
	"net"
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

// loopbackAddr returns loopback interface IP address as a string or empty
// string if no loopback interface can be found.
// loopbackAddr prefers IP4 addresses over IP6.
func loopbackAddr() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			addrs, err := iface.Addrs()
			if err != nil {
				return ""
			}
			ip4 := ""
			ip6 := ""
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}
				if t := ip.To4(); len(t) == net.IPv4len {
					ip4 = t.String()
				} else {
					ip6 = t.String()
				}
			}
			if ip4 != "" {
				return ip4
			}
			return ip6
		}
	}
	return ""
}

func main() {

	lb := loopbackAddr()
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
