// Copyright 2025 Canonical.

package ssh

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/gliderlabs/ssh"
	"github.com/juju/names/v5"
	"github.com/juju/zaputil/zapctx"
	"go.uber.org/zap"
	gossh "golang.org/x/crypto/ssh"

	"github.com/canonical/jimm/v3/internal/openfga"
)

// juju_ssh_default_port is the default port we expect the juju controllers to respond on.
const juju_ssh_default_port = 17022

type publicKeySSHUserKey struct{}

// SSHManager is the interface with the methods needed by the ssh jump server to route request.
type SSHManager interface {
	// AddrFromModelUUID is the method to resolve the address of the controller to contact given the model UUID.
	AddrFromModelUUID(ctx context.Context, user *openfga.User, modelTag names.ModelTag) (string, error)

	// FetchIdentity
	FetchIdentity(ctx context.Context, id string) (*openfga.User, error)

	// VerifyPublicKey verifies the identityName is
	VerifyPublicKey(ctx context.Context, user *openfga.User, fingerprint string) (bool, error)
}

// forwardMessage is the struct holding the information about the jump message received by the ssh client.
type forwardMessage struct {
	DestAddr string
	DestPort uint32
	SrcAddr  string
	SrcPort  uint32
}

// Server is the custom struct to embed the gliderlabs.ssh server and a sshManager.
type Server struct {
	*ssh.Server

	sshManager SSHManager
}

// Config is the struct holding the configuration for the jump server.
type Config struct {
	Port                     string
	HostKey                  []byte
	MaxConcurrentConnections string
}

// NewJumpServer creates the jump server struct.
func NewJumpServer(ctx context.Context, config Config, sshManager SSHManager) (Server, error) {
	zapctx.Info(ctx, "NewJumpServer")

	if sshManager == nil {
		return Server{}, fmt.Errorf("Cannot create JumpSSHServer with a nil resolver.")
	}
	server := Server{
		Server: &ssh.Server{
			Addr: fmt.Sprintf(":%s", config.Port),
			ChannelHandlers: map[string]ssh.ChannelHandler{
				"direct-tcpip": directTCPIPHandler(sshManager),
			},
			PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
				user, err := sshManager.FetchIdentity(ctx, ctx.User())
				if err != nil {
					zapctx.Info(ctx, fmt.Sprintf("cannot find user %s", ctx.User()))
					return false
				}
				if ok, err := sshManager.VerifyPublicKey(ctx, user, gossh.FingerprintSHA256(key)); !ok || err != nil {
					zapctx.Info(ctx, fmt.Sprintf("cannot verify key for user %s", ctx.User()), zap.Error(err))
					return false
				}
				ctx.SetValue(publicKeySSHUserKey{}, user)
				return true
			},
		},
		sshManager: sshManager,
	}
	s, err := gossh.ParsePrivateKey([]byte(config.HostKey))
	if err != nil {
		return Server{}, fmt.Errorf("Cannot parse hostkey.")
	}
	server.AddHostKey(s)

	return server, nil
}

func directTCPIPHandler(sshManager SSHManager) func(srv *ssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx ssh.Context) {
	return func(srv *ssh.Server, conn *gossh.ServerConn, newChan gossh.NewChannel, ctx ssh.Context) {
		d := forwardMessage{}
		k := newChan.ExtraData()

		if err := gossh.Unmarshal(k, &d); err != nil {
			rejectConnectionAndLogError(ctx, newChan, "failed to parse channel data", err)
			return
		}
		if d.DestPort == 0 {
			d.DestPort = juju_ssh_default_port
		}
		if !names.IsValidModel(d.DestAddr) {
			rejectConnectionAndLogError(ctx, newChan, "invalid model uuid", nil)
			return
		}
		modelTag := names.NewModelTag(d.DestAddr)
		user, err := fetchAndVerifySSHUser(ctx, modelTag)
		if err != nil {
			rejectConnectionAndLogError(ctx, newChan, err.Error(), err)
			return
		}
		addr, err := sshManager.AddrFromModelUUID(ctx, user, modelTag)
		if err != nil {
			rejectConnectionAndLogError(ctx, newChan, "failed to resolve address from model uuid", err)
			return
		}
		dest := net.JoinHostPort(addr, fmt.Sprint(d.DestPort))
		// this is temporary. The way we dial to the controller will heavily change.
		client, err := gossh.Dial("tcp", dest, &gossh.ClientConfig{
			//nolint:gosec // this will be removed once we handle hostkeys
			HostKeyCallback: gossh.InsecureIgnoreHostKey(),
			Auth: []gossh.AuthMethod{
				gossh.PasswordCallback(func() (secret string, err error) {
					return "jwt", nil
				}),
			},
		})
		if err != nil {
			rejectConnectionAndLogError(ctx, newChan, fmt.Sprintf("failed to connect to %s: %v", dest, err), err)
			return
		}

		dstChan, reqs, err := client.OpenChannel("direct-tcpip", gossh.Marshal(d))
		if err != nil {
			rejectConnectionAndLogError(ctx, newChan, "failed to open destination channel", err)
			return
		}
		// gossh.Request are requests sent outside of the normal stream of data (ex. pty-req for an interactive session).
		// Since we only need the raw data to redirect, we can discard them.
		go gossh.DiscardRequests(reqs)

		srcDest, reqs, err := newChan.Accept()
		if err != nil {
			dstChan.Close()
			return
		}
		// gossh.Request are requests sent outside of the normal stream of data (ex. pty-req for an interactive session).
		// Since we only need the raw data to redirect, we can discard them.
		go gossh.DiscardRequests(reqs)

		go func() {
			defer srcDest.Close()
			defer dstChan.Close()
			_, err := io.Copy(srcDest, dstChan)
			if err != nil {
				rejectConnectionAndLogError(ctx, newChan, "failed to copy data from src to dts", err)
			}
		}()
		go func() {
			defer srcDest.Close()
			defer dstChan.Close()
			_, err := io.Copy(dstChan, srcDest)
			if err != nil {
				rejectConnectionAndLogError(ctx, newChan, "failed to copy data from dst to src", err)
			}
		}()
		zapctx.Info(ctx, fmt.Sprintf("Proxying connection from %s:%d to %s:%d \n", d.SrcAddr, d.SrcPort, d.DestAddr, d.DestPort))
	}
}

// fetchAndVerifySSHUser extracts the user from the context and checks the user has permission to ssh.
func fetchAndVerifySSHUser(ctx ssh.Context, modelTag names.ModelTag) (*openfga.User, error) {
	user, ok := ctx.Value(publicKeySSHUserKey{}).(*openfga.User)
	if !ok {
		return nil, fmt.Errorf("fo user in the context")
	}
	ok, err := user.IsModelWriter(ctx, modelTag)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve address from model uuid")
	}
	if !ok {
		return nil, fmt.Errorf("user doesn't have permission")
	}
	return user, nil
}

// rejectConnectionAndLogError logs the error and rejects the channel with a message.
func rejectConnectionAndLogError(ctx context.Context, newChan gossh.NewChannel, msg string, err error) {
	zapctx.Error(ctx, msg, zap.Error(err))
	err = newChan.Reject(gossh.ConnectionFailed, msg)
	if err != nil {
		zapctx.Error(ctx, msg, zap.Error(err))
	}
}
