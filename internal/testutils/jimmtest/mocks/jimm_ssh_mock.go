// Copyright 2025 Canonical.

package mocks

import (
	"context"

	"github.com/juju/names/v5"

	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/internal/openfga"
)

// SSHManager is an implementation of the sshManager interface.
type SSHManager struct {
	AddrFromModelUUID_ func(ctx context.Context, user *openfga.User, modelTag names.ModelTag) (string, error)
	FetchIdentity_     func(ctx context.Context, id string) (*openfga.User, error)
	VerifyPublicKey_   func(ctx context.Context, claimUser string, publicKey []byte) (bool, error)
}

func (j SSHManager) FetchIdentity(ctx context.Context, id string) (*openfga.User, error) {
	if j.FetchIdentity_ == nil {
		return nil, errors.E(errors.CodeNotImplemented)
	}
	return j.FetchIdentity_(ctx, id)
}

func (j SSHManager) AddrFromModelUUID(ctx context.Context, user *openfga.User, modelTag names.ModelTag) (string, error) {
	if j.AddrFromModelUUID_ == nil {
		return "", errors.E(errors.CodeNotImplemented)
	}
	return j.AddrFromModelUUID_(ctx, user, modelTag)
}

func (j SSHManager) VerifyPublicKey(ctx context.Context, claimUser string, publicKey []byte) (bool, error) {
	if j.VerifyPublicKey_ == nil {
		return true, errors.E(errors.CodeNotImplemented)
	}
	return j.VerifyPublicKey_(ctx, claimUser, publicKey)
}
