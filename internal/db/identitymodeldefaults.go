// Copyright 2024 Canonical.

package db

import (
	"context"

	"gorm.io/gorm/clause"

	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/internal/servermon"
)

// SetIdentityModelDefaults sets default model setting values for the identity.
func (d *Database) SetIdentityModelDefaults(ctx context.Context, defaults *dbmodel.IdentityModelDefaults) (err error) {
	const op = errors.Op("db.SetIdentityModelDefaults")

	if err := d.ready(); err != nil {
		return errors.E(op, err)
	}

	durationObserver := servermon.DurationObserver(servermon.DBQueryDurationHistogram, string(op))
	defer durationObserver()
	defer servermon.ErrorCounter(servermon.DBQueryErrorCount, &err, string(op))

	err = d.Transaction(func(d *Database) error {
		db := d.DB.WithContext(ctx)

		dbDefaults := dbmodel.IdentityModelDefaults{
			IdentityName: defaults.IdentityName,
		}
		// we try to get identity defaults, if not found we create them.
		err := d.IdentityModelDefaults(ctx, &dbDefaults)
		if err != nil {
			if errors.ErrorCode(err) == errors.CodeNotFound {
				// if defaults do not exist, we create them
				if err := db.Create(&defaults).Error; err != nil {
					return errors.E(op, dbError(err))
				}
				return nil
			}
			return errors.E(op, err)
		}

		// if they are found, we merge old ones with the new ones.
		for k, v := range defaults.Defaults {
			dbDefaults.Defaults[k] = v
		}
		if err := db.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "identity_name"},
			},
			DoUpdates: clause.AssignmentColumns([]string{"defaults"}),
		}).Create(&dbDefaults).Error; err != nil {
			return errors.E(op, dbError(err))
		}
		return nil
	})
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// UnsetIdentityModelDefaults unset defaults from identity.
func (d *Database) UnsetIdentityModelDefaults(ctx context.Context, defaults *dbmodel.IdentityModelDefaults, keys []string) (err error) {
	const op = errors.Op("db.SetIdentityModelDefaults")

	if err := d.ready(); err != nil {
		return errors.E(op, err)
	}

	durationObserver := servermon.DurationObserver(servermon.DBQueryDurationHistogram, string(op))
	defer durationObserver()
	defer servermon.ErrorCounter(servermon.DBQueryErrorCount, &err, string(op))

	err = d.Transaction(func(d *Database) error {
		db := d.DB.WithContext(ctx)

		dbDefaults := dbmodel.IdentityModelDefaults{
			IdentityName: defaults.IdentityName,
		}
		// we try to get identity defaults, if not found we return an error.
		err := d.IdentityModelDefaults(ctx, &dbDefaults)
		if err != nil {
			return errors.E(op, err)
		}
		// if they are found, we merge old ones with the new deleted ones.
		for _, key := range keys {
			delete(dbDefaults.Defaults, key)
		}
		if err := db.Clauses(clause.OnConflict{
			Columns: []clause.Column{
				{Name: "identity_name"},
			},
			DoUpdates: clause.AssignmentColumns([]string{"defaults"}),
		}).Create(&dbDefaults).Error; err != nil {
			return errors.E(op, dbError(err))
		}
		return nil
	})
	if err != nil {
		return errors.E(op, err)
	}
	return nil
}

// IdentityModelDefaults fetches identities defaults.
func (d *Database) IdentityModelDefaults(ctx context.Context, defaults *dbmodel.IdentityModelDefaults) (err error) {
	const op = errors.Op("db.IdentityModelDefaults")

	if err := d.ready(); err != nil {
		return errors.E(op, err)
	}

	durationObserver := servermon.DurationObserver(servermon.DBQueryDurationHistogram, string(op))
	defer durationObserver()
	defer servermon.ErrorCounter(servermon.DBQueryErrorCount, &err, string(op))

	db := d.DB.WithContext(ctx)

	db = db.Where("identity_name = ?", defaults.IdentityName)

	result := db.Preload("Identity").First(&defaults)
	if result.Error != nil {
		err := dbError(result.Error)
		if errors.ErrorCode(err) == errors.CodeNotFound {
			return errors.E(op, errors.CodeNotFound, "identitymodeldefaults not found", err)
		}
		return errors.E(op, err)
	}
	return nil
}
