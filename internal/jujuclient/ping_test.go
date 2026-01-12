// Copyright 2025 Canonical.

package jujuclient_test

import (
	"context"

	"github.com/juju/names/v5"
	gc "gopkg.in/check.v1"

	"github.com/canonical/jimm/v3/internal/dbmodel"
)

type pingSuite struct {
	jujuclientSuite
}

var _ = gc.Suite(&pingSuite{})

func (s *pingSuite) TestPing(c *gc.C) {
	ctx := context.Background()

	controllerName, conf := s.GetOneControllerConfig(c)
	info := conf.ToAPIInfo()
	ctl := dbmodel.Controller{
		UUID:          conf.UUID,
		Name:          controllerName,
		CACertificate: info.CACert,
		PublicAddress: info.Addrs[0],
	}
	api, err := s.Dialer.Dial(ctx, &ctl, names.ModelTag{}, nil, nil)
	c.Assert(err, gc.Equals, nil)
	defer api.Close()

	err = api.Ping(ctx)
	c.Assert(err, gc.Equals, nil)
}
