// Copyright 2024 Canonical.
package jujuapi_test

import (
	"context"

	jc "github.com/juju/testing/checkers"
	gc "gopkg.in/check.v1"

	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/openfga"
	ofganames "github.com/canonical/jimm/v3/internal/openfga/names"
	"github.com/canonical/jimm/v3/pkg/api"
	apiparams "github.com/canonical/jimm/v3/pkg/api/params"
)

/*
 Role facade related tests
*/

func (s *accessControlSuite) TestAddRole(c *gc.C) {
	conn := s.open(c, nil, "alice")
	defer conn.Close()

	client := api.NewClient(conn)
	res, err := client.AddRole(&apiparams.AddRoleRequest{Name: "test-role"})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(res.UUID, gc.Not(gc.Equals), "")

	_, err = client.AddRole(&apiparams.AddRoleRequest{Name: "test-role"})
	c.Assert(err, gc.ErrorMatches, ".*already exists.*")
}

func (s *accessControlSuite) TestGetRole(c *gc.C) {
	conn := s.open(c, nil, "alice")
	defer conn.Close()

	client := api.NewClient(conn)

	created, err := client.AddRole(&apiparams.AddRoleRequest{Name: "test-role"})
	c.Assert(err, jc.ErrorIsNil)

	retrievedUuid, err := client.GetRole(&apiparams.GetRoleRequest{UUID: created.UUID})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(retrievedUuid.Role, gc.DeepEquals, created.Role)

	retrievedName, err := client.GetRole(&apiparams.GetRoleRequest{Name: created.Name})
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(retrievedName.Role, gc.DeepEquals, created.Role)

	_, err = client.GetRole(&apiparams.GetRoleRequest{UUID: "non-existent"})
	c.Assert(err, gc.ErrorMatches, ".*not found.*")

	_, err = client.GetRole(&apiparams.GetRoleRequest{Name: created.Name, UUID: created.UUID})
	c.Assert(err, gc.ErrorMatches, ".*only one of.*")
}

func (s *accessControlSuite) TestRemoveRole(c *gc.C) {
	conn := s.open(c, nil, "alice")
	defer conn.Close()

	client := api.NewClient(conn)

	err := client.RemoveRole(&apiparams.RemoveRoleRequest{
		Name: "test-role",
	})
	c.Assert(err, gc.ErrorMatches, ".*not found.*")

	_, err = client.AddRole(&apiparams.AddRoleRequest{Name: "test-role"})
	c.Assert(err, jc.ErrorIsNil)

	err = client.RemoveRole(&apiparams.RemoveRoleRequest{
		Name: "test-role",
	})
	c.Assert(err, jc.ErrorIsNil)
}

