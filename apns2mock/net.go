// Copyright 2017 Aleksey Blinov. All rights reserved.

package apns2mock

import (
	"net"
	"sync"
	"time"
)

// LoopbackAddr returns loopback interface IP address as a string or empty
// string if no loopback interface can be found.
// LoopbackAddr prefers IP4 addresses over IP6.
func LoopbackAddr() string {
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

// CappedConnListener extends net.Listener by adding the ability to limit
// the number of concurrent connections that it accepts. Calls to Accept()
// can additionally be delayed by a specified amount.
type CappedConnListener struct {
	net.Listener

	// Cap specifies the maximum number of concurrent connections.
	// If set to 0 all connect attempts will be rejected.
	Cap uint32

	// Delay specifies the amount of time by which Accept() calls
	// are delayed.
	Delay time.Duration

	mu  sync.Mutex
	cnt uint32
}

// See net.Listener.Accept() for more information.
func (l *CappedConnListener) Accept() (res net.Conn, err error) {
	for {
		res, err = l.Listener.Accept()
		if res == nil || err != nil {
			break
		}
		l.mu.Lock()
		hasCap := l.cnt < l.Cap
		if hasCap {
			l.cnt++
			res = &netConn{res.(*net.TCPConn), l}
			l.mu.Unlock()
			if l.Delay > 0 {
				time.Sleep(l.Delay)
			}
			return
		}
		l.mu.Unlock()
		res.Close()
	}
	return
}

type netConn struct {
	*net.TCPConn
	l *CappedConnListener
}

func (c *netConn) Close() error {
	res := c.TCPConn.Close()
	c.l.mu.Lock()
	defer c.l.mu.Unlock()
	if c.l.cnt > 0 {
		c.l.cnt--
	}
	return res
}
