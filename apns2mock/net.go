// Copyright 2017 Aleksey Blinov. All rights reserved.

package apns2mock

import (
	"net"
	"sync"
	"time"
)

// cappedConnListener extends net.Listener by adding the ability to limit
// the number of concurrent connections that it accepts. Calls to Accept()
// can additionally be delayed by a specified amount.
type cappedConnListener struct {
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
func (l *cappedConnListener) Accept() (res net.Conn, err error) {
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
	l *cappedConnListener
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
