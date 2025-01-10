// Copyright 2024 Canonical.
package utils

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net"

	"github.com/juju/zaputil/zapctx"
	"go.uber.org/zap"
)

// NewConversationID generates a unique ID that is used for the
// lifetime of a websocket connection.
func NewConversationID() string {
	buf := make([]byte, 8)
	_, err := rand.Read(buf)
	if err != nil {
		zapctx.Error(context.Background(), "failed to generate rand", zap.Error(err))

	}
	return hex.EncodeToString(buf)
}

// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreePort() (int, error) {
	if a, err := net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return 0, errors.New("Couldn't find any free port")
}
