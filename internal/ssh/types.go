// Copyright 2025 Canonical.

package ssh

import "github.com/gliderlabs/ssh"

// fowardMessage is the struct holding the information about the jump message received by the ssh client.
type forwardMessage struct {
	DestAddr string
	DestPort uint32
	SrcAddr  string
	SrcPort  uint32
}

// Server is the custom struct to embed the gliderlabs.ssh server and a resolver.
type Server struct {
	*ssh.Server

	resolver Resolver
}

type Config struct {
	Port                     string
	HostKey                  []byte
	MaxConcurrentConnections string
}
