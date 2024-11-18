// Copyright 2024 Canonical.

package db

import (
	"context"

	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/internal/servermon"
)

// AddRole adds a new role.
func (d *Database) AddRole(ctx context.Context, name string) (re *dbmodel.RoleEntry, err error) {
	const op = errors.Op("db.AddRole")
	if err := d.ready(); err != nil {
		return nil, errors.E(op, err)
	}

	durationObserver := servermon.DurationObserver(servermon.DBQueryDurationHistogram, string(op))
	defer durationObserver()
	defer servermon.ErrorCounter(servermon.DBQueryErrorCount, &err, string(op))

	re = &dbmodel.RoleEntry{
		Name: name,
		UUID: newUUID(),
	}

	if err := d.DB.WithContext(ctx).Create(re).Error; err != nil {
		return nil, errors.E(op, dbError(err))
	}
	return re, nil
}
