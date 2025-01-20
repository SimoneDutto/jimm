// Copyright 2025 Canonical.
package ssh

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/juju/zaputil/zapctx"
	"go.uber.org/zap"
	gossh "golang.org/x/crypto/ssh"
)

// rejectConnectionAndLogError logs the error and rejects the channel with a message.
func rejectConnectionAndLogError(ctx context.Context, newChan gossh.NewChannel, msg string, err error) {
	zapctx.Error(ctx, msg, zap.Error(err))
	err = newChan.Reject(gossh.ConnectionFailed, msg)
	if err != nil {
		zapctx.Error(ctx, msg, zap.Error(err))
	}
}

// dialControllerSSHServer dials the controller ssh server, trying the addresses sequentially and returning a go ssh client.
func dialControllerSSHServer(addrs []string, destPort uint32) (*gossh.Client, error) {
	var client *gossh.Client
	var err error
	var errors []error
	for _, addr := range addrs {
		dest := net.JoinHostPort(addr, fmt.Sprint(destPort))
		client, err = gossh.Dial("tcp", dest, &gossh.ClientConfig{
			//nolint:gosec // this will be removed once we handle hostkeys
			HostKeyCallback: gossh.InsecureIgnoreHostKey(),
			Auth: []gossh.AuthMethod{
				gossh.PasswordCallback(func() (secret string, err error) {
					return "jwt", nil
				}),
			},
			Timeout: 5 * time.Second,
		})
		if err != nil {
			errors = append(errors, err)
		}
	}
	if client == nil {
		return nil, fmt.Errorf("failed to dial controller ssh server: %v", errors)
	}
	return client, nil
}
