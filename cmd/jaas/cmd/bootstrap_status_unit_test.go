// Copyright 2025 Canonical.

package cmd

import (
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/frankban/quicktest/qtsuite"
	"github.com/juju/cmd/v3"
	"go.uber.org/mock/gomock"

	"github.com/canonical/jimm/v3/cmd/jaas/cmd/mocks"
	"github.com/canonical/jimm/v3/pkg/api/params"
)

type bootstrapStatusSuite struct {
	client *mocks.MockJIMMClient
	writer *mocks.MockWriter
}

func (s *bootstrapStatusSuite) SetupMocks(c *qt.C) *gomock.Controller {
	ctrl := gomock.NewController(c)
	s.client = mocks.NewMockJIMMClient(ctrl)
	s.writer = mocks.NewMockWriter(ctrl)

	return ctrl
}

func (s *bootstrapStatusSuite) TestBootstrapStatus(c *qt.C) {
	ctrl := s.SetupMocks(c)
	defer ctrl.Finish()

	s.client.EXPECT().BootstrapStatus(gomock.Any()).Return(params.BootstrapStatusResponse{
		Status: params.StatusSuccessful,
	}, nil)
	s.writer.EXPECT().Write([]byte("Bootstrap job completed successfully.\n"))

	command := &bootstrapStatusCommand{
		client: s.client,
		jobId:  "test-job-id",
	}
	ctx := &cmd.Context{
		Context: c.Context(),
		Stdout:  s.writer,
	}
	err := command.Run(ctx)
	c.Assert(err, qt.IsNil)
}

func TestBootstrapStatusCmd(t *testing.T) {
	qtsuite.Run(qt.New(t), &bootstrapStatusSuite{})
}
