// Copyright 2025 Canonical.
package ssh

import (
	"fmt"
	"net"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

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
