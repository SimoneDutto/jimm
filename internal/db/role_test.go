// Copyright 2024 Canonical.

package db_test

import (
	"context"

	qt "github.com/frankban/quicktest"
	"github.com/google/uuid"

	"github.com/canonical/jimm/v3/internal/db"
	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/errors"
)

func (s *dbSuite) TestDatabase_AddRole(c *qt.C) {
	ctx := context.Background()

	uuid := uuid.NewString()
	c.Patch(db.NewUUID, func() string {
		return uuid
	})

	_, err := s.Database.AddRole(ctx, "test-role")
	c.Check(errors.ErrorCode(err), qt.Equals, errors.CodeUpgradeInProgress)

	err = s.Database.Migrate(context.Background(), false)
	c.Assert(err, qt.IsNil)

	roleEntry, err := s.Database.AddRole(ctx, "test-role")
	c.Assert(err, qt.IsNil)
	c.Assert(roleEntry.UUID, qt.Not(qt.Equals), "")

	_, err = s.Database.AddRole(ctx, "test-role")
	c.Assert(errors.ErrorCode(err), qt.Equals, errors.CodeAlreadyExists)

	re := dbmodel.RoleEntry{
		Name: "test-role",
	}
	tx := s.Database.DB.First(&re)
	c.Assert(tx.Error, qt.IsNil)
	c.Assert(re.ID, qt.Equals, uint(1))
	c.Assert(re.Name, qt.Equals, "test-role")
	c.Assert(re.UUID, qt.Equals, uuid)
}
