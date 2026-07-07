// Copyright 2026 Canonical.

package testing

import (
	"context"
	"strings"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/juju/juju/api/base"
	"github.com/juju/juju/api/client/applicationoffers"
	"github.com/juju/juju/core/crossmodel"
	"github.com/juju/names/v5"

	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/internal/jimm/juju"
	"github.com/canonical/jimm/v3/internal/testutils/jimmtest"
)

// controllerModel returns the tracked backing controller model for the named
// controller, failing the test if it is not present.
func controllerModel(c *qt.C, s jimmtest.JimmWithControllers, controllerName string) *dbmodel.Model {
	m := &dbmodel.Model{
		OwnerIdentityName: controllerName + dbmodel.ControllerModelOwnerSuffix,
		Name:              dbmodel.ControllerModelName,
	}
	err := s.JIMM.Database.GetModel(context.Background(), m)
	c.Assert(err, qt.IsNil, qt.Commentf("controller model for %q not found", controllerName))
	return m
}

// TestAddControllerTracksControllerModel checks that AddController (invoked by
// the test setup) tracks the backing controller's model as a first-class JAAS
// model under the "<controller-name>@controller" owner, without needing a poll,
// and that it is only visible to controller admins.
func TestAddControllerTracksControllerModel(t *testing.T) {
	c := qt.New(t)
	s := jimmtest.SetupJimmWithControllers(c)
	ctx := context.Background()

	// Note: no PollModels call here. AddController alone must have tracked the
	// controller model.
	controllerName, _ := s.GetOneControllerConfig(c)
	m := controllerModel(c, s, controllerName)

	c.Check(m.Name, qt.Equals, dbmodel.ControllerModelName)
	c.Check(m.IsControllerModel, qt.IsTrue)
	c.Check(m.OwnerIdentityName, qt.Equals, controllerName+dbmodel.ControllerModelOwnerSuffix)
	// By default Juju owns the controller model as "admin"; JAAS reads this from
	// the controller when tracking the model.
	c.Check(m.ControllerFacingOwner(), qt.Equals, "admin")

	// A JIMM/controller admin (alice) can see the controller model.
	summaries, err := s.JIMM.JujuManager.ListModelSummaries(ctx, s.AdminUser, "")
	c.Assert(err, qt.IsNil)
	c.Check(containsModelUUID(summaries, m.UUID.String), qt.IsTrue)

	// A non-admin user (bob) must NOT see the controller model.
	bobIdentity, err := dbmodel.NewIdentity("bob@canonical.com")
	c.Assert(err, qt.IsNil)
	c.Assert(s.JIMM.Database.GetIdentity(ctx, bobIdentity), qt.IsNil)
	bob := s.NewUser(bobIdentity)

	bobSummaries, err := s.JIMM.JujuManager.ListModelSummaries(ctx, bob, "")
	c.Assert(err, qt.IsNil)
	c.Check(containsModelUUID(bobSummaries, m.UUID.String), qt.IsFalse)
}

// TestPollModelsBackfillsControllerModel checks that if the controller model is
// missing (e.g. a controller added before tracking existed), PollModels
// re-creates it. It is also idempotent.
func TestPollModelsBackfillsControllerModel(t *testing.T) {
	c := qt.New(t)
	s := jimmtest.SetupJimmWithControllers(c)
	ctx := context.Background()

	controllerName, _ := s.GetOneControllerConfig(c)
	m := controllerModel(c, s, controllerName)

	// Delete the tracked controller model to simulate an untracked controller.
	err := s.JIMM.OpenFGAClient.RemoveControllerModel(ctx, m.Controller.ResourceTag(), m.ResourceTag())
	c.Assert(err, qt.IsNil)
	err = s.JIMM.Database.DeleteModel(ctx, m)
	c.Assert(err, qt.IsNil)

	// Confirm it is gone.
	gone := &dbmodel.Model{UUID: m.UUID}
	err = s.JIMM.Database.GetModel(ctx, gone)
	c.Assert(errors.ErrorCode(err), qt.Equals, errors.CodeNotFound)

	// PollModels must re-create it.
	err = s.JIMM.JujuManager.PollModels(ctx)
	c.Assert(err, qt.IsNil)
	recreated := controllerModel(c, s, controllerName)
	c.Check(recreated.UUID.String, qt.Equals, m.UUID.String)
	c.Check(recreated.IsControllerModel, qt.IsTrue)

	// Polling again is idempotent and does not fail or duplicate.
	err = s.JIMM.JujuManager.PollModels(ctx)
	c.Assert(err, qt.IsNil)
	_ = controllerModel(c, s, controllerName)
}

func containsModelUUID(summaries []base.UserModelSummary, uuid string) bool {
	for _, s := range summaries {
		if s.UUID == uuid {
			return true
		}
	}
	return false
}

