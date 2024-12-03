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

func (j *JIMM) WatchModelsDying(ctx context.Context) error {
	const op = errors.Op("jimm.WatchModelsDying")

	// Ensure that if the watcher stops because of a database error all
	// the controller connections get closed.
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	adminUser := j.everyoneUser()
	adminUser.JimmAdmin = true
	err := j.ForEachModel(ctx, adminUser, func(m *dbmodel.Model, _ jujuparams.UserAccessPermission) error {
		if m.Life == state.Dying.String() {
			mt := m.ResourceTag()
			// if the model is dying and not found by querying the controller we can assume it is dead.
			// And safely delete the reference from our db.
			j.doModelAdmin(ctx, adminUser, mt, func(m *dbmodel.Model, api API) error {
				if err := api.ModelInfo(ctx, &jujuparams.ModelInfo{}); err != nil {
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
				return nil
			})
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
