// Copyright 2024 Canonical.

package jimm_test

import (
	"context"
	"database/sql"
	"testing"
	"time"

	qt "github.com/frankban/quicktest"
	"github.com/google/uuid"
	jujuparams "github.com/juju/juju/rpc/params"
	"github.com/juju/juju/state"
	"github.com/juju/names/v5"

	"github.com/canonical/jimm/v3/internal/db"
	"github.com/canonical/jimm/v3/internal/dbmodel"
	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/internal/jimm"
	"github.com/canonical/jimm/v3/internal/testutils/jimmtest"
)

const modelPollerTestEnv = `clouds:
- name: test-cloud
  type: test-provider
  regions:
  - name: test-cloud-region
cloud-credentials:
- owner: alice@canonical.com
  name: cred-1
  cloud: test-cloud
controllers:
- name: controller-1
  uuid: 00000001-0000-0000-0000-000000000001
  cloud: test-cloud
  region: test-cloud-region
models:
- name: model-1
  uuid: 00000002-0000-0000-0000-000000000001
  controller: controller-1
  cloud: test-cloud
  region: test-cloud-region
  cloud-credential: cred-1
  owner: alice@canonical.com
  life: alive
  users:
  - user: alice@canonical.com
    access: admin
  - user: bob@canonical.com
    access: admin
- name: model-2
  uuid: 00000002-0000-0000-0000-000000000002
  controller: controller-1
  cloud: test-cloud
  region: test-cloud-region
  cloud-credential: cred-1
  owner: alice@canonical.com
  life: alive
  users:
  - user: alice@canonical.com
    access: admin
  - user: bob@canonical.com
    access: write
- name: model-3
  uuid: 00000002-0000-0000-0000-000000000003
  controller: controller-1
  cloud: test-cloud
  region: test-cloud-region
  cloud-credential: cred-1
  owner: alice@canonical.com
  life: alive
  users:
  - user: alice@canonical.com
    access: admin
  - user: bob@canonical.com
    access: write
users:
- username: alice@canonical.com
  controller-access: superuser
`

func TestPollModelsDying(t *testing.T) {
	// init
	c := qt.New(t)
	ctx := context.Background()

	client, _, _, err := jimmtest.SetupTestOFGAClient(c.Name())
	c.Assert(err, qt.IsNil)
	j := &jimm.JIMM{
		UUID:          uuid.NewString(),
		OpenFGAClient: client,
		Database: db.Database{
			DB: jimmtest.PostgresDB(c, nil),
		},
	}
	err = j.Database.Migrate(ctx, false)
	c.Assert(err, qt.IsNil)
	jimmAdmin, err := j.GetUser(ctx, "alice@canonical.com")
	c.Assert(err, qt.IsNil)

	env := jimmtest.ParseEnvironment(c, modelPollerTestEnv)
	env.PopulateDBAndPermissions(c, j.ResourceTag(), j.Database, client)

	j.Dialer = &jimmtest.Dialer{
		API: &jimmtest.API{
			ModelInfo_: func(ctx context.Context, mi *jujuparams.ModelInfo) error {
				switch mi.UUID {
				case env.Models[0].UUID:
					return errors.E(errors.CodeNotFound)
				case env.Models[1].UUID:
					return nil
				default:
					return errors.E("new error")
				}
			},
			DestroyModel_: func(ctx context.Context, mt names.ModelTag, b1, b2 *bool, d1, d2 *time.Duration) error {
				return nil
			},
		},
	}
	err = j.DestroyModel(ctx, jimmAdmin, names.NewModelTag(env.Models[0].UUID), nil, nil, nil, nil)
	c.Assert(err, qt.IsNil)

	// test
	err = j.CleanupModelsDying(ctx)
	c.Assert(err, qt.IsNil)

	model := dbmodel.Model{
		UUID: sql.NullString{
			String: env.Models[0].UUID,
			Valid:  true,
		},
	}
	err = j.DB().GetModel(ctx, &model)
	c.Assert(err, qt.ErrorMatches, "model not found")

	model = dbmodel.Model{
		UUID: sql.NullString{
			String: env.Models[1].UUID,
			Valid:  true,
		},
	}
	err = j.DB().GetModel(ctx, &model)
	c.Assert(err, qt.IsNil)
}

func TestPollModelsDyingControllerErrors(t *testing.T) {
	// init
	c := qt.New(t)
	ctx := context.Background()

	client, _, _, err := jimmtest.SetupTestOFGAClient(c.Name())
	c.Assert(err, qt.IsNil)
	j := &jimm.JIMM{
		UUID:          uuid.NewString(),
		OpenFGAClient: client,
		Database: db.Database{
			DB: jimmtest.PostgresDB(c, nil),
		},
	}
	err = j.Database.Migrate(ctx, false)
	c.Assert(err, qt.IsNil)

	env := jimmtest.ParseEnvironment(c, modelPollerTestEnv)
	env.PopulateDBAndPermissions(c, j.ResourceTag(), j.Database, client)
	jimmAdmin, err := j.GetUser(ctx, "alice@canonical.com")
	c.Assert(err, qt.IsNil)
	j.Dialer = &jimmtest.Dialer{
		API: &jimmtest.API{
			ModelInfo_: func(ctx context.Context, mi *jujuparams.ModelInfo) error {
				return errors.E("controller not available")
			},
			DestroyModel_: func(ctx context.Context, mt names.ModelTag, b1, b2 *bool, d1, d2 *time.Duration) error {
				return nil
			},
		},
	}
	err = j.DestroyModel(ctx, jimmAdmin, names.NewModelTag(env.Models[0].UUID), nil, nil, nil, nil)
	c.Assert(err, qt.IsNil)

	// test
	err = j.CleanupModelsDying(ctx)
	c.Assert(err, qt.IsNil)

	model := dbmodel.Model{
		UUID: sql.NullString{
			String: env.Models[0].UUID,
			Valid:  true,
		},
	}
	err = j.DB().GetModel(ctx, &model)
	c.Assert(err, qt.IsNil)
	c.Assert(model.Life, qt.Equals, state.Dying.String())
}
