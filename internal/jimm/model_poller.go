// Copyright 2024 Canonical.
package jimm

import (
	"context"

	jujuparams "github.com/juju/juju/rpc/params"
	"github.com/juju/juju/state"
	"github.com/juju/zaputil/zapctx"
	"go.uber.org/zap"

	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/errors"
)

func (j *JIMM) PollModelsDying(ctx context.Context) error {
	const op = errors.Op("jimm.WatchModelsDying")

	// Ensure that if the watcher stops because of a database error all
	// the controller connections get closed.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := j.DB().ForEachModel(ctx, func(m *dbmodel.Model) error {
		if m.Life == state.Dying.String() {
			// if the model is dying and not found by querying the controller we can assume it is dead.
			// And safely delete the reference from our db.
			api, err := j.dialModel(ctx, &m.Controller, m.ResourceTag())
			if err != nil {
				return err
			}
			if err := api.ModelInfo(ctx, &jujuparams.ModelInfo{UUID: m.UUID.String}); err != nil {
				// Some versions of juju return unauthorized for models that cannot be found.
				if errors.ErrorCode(err) == errors.CodeNotFound || errors.ErrorCode(err) == errors.CodeUnauthorized {
					if err := j.DB().DeleteModel(ctx, m); err != nil {
						return errors.E(op, err)
					} else {
						return nil
					}
				} else {
					return errors.E(op, err)
				}
			}
		}
		return nil
	})
	if err != nil {
		// Ignore temporary database errors.
		if errors.ErrorCode(err) != errors.CodeDatabaseLocked {
			return errors.E(op, err)
		}
		zapctx.Warn(ctx, "temporary error polling for controllers", zap.Error(err))
	}
	return nil
}
