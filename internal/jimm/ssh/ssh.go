// Copyright 2025 Canonical.

package ssh

import (
	"context"

	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/internal/openfga"
)

type identityManager interface {
	FetchIdentity(ctx context.Context, id string) (*openfga.User, error)
}

type modelManager interface {
	GetModel(ctx context.Context, uuid string) (dbmodel.Model, error)
}

type sshKeyManager interface {
	VerifyPublicKey(ctx context.Context, user *openfga.User, fingerprint string) (bool, error)
}

// sshManager provides a means to manage ssh server within JIMM.
type sshManager struct {
	modelManager
	identityManager
	sshKeyManager
}

// NewSSHManager returns a new SSHManager that offers jimm functionality to the SSHJumpServer.
func NewSSHManager(identityManager identityManager, modelManager modelManager, sshKeyManager sshKeyManager) (*sshManager, error) {
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
		modelManager:    modelManager,
		identityManager: identityManager,
		sshKeyManager:   sshKeyManager,
	}, nil
}
