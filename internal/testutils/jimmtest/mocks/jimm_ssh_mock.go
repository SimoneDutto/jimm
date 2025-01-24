// Copyright 2025 Canonical.

package mocks

import (
	"context"

	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/internal/openfga"
)

// SSHManager is an implementation of the SshManager interface.
type SSHManager struct {
	PublicKeyHandler_      func(ctx context.Context, claimUser string, key []byte) (*openfga.User, error)
	ConnInfoFromModelUUID_ func(ctx context.Context, modelUUID string, user *openfga.User) ([]string, string, error)
}

func (j SSHManager) PublicKeyHandler(ctx context.Context, claimUser string, key []byte) (*openfga.User, error) {
	if j.PublicKeyHandler_ == nil {
		return nil, errors.E(errors.CodeNotImplemented)
	}
	return j.PublicKeyHandler_(ctx, claimUser, key)
}

func (j SSHManager) ConnInfoFromModelUUID(ctx context.Context, modelUUID string, user *openfga.User) ([]string, string, error) {
	if j.ConnInfoFromModelUUID_ == nil {
		return nil, "", errors.E(errors.CodeNotImplemented)
	}
	return j.ConnInfoFromModelUUID_(ctx, modelUUID, user)
}