// TestImportModelWithOwnerCreateOffer imports a model into JAAS under a new
// owner (so the JAAS owner differs from the backing controller owner) and then
// creates an application offer for it.
//
// UX note: JAAS does not rewrite offer URLs into the JAAS namespace; offer URLs
// are pass-through and always reflect the backing controller (the model's
// original/controller-facing owner). So after importing under a new owner, the
// model is listed under the new owner, but the offer URL keeps the ORIGINAL
// owner. This test documents that behaviour: the offer is created successfully
// (using the controller-facing owner) and its URL contains the original owner,
// not the new JAAS owner.
func TestImportModelWithOwnerCreateOffer(t *testing.T) {
	c := qt.New(t)
	s := jimmtest.SetupJimmWithControllers(c)
	ctx := context.Background()

	// Create a model owned by charlie and deploy an application to offer.
	model := s.CreateModelForCharlie(c)
	controllerName := model.Controller.Name
	originalOwner := model.OwnerIdentityName

	s.DeployApplication(c, s.AdminUser, model.Tag(), jimmtest.DeployApplicationParams{
		App:   "test-app",
		Charm: "juju-qa-dummy-sink",
	})

	// Remove the model from JIMM (but keep it on the backing controller) so we
	// can re-import it under a different owner.
	err := s.JIMM.OpenFGAClient.RemoveControllerModel(ctx, model.Controller.ResourceTag(), model.ResourceTag())
	c.Assert(err, qt.IsNil)
	err = s.JIMM.Database.DeleteModel(ctx, model)
	c.Assert(err, qt.IsNil)

	// Re-import the model assigning bob as the new owner.
	newOwner := "bob@canonical.com"
	err = s.JIMM.JujuManager.ImportModel(ctx, s.AdminUser, controllerName, model.ResourceTag(), newOwner)
	c.Assert(err, qt.IsNil)
	c.Cleanup(func() {
		s.DestroyModelAndDeleteFromDatabase(c, model.ResourceTag())
	})

	imported := &dbmodel.Model{}
	imported.SetTag(model.ResourceTag())
	c.Assert(s.JIMM.Database.GetModel(ctx, imported), qt.IsNil)

	// In JAAS the model is now owned by the new owner, but the backing
	// controller still knows it under its original owner.
	c.Check(imported.OwnerIdentityName, qt.Equals, newOwner)
	c.Check(imported.ControllerFacingOwner(), qt.Equals, originalOwner)

	// Creating an offer succeeds even though the owners differ, because the
	// offer URL is built with the controller-facing (original) owner. Using the
	// new JAAS owner would make the backing controller report "not found".
	err = s.JIMM.JujuManager.Offer(ctx, s.AdminUser, juju.AddApplicationOfferParams{
		ModelTag:        model.ResourceTag(),
		OwnerTag:        names.NewUserTag(newOwner),
		OfferName:       "test-offer",
		ApplicationName: "test-app",
		Endpoints:       map[string]string{"source": "source"},
	})
	c.Assert(err, qt.IsNil)

	// Listing offers works using the JAAS owner (the owner the user sees). JIMM
	// resolves the model by its JAAS owner and translates to the controller-facing
	// owner when querying the backing controller.
	offers, err := s.JIMM.JujuManager.ListApplicationOffers(ctx, s.AdminUser, crossmodel.ApplicationOfferFilter{
		OwnerName: newOwner,
		ModelName: imported.Name,
	})
	c.Assert(err, qt.IsNil)
	c.Assert(offers, qt.HasLen, 1)

	// Pass-through UX: the returned offer URL keeps the model's ORIGINAL owner
	// and does NOT use the new JAAS owner. This is what a user consuming the
	// offer will use.
	c.Check(strings.Contains(offers[0].OfferURL, originalOwner+"/"+imported.Name), qt.IsTrue,
		qt.Commentf("offer URL should keep original owner %q; got %q", originalOwner, offers[0].OfferURL))
	c.Check(strings.Contains(offers[0].OfferURL, newOwner), qt.IsFalse,
		qt.Commentf("offer URL should NOT contain the new JAAS owner %q; got %q", newOwner, offers[0].OfferURL))

	// Consuming the offer using the JAAS owner (the owner the user sees) must
	// work. JIMM translates the JAAS URL to the controller-facing URL when
	// looking up the offer and forwarding to the backing controller.
	conn := s.Open(c, nil, "alice@canonical.com", nil)
	defer conn.Close()
	offerClient := applicationoffers.NewClient(conn)

	jaasURL := crossmodel.OfferURL{
		User:            newOwner,
		ModelName:       imported.Name,
		ApplicationName: "test-offer",
	}
	consumeDetails, err := offerClient.GetConsumeDetails(jaasURL.Path())
	c.Assert(err, qt.IsNil)
	c.Check(consumeDetails.Offer.OfferName, qt.Equals, "test-offer")
	c.Check(consumeDetails.Offer.OfferURL, qt.Not(qt.Equals), "")
}