func (s *accessControlSuite) TestRemoveRoleRemovesTuples(c *gc.C) {
	ctx := context.Background()
	db := s.JIMM.Database

	user, role, controller, model, _, _, _, client, closeClient := createTestControllerEnvironment(ctx, c, s)
	defer closeClient()

	_, err := db.AddRole(ctx, "test-group2")
	c.Assert(err, gc.IsNil)

	group2 := &dbmodel.RoleEntry{
		Name: "test-group2",
	}
	err = db.GetRole(ctx, group2)
	c.Assert(err, gc.IsNil)

	tuples := []openfga.Tuple{
		// This tuple should remain as it has no relation to group2
		{
			Object:   ofganames.ConvertTag(user.ResourceTag()),
			Relation: "member",
			Target:   ofganames.ConvertTag(role.ResourceTag()),
		},
		// Below tuples should all be removed as they relate to group2
		{
			Object:   ofganames.ConvertTag(user.ResourceTag()),
			Relation: "member",
			Target:   ofganames.ConvertTag(group2.ResourceTag()),
		},
		{
			Object:   ofganames.ConvertTagWithRelation(group2.ResourceTag(), ofganames.MemberRelation),
			Relation: "member",
			Target:   ofganames.ConvertTag(role.ResourceTag()),
		},
		{
			Object:   ofganames.ConvertTagWithRelation(group2.ResourceTag(), ofganames.MemberRelation),
			Relation: "administrator",
			Target:   ofganames.ConvertTag(controller.ResourceTag()),
		},
		{
			Object:   ofganames.ConvertTagWithRelation(group2.ResourceTag(), ofganames.MemberRelation),
			Relation: "writer",
			Target:   ofganames.ConvertTag(model.ResourceTag()),
		},
	}

	u := user.Tag().String()

	checkAccessTupleController := apiparams.RelationshipTuple{Object: u, Relation: "administrator", TargetObject: "controller-" + controller.UUID}
	checkAccessTupleModel := apiparams.RelationshipTuple{Object: u, Relation: "writer", TargetObject: "model-" + model.UUID.String}

	err = s.JIMM.OpenFGAClient.AddRelation(context.Background(), tuples...)
	c.Assert(err, gc.IsNil)
	// Check user has access to model and controller through group2
	checkResp, err := client.CheckRelation(&apiparams.CheckRelationRequest{Tuple: checkAccessTupleController})
	c.Assert(err, gc.IsNil)
	c.Assert(checkResp.Allowed, gc.Equals, true)
	checkResp, err = client.CheckRelation(&apiparams.CheckRelationRequest{Tuple: checkAccessTupleModel})
	c.Assert(err, gc.IsNil)
	c.Assert(checkResp.Allowed, gc.Equals, true)

	err = client.RemoveRole(&apiparams.RemoveRoleRequest{Name: group2.Name})
	c.Assert(err, gc.IsNil)

	resp, err := client.ListRelationshipTuples(&apiparams.ListRelationshipTuplesRequest{})
	c.Assert(err, gc.IsNil)
	c.Assert(len(resp.Tuples), gc.Equals, 13)

	// Check user access has been revoked.
	checkResp, err = client.CheckRelation(&apiparams.CheckRelationRequest{Tuple: checkAccessTupleController})
	c.Assert(err, gc.IsNil)
	c.Assert(checkResp.Allowed, gc.Equals, false)
	checkResp, err = client.CheckRelation(&apiparams.CheckRelationRequest{Tuple: checkAccessTupleModel})
	c.Assert(err, gc.IsNil)
	c.Assert(checkResp.Allowed, gc.Equals, false)
}

func (s *accessControlSuite) TestRenameRole(c *gc.C) {
	conn := s.open(c, nil, "alice")
	defer conn.Close()

	client := api.NewClient(conn)

	err := client.RenameRole(&apiparams.RenameRoleRequest{
		Name:    "test-role",
		NewName: "renamed-role",
	})
	c.Assert(err, gc.ErrorMatches, ".*not found.*")

	_, err = client.AddRole(&apiparams.AddRoleRequest{Name: "test-role"})
	c.Assert(err, jc.ErrorIsNil)

	err = client.RenameRole(&apiparams.RenameRoleRequest{
		Name:    "test-role",
		NewName: "renamed-role",
	})
	c.Assert(err, jc.ErrorIsNil)
}

func (s *accessControlSuite) TestListRoles(c *gc.C) {
	conn := s.open(c, nil, "alice")
	defer conn.Close()

	client := api.NewClient(conn)

	groupNames := []string{
		"test-group0",
		"test-group1",
		"test-group2",
		"aaaFinalRole",
	}

	for _, name := range groupNames {
		_, err := client.AddRole(&apiparams.AddRoleRequest{Name: name})
		c.Assert(err, jc.ErrorIsNil)
	}
	req := apiparams.ListRolesRequest{Limit: 10, Offset: 0}
	groups, err := client.ListRoles(&req)
	c.Assert(err, jc.ErrorIsNil)
	c.Assert(groups, gc.HasLen, 4)
	// Verify the UUID is not empty.
	c.Assert(groups[0].UUID, gc.Not(gc.Equals), "")
	// groups should be returned in ascending order of name
	c.Assert(groups[0].Name, gc.Equals, "aaaFinalRole")
	c.Assert(groups[1].Name, gc.Equals, "test-group0")
	c.Assert(groups[2].Name, gc.Equals, "test-group1")
	c.Assert(groups[3].Name, gc.Equals, "test-group2")
}
