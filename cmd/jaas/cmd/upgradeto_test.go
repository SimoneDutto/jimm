// Copyright 2025 Canonical.

// Note that this file is not an integration test
// because of limitations with the JujuConnSuite
// so it is placed under the cmd package.

package cmd

import (
	"context"
	"errors"
	"strings"

	"github.com/juju/cmd/v3"
	"github.com/juju/cmd/v3/cmdtesting"
	"github.com/juju/gnuflag"
	jjclient "github.com/juju/juju/jujuclient"
	"go.uber.org/mock/gomock"
	gc "gopkg.in/check.v1"

	"github.com/canonical/jimm/v3/cmd/jaas/cmd/mocks"
	apiparams "github.com/canonical/jimm/v3/pkg/api/params"
)

// upgradeToSuite is a test suite for the upgrade-to command.
type upgradeToSuite struct {
	jimmClient *mocks.MockJIMMAPI
	writer     *mocks.MockWriter
	store      *mocks.MockClientStore
}

var _ = gc.Suite(&upgradeToSuite{})

func (s *upgradeToSuite) SetupMocks(c *gc.C) *gomock.Controller {
	ctrl := gomock.NewController(c)
	s.jimmClient = mocks.NewMockJIMMAPI(ctrl)
	s.writer = mocks.NewMockWriter(ctrl)
	s.store = mocks.NewMockClientStore(ctrl)

	return ctrl
}

func (s *upgradeToSuite) TestUpgradeTo(c *gc.C) {
	defer s.SetupMocks(c).Finish()

	testModelUUID := "93608db4-f1cb-4da5-9926-8233981aef0a"
	testTargetVersion := "3.5.0"
	testLogs := "Upgrade initiated successfully\nController is being upgraded to version 3.5.0"

	upgradeToParams := &apiparams.UpgradeToRequest{
		TargetVersion: testTargetVersion,
		ModelUUID:     testModelUUID,
	}

	s.jimmClient.EXPECT().UpgradeTo(upgradeToParams).Return(apiparams.UpgradeToResponse{
		Logs: testLogs,
	}, nil)
	s.jimmClient.EXPECT().Close().Return(nil)

	s.writer.EXPECT().Write(gomock.Any()).DoAndReturn(func(b []byte) (int, error) {
		c.Check(strings.Contains(string(b), "Upgrade initiated successfully"), gc.Equals, true)
		return len(b), nil
	})

	upgradeToCmd := &upgradeToCommand{
		jimmAPIFunc: func() (JIMMAPI, error) {
			return s.jimmClient, nil
		},
		store: s.store,
	}
	f := gnuflag.NewFlagSet("test", gnuflag.ExitOnError)
	f.SetOutput(s.writer)
	upgradeToCmd.SetFlags(f)

	// Set args after setting flags to avoid resetting them.
	upgradeToCmd.version = testTargetVersion
	upgradeToCmd.modelUUID = testModelUUID

	ctx := &cmd.Context{
		Context: context.Background(),
		Stdout:  s.writer,
	}
	err := upgradeToCmd.Run(ctx)
	c.Assert(err, gc.IsNil)
}

func (s *upgradeToSuite) TestUpgradeToWithError(c *gc.C) {
	defer s.SetupMocks(c).Finish()

	testModelUUID := "93608db4-f1cb-4da5-9926-8233981aef0a"
	testTargetVersion := "3.5.0"

	upgradeToParams := &apiparams.UpgradeToRequest{
		TargetVersion: testTargetVersion,
		ModelUUID:     testModelUUID,
	}
	errorToReturn := errors.New("failed to initiate upgrade")
	s.jimmClient.EXPECT().UpgradeTo(upgradeToParams).Return(apiparams.UpgradeToResponse{}, errorToReturn)
	s.jimmClient.EXPECT().Close().Return(nil)

	upgradeToCmd := &upgradeToCommand{
		jimmAPIFunc: func() (JIMMAPI, error) {
			return s.jimmClient, nil
		},
		store: s.store,
	}
	f := gnuflag.NewFlagSet("test", gnuflag.ExitOnError)
	f.SetOutput(s.writer)
	upgradeToCmd.SetFlags(f)

	// Set args after setting flags to avoid resetting them.
	upgradeToCmd.version = testTargetVersion
	upgradeToCmd.modelUUID = testModelUUID

	ctx := &cmd.Context{
		Context: context.Background(),
		Stdout:  s.writer,
	}
	err := upgradeToCmd.Run(ctx)
	c.Assert(err, gc.ErrorMatches, ".*failed to initiate upgrade.*")
}

func (s *upgradeToSuite) TestCommandsFailsWithMissingArgs(c *gc.C) {
	_, err := cmdtesting.RunCommand(c, NewUpgradeToCommandForTesting(jjclient.NewMemStore(), nil))
	c.Assert(err, gc.ErrorMatches, "missing required arguments: model UUID and version")
}

func (s *upgradeToSuite) TestCommandsFailsWithOnlyOneArg(c *gc.C) {
	_, err := cmdtesting.RunCommand(c, NewUpgradeToCommandForTesting(jjclient.NewMemStore(), nil), "93608db4-f1cb-4da5-9926-8233981aef0a")
	c.Assert(err, gc.ErrorMatches, "missing required arguments: model UUID and version")
}

func (s *upgradeToSuite) TestCommandWithPositionalArgs(c *gc.C) {
	defer s.SetupMocks(c).Finish()

	testModelUUID := "93608db4-f1cb-4da5-9926-8233981aef0a"
	testTargetVersion := "3.5.0"

	upgradeToParams := &apiparams.UpgradeToRequest{
		TargetVersion: testTargetVersion,
		ModelUUID:     testModelUUID,
	}

	s.jimmClient.EXPECT().UpgradeTo(upgradeToParams).Return(apiparams.UpgradeToResponse{
		Logs: "success",
	}, nil)
	s.jimmClient.EXPECT().Close().Return(nil)

	s.writer.EXPECT().Write(gomock.Any()).Return(0, nil)

	upgradeToCmd := &upgradeToCommand{
		jimmAPIFunc: func() (JIMMAPI, error) {
			return s.jimmClient, nil
		},
		store: s.store,
	}
	f := gnuflag.NewFlagSet("test", gnuflag.ExitOnError)
	f.SetOutput(s.writer)
	upgradeToCmd.SetFlags(f)

	err := upgradeToCmd.Init([]string{testModelUUID, testTargetVersion})
	c.Assert(err, gc.IsNil)

	ctx := &cmd.Context{
		Context: context.Background(),
		Stdout:  s.writer,
	}
	err = upgradeToCmd.Run(ctx)
	c.Assert(err, gc.IsNil)
}
