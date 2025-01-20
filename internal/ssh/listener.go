// Copyright 2025 Canonical.
package ssh

import (
	"net"
	"sync"
	"time"
)

// N.B.:
// This is a copypaste of netutil.LimiLister (link: https://cs.opensource.google/go/x/net/+/refs/tags/v0.34.0:netutil/listen.go),
// but we add a timeout so when we are at the limit we actively close connections instead of waiting indefinetely. (Look at line 44)

// LimitListenerWithTimeout returns a Listener that accepts at most n simultaneous
// connections from the provided Listener, and it timeouts when the max
// has been reached and no seats has been freed for the timeout period.
func LimitListenerWithTimeout(l net.Listener, n int, timeout time.Duration) net.Listener {
	return &limitListener{
		Listener: l,
		sem:      make(chan struct{}, n),
		done:     make(chan struct{}),
		timeout:  timeout,
	}
}

type limitListener struct {
	net.Listener
	sem       chan struct{}
	closeOnce sync.Once     // ensures the done chan is only closed once
	done      chan struct{} // no values sent; closed when Close is called
	timeout   time.Duration // timeout for acquiring the connection
}

// acquire acquires the limiting semaphore. Returns true if successfully
// acquired, false if the listener is closed and the semaphore is not
// acquired.
func (l *limitListener) acquire() bool {
	select {
	case <-l.done:
		return false
	case l.sem <- struct{}{}:
		return true
	// we add a timeout here, so the connection is closed when the timeout has passed instead of waiting.
	case <-time.After(l.timeout):
		return false
	}
}
func (l *limitListener) release() { <-l.sem }

func (l *limitListener) Accept() (net.Conn, error) {
	if !l.acquire() {
		// If the semaphore isn't acquired because the listener was closed, expect
		// that this call to accept won't block, but immediately return an error.
		// If it instead returns a spurious connection (due to a bug in the
		// Listener, such as https://golang.org/issue/50216), we immediately close
		// it and try again. Some buggy Listener implementations (like the one in
		// the aforementioned issue) seem to assume that Accept will be called to
		// completion, and may otherwise fail to clean up the client end of pending
		// connections.
		for {
			c, err := l.Listener.Accept()
			if err != nil {
				return nil, err
			}
			c.Close()
		}
	}

	c, err := l.Listener.Accept()
	if err != nil {
		l.release()
		return nil, err
	}
	return &limitListenerConn{Conn: c, release: l.release}, nil
}

func (l *limitListener) Close() error {
	err := l.Listener.Close()
	l.closeOnce.Do(func() { close(l.done) })
	return err
}

type limitListenerConn struct {
	net.Conn
	releaseOnce sync.Once
	release     func()
}

func (l *limitListenerConn) Close() error {
	err := l.Conn.Close()
	l.releaseOnce.Do(l.release)
	return err
}
