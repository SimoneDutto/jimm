// Copyright 2024 Canonical.

package jimm

import (
	"context"
	"fmt"

	jujuparams "github.com/juju/juju/rpc/params"
	"github.com/juju/juju/state"
	"github.com/juju/zaputil/zapctx"
	"go.uber.org/zap"

	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/errors"
)

// CleanupModelsDying loops over dying models, contacting the respective controller.
// And deleting the model from our database if the error is `NotFound` which means the model was successfully deleted.
func (j *JIMM) CleanupModelsDying(ctx context.Context) error {
	const op = errors.Op("jimm.CleanupModelsDying")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := j.DB().ForEachModel(ctx, func(m *dbmodel.Model) error {
		if m.Life == state.Dying.String() {
			// if the model is dying and not found by querying the controller we can assume it is dead.
			// And safely delete the reference from our db.
			api, err := j.dialModel(ctx, &m.Controller, m.ResourceTag())
			if err != nil {
				zapctx.Error(ctx, fmt.Sprintf("Cannot dial model %s: %s\n", m.UUID.String, err))
				return nil
			}
			if err := api.ModelInfo(ctx, &jujuparams.ModelInfo{UUID: m.UUID.String}); err != nil {
				// Some versions of juju return unauthorized for models that cannot be found.
				if errors.ErrorCode(err) == errors.CodeNotFound || errors.ErrorCode(err) == errors.CodeUnauthorized {
					if err := j.DB().DeleteModel(ctx, m); err != nil {
						zapctx.Error(ctx, fmt.Sprintf("Cannot delete model %s: %s\n", m.UUID.String, err))
					} else {
						return nil
					}
				} else {
					zapctx.Error(ctx, fmt.Sprintf("Cannot get ModelInfo for model %s: %s\n", m.UUID.String, err))
					return nil
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
