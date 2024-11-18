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

// GetRole populates the provided *dbmodel.RoleEntry based on ID, name or UUID.
func (d *Database) GetRole(ctx context.Context, role *dbmodel.RoleEntry) (err error) {
	const op = errors.Op("db.GetRole")
	if err := d.ready(); err != nil {
		return errors.E(op, err)
	}

	durationObserver := servermon.DurationObserver(servermon.DBQueryDurationHistogram, string(op))
	defer durationObserver()
	defer servermon.ErrorCounter(servermon.DBQueryErrorCount, &err, string(op))

	db := d.DB.WithContext(ctx)
	if role.ID != 0 {
		db = db.Where("id = ?", role.ID)
	}
	if role.UUID != "" {
		db = db.Where("uuid = ?", role.UUID)
	}
	if role.Name != "" {
		db = db.Where("name = ?", role.Name)
	}
	if err := db.First(&role).Error; err != nil {
		return errors.E(op, dbError(err))
	}
	return nil
}

// UpdateRole updates the role identified by its ID or UUID.
func (d *Database) UpdateRole(ctx context.Context, role *dbmodel.RoleEntry) (err error) {
	const op = errors.Op("db.UpdateRole")

	if role.ID == 0 {
		return errors.E(errors.CodeNotFound)
	}
	if role.UUID == "" {
		return errors.E("role uuid not specified", errors.CodeNotFound)
	}

	if err := d.ready(); err != nil {
		return errors.E(op, err)
	}

	durationObserver := servermon.DurationObserver(servermon.DBQueryDurationHistogram, string(op))
	defer durationObserver()
	defer servermon.ErrorCounter(servermon.DBQueryErrorCount, &err, string(op))

	if err := d.DB.WithContext(ctx).Save(role).Error; err != nil {
		return errors.E(op, dbError(err))
	}
	return nil
}

// RemoveRole removes the role identified by its ID or UUID.
func (d *Database) RemoveRole(ctx context.Context, role *dbmodel.RoleEntry) (err error) {
	const op = errors.Op("db.RemoveRole")

	if role.ID == 0 {
		return errors.E(errors.CodeNotFound)
	}
	if role.UUID == "" {
		return errors.E(errors.CodeNotFound)
	}

	if err := d.ready(); err != nil {
		return errors.E(op, err)
	}

	durationObserver := servermon.DurationObserver(servermon.DBQueryDurationHistogram, string(op))
	defer durationObserver()
	defer servermon.ErrorCounter(servermon.DBQueryErrorCount, &err, string(op))

	if err := d.DB.WithContext(ctx).Delete(role).Error; err != nil {
		return errors.E(op, dbError(err))
	}
	return nil
}

// ForEachRole iterates through all role entries applying the provided callback function.
func (d *Database) ForEachRole(ctx context.Context, f func(*dbmodel.RoleEntry) error) (err error) {
	const op = errors.Op("db.ForEachRole")
	if err := d.ready(); err != nil {
		return errors.E(op, err)
	}

	durationObserver := servermon.DurationObserver(servermon.DBQueryDurationHistogram, string(op))
	defer durationObserver()
	defer servermon.ErrorCounter(servermon.DBQueryErrorCount, &err, string(op))

	db := d.DB.WithContext(ctx).Model(&dbmodel.RoleEntry{})

	rows, err := db.Rows()
	if err != nil {
		return errors.E(op, err)
	}
	defer rows.Close()
	for rows.Next() {
		var ale dbmodel.RoleEntry
		if err := db.ScanRows(rows, &ale); err != nil {
			return errors.E(op, err)
		}
		if err := f(&ale); err != nil {
			return err
		}
	}
	if rows.Err() != nil {
		return errors.E(op, rows.Err())
	}
	return nil
}
