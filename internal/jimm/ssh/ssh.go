// Copyright 2025 Canonical.

package ssh

import (
	"context"

	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/internal/openfga"
)

// IdentityManager provides a means to fetch an identity from the identity service.
type IdentityManager interface {
	FetchIdentity(ctx context.Context, id string) (*openfga.User, error)
}

// ModelManager provides a means to fetch a model from the model service.
type ModelManager interface {
	GetModel(ctx context.Context, uuid string) (dbmodel.Model, error)
}

// SSHKeyManager provides a means to manage ssh keys within JIMM.
type SSHKeyManager interface {
	VerifyPublicKey(ctx context.Context, claimUser string, publicKey []byte) (bool, error)
}

// sshManager provides a means to manage ssh server within JIMM.
type sshManager struct {
	ModelManager
	IdentityManager
	SSHKeyManager
}

// NewSSHManager returns a new SSHManager that offers jimm functionality to the SSHJumpServer.
func NewSSHManager(identityManager IdentityManager, modelManager ModelManager, sshKeyManager SSHKeyManager) (*sshManager, error) {
	if identityManager == nil {
		return nil, errors.E("identityManager cannot be nil")
	}
	if modelManager == nil {
		return nil, errors.E("modelManager cannot be nil")
	}
	if sshKeyManager == nil {
		return nil, errors.E("sshManager cannot be nil")
	}
	return &sshManager{
		ModelManager:    modelManager,
		IdentityManager: identityManager,
		SSHKeyManager:   sshKeyManager,
	}, nil
}
