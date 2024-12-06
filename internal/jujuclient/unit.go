package jujuclient

import (
	"context"

	"github.com/canonical/jimm/v3/internal/errors"
	jujuerrors "github.com/juju/errors"
	jujuparams "github.com/juju/juju/rpc/params"
	"github.com/juju/names/v5"
)

func (c Connection) UnitsInfo(ctx context.Context, units []names.UnitTag) (jujuparams.UnitInfoResults, error) {
	const op = errors.Op("jujuclient.UnitsInfo")

	all := make([]jujuparams.Entity, len(units))
	for i, one := range units {
		all[i] = jujuparams.Entity{Tag: one.String()}
	}
	args := jujuparams.Entities{Entities: all}
	var resp jujuparams.UnitInfoResults
	err := c.Call(ctx, "Application", 20, "", "UnitsInfo", &args, &resp)
	if err != nil {
		return jujuparams.UnitInfoResults{}, errors.E(op, jujuerrors.Cause(err))
	}

	return resp, nil
}
