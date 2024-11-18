// Copyright 2024 Canonical.

package db_test

import (
	"context"
	"sort"

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

func (s *dbSuite) TestDatabase_GetRole(c *qt.C) {
	uuid1 := uuid.NewString()
	c.Patch(db.NewUUID, func() string {
		return uuid1
	})

	err := s.Database.GetRole(context.Background(), &dbmodel.RoleEntry{
		Name: "test-role",
	})
	c.Check(errors.ErrorCode(err), qt.Equals, errors.CodeUpgradeInProgress)

	err = s.Database.Migrate(context.Background(), false)
	c.Assert(err, qt.IsNil)

	role := &dbmodel.RoleEntry{
		Name: "test-role",
	}
	err = s.Database.GetRole(context.Background(), role)
	c.Check(errors.ErrorCode(err), qt.Equals, errors.CodeNotFound)

	roleEntry, err := s.Database.AddRole(context.TODO(), "test-role")
	c.Assert(err, qt.IsNil)
	c.Assert(roleEntry.UUID, qt.Equals, uuid1)

	err = s.Database.GetRole(context.Background(), role)
	c.Check(err, qt.IsNil)
	c.Assert(role.ID, qt.Equals, uint(1))
	c.Assert(role.Name, qt.Equals, "test-role")
	c.Assert(role.UUID, qt.Equals, uuid1)

	uuid2 := uuid.NewString()
	c.Patch(db.NewUUID, func() string {
		return uuid2
	})

	roleEntry, err = s.Database.AddRole(context.Background(), "test-role1")
	c.Assert(err, qt.IsNil)
	c.Assert(roleEntry.UUID, qt.Equals, uuid2)

	role = &dbmodel.RoleEntry{
		Name: "test-role1",
	}

	err = s.Database.GetRole(context.Background(), role)
	c.Check(err, qt.IsNil)
	c.Assert(role.ID, qt.Equals, uint(2))
	c.Assert(role.Name, qt.Equals, "test-role1")
	c.Assert(role.UUID, qt.Equals, uuid2)
}

func (s *dbSuite) TestDatabase_UpdateRole(c *qt.C) {
	err := s.Database.UpdateRole(context.Background(), &dbmodel.RoleEntry{Name: "test-role"})
	c.Check(errors.ErrorCode(err), qt.Equals, errors.CodeNotFound)

	err = s.Database.Migrate(context.Background(), false)
	c.Assert(err, qt.IsNil)

	ge := &dbmodel.RoleEntry{
		Name: "test-role",
	}

	err = s.Database.UpdateRole(context.Background(), ge)
	c.Check(errors.ErrorCode(err), qt.Equals, errors.CodeNotFound)

	_, err = s.Database.AddRole(context.Background(), "test-role")
	c.Assert(err, qt.IsNil)

	ge1 := &dbmodel.RoleEntry{
		Name: "test-role",
	}
	err = s.Database.GetRole(context.Background(), ge1)
	c.Assert(err, qt.IsNil)

	ge1.Name = "renamed-role"
	err = s.Database.UpdateRole(context.Background(), ge1)
	c.Check(err, qt.IsNil)

	ge2 := &dbmodel.RoleEntry{
		Name: "renamed-role",
	}
	err = s.Database.GetRole(context.Background(), ge2)
	c.Check(err, qt.IsNil)
	c.Assert(ge2, qt.DeepEquals, ge1)
}

func (s *dbSuite) TestDatabase_RemoveRole(c *qt.C) {
	err := s.Database.RemoveRole(context.Background(), &dbmodel.RoleEntry{Name: "test-role"})
	c.Check(errors.ErrorCode(err), qt.Equals, errors.CodeNotFound)

	err = s.Database.Migrate(context.Background(), false)
	c.Assert(err, qt.IsNil)

	ge := &dbmodel.RoleEntry{
		Name: "test-role",
	}
	err = s.Database.RemoveRole(context.Background(), ge)
	c.Check(errors.ErrorCode(err), qt.Equals, errors.CodeNotFound)

	roleEntry, err := s.Database.AddRole(context.Background(), ge.Name)
	c.Assert(err, qt.IsNil)

	ge1 := &dbmodel.RoleEntry{
		Name: "test-role",
	}
	err = s.Database.GetRole(context.Background(), ge1)
	c.Assert(err, qt.IsNil)
	c.Assert(roleEntry.UUID, qt.Equals, ge1.UUID)

	err = s.Database.RemoveRole(context.Background(), ge1)
	c.Check(err, qt.IsNil)

	err = s.Database.GetRole(context.Background(), ge1)
	c.Check(errors.ErrorCode(err), qt.Equals, errors.CodeNotFound)
}

func (s *dbSuite) TestDatabase_ForEachRole(c *qt.C) {
	ctx := context.Background()

	err := s.Database.Migrate(context.Background(), false)
	c.Assert(err, qt.IsNil)

	_, err = s.Database.AddRole(ctx, "role-1")
	c.Assert(err, qt.IsNil)

	_, err = s.Database.AddRole(ctx, "role-2")
	c.Assert(err, qt.IsNil)

	_, err = s.Database.AddRole(ctx, "role-3")
	c.Assert(err, qt.IsNil)

	var roleNames []string

	err = s.Database.ForEachRole(ctx, func(re *dbmodel.RoleEntry) error {
		roleNames = append(roleNames, re.Name)
		return nil
	})
	c.Assert(err, qt.IsNil)

	sort.Slice(roleNames, func(i, j int) bool {
		return i < j
	})

	c.Assert(roleNames, qt.DeepEquals, []string{
		"role-1",
		"role-2",
		"role-3",
	})
}
