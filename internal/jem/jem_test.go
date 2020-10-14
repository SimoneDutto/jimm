// Copyright 2015 Canonical Ltd.

package jem_test

import (
	"context"
	"time"

	jujuapi "github.com/juju/juju/api"
	cloudapi "github.com/juju/juju/api/cloud"
	jujuparams "github.com/juju/juju/apiserver/params"
	"github.com/juju/juju/cloud"
	"github.com/juju/names/v4"
	jt "github.com/juju/testing"
	jc "github.com/juju/testing/checkers"
	"github.com/juju/utils"
	"github.com/juju/version"
	gc "gopkg.in/check.v1"
	"gopkg.in/errgo.v1"
	"gopkg.in/macaroon-bakery.v2/httpbakery"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/CanonicalLtd/jimm/internal/conv"
	"github.com/CanonicalLtd/jimm/internal/jem"
	"github.com/CanonicalLtd/jimm/internal/jemtest"
	"github.com/CanonicalLtd/jimm/internal/mgosession"
	"github.com/CanonicalLtd/jimm/internal/mongodoc"
	"github.com/CanonicalLtd/jimm/internal/pubsub"
	"github.com/CanonicalLtd/jimm/params"
)

var testContext = context.Background()

type jemSuite struct {
	jemtest.JujuConnSuite
	pool                           *jem.Pool
	sessionPool                    *mgosession.Pool
	jem                            *jem.JEM
	usageSenderAuthorizationClient *testUsageSenderAuthorizationClient

	suiteCleanups []func()
}

var _ = gc.Suite(&jemSuite{})

func (s *jemSuite) SetUpTest(c *gc.C) {
	s.JujuConnSuite.SetUpTest(c)
	s.sessionPool = mgosession.NewPool(context.TODO(), s.Session, 5)
	publicCloudMetadata, _, err := cloud.PublicCloudMetadata()
	c.Assert(err, gc.Equals, nil)
	s.usageSenderAuthorizationClient = &testUsageSenderAuthorizationClient{}
	s.PatchValue(&jem.NewUsageSenderAuthorizationClient, func(_ string, _ *httpbakery.Client) (jem.UsageSenderAuthorizationClient, error) {
		return s.usageSenderAuthorizationClient, nil
	})
	pool, err := jem.NewPool(context.TODO(), jem.Params{
		DB:                  s.Session.DB("jem"),
		ControllerAdmin:     "controller-admin",
		SessionPool:         s.sessionPool,
		PublicCloudMetadata: publicCloudMetadata,
		UsageSenderURL:      "test-usage-sender-url",
		Pubsub: &pubsub.Hub{
			MaxConcurrency: 10,
		},
	})
	c.Assert(err, gc.Equals, nil)
	s.pool = pool
	s.jem = s.pool.JEM(context.TODO())
	s.PatchValue(&utils.OutgoingAccessAllowed, true)
}

func (s *jemSuite) TearDownTest(c *gc.C) {
	s.jem.Close()
	s.pool.Close()
	s.sessionPool.Close()
	s.JujuConnSuite.TearDownTest(c)
}

func (s *jemSuite) TestPoolRequiresControllerAdmin(c *gc.C) {
	pool, err := jem.NewPool(context.TODO(), jem.Params{
		DB: s.Session.DB("jem"),
	})
	c.Assert(err, gc.ErrorMatches, "no controller admin group specified")
	c.Assert(pool, gc.IsNil)
}

func (s *jemSuite) TestPoolDoesNotReuseDeadConnection(c *gc.C) {
	session := jt.NewProxiedSession(c)
	defer session.Close()
	sessionPool := mgosession.NewPool(context.TODO(), session.Session, 3)
	defer sessionPool.Close()
	pool, err := jem.NewPool(context.TODO(), jem.Params{
		DB:              session.DB("jem"),
		ControllerAdmin: "controller-admin",
		SessionPool:     sessionPool,
	})
	c.Assert(err, gc.Equals, nil)
	defer pool.Close()

	assertOK := func(j *jem.JEM) {
		m := mongodoc.Model{Path: params.EntityPath{"bob", "x"}}
		err := j.DB.GetModel(testContext, &m)
		c.Assert(errgo.Cause(err), gc.Equals, params.ErrNotFound)
	}
	assertBroken := func(j *jem.JEM) {
		m := mongodoc.Model{Path: params.EntityPath{"bob", "x"}}
		err = j.DB.GetModel(testContext, &m)
		c.Assert(err, gc.ErrorMatches, `cannot get model: EOF`)
	}

	// Get a JEM instance and perform a single operation so that the session used by the
	// JEM instance obtains a mongo socket.
	c.Logf("make jem0")
	jem0 := pool.JEM(context.TODO())
	defer jem0.Close()
	assertOK(jem0)

	c.Logf("close connections")
	// Close all current connections to the mongo instance,
	// which should cause subsequent operations on jem1 to fail.
	session.CloseConns()

	// Get another JEM instance, which should be a new session,
	// so operations on it should not fail.
	c.Logf("make jem1")
	jem1 := pool.JEM(context.TODO())
	defer jem1.Close()
	assertOK(jem1)

	// Get another JEM instance which should clone the same session
	// used by jem0 because only two sessions are available.
	c.Logf("make jem2")
	jem2 := pool.JEM(context.TODO())
	defer jem2.Close()

	// Perform another operation on jem0, which should fail and
	// cause its session not to be reused.
	c.Logf("check jem0 is broken")
	assertBroken(jem0)

	// The jem1 connection should still be working because it
	// was created after the connections were broken.
	c.Logf("check jem1 is ok")
	assertOK(jem1)

	c.Logf("check jem2 is ok")
	// The jem2 connection should also be broken because it
	// reused the same sessions as jem0
	assertBroken(jem2)

	// Get another instance, which should reuse the jem3 connection
	// and work OK.
	c.Logf("make jem3")
	jem3 := pool.JEM(context.TODO())
	defer jem3.Close()
	assertOK(jem3)

	// When getting the next instance, we should see that the connection
	// that we would have used is broken and create another one.
	c.Logf("make jem4")
	jem4 := pool.JEM(context.TODO())
	defer jem4.Close()
	assertOK(jem4)
}

func (s *jemSuite) TestClone(c *gc.C) {
	j := s.jem.Clone()
	j.Close()
	m := mongodoc.Model{Path: params.EntityPath{"bob", "x"}}
	err := s.jem.DB.GetModel(testContext, &m)
	c.Assert(errgo.Cause(err), gc.Equals, params.ErrNotFound)
}

func (s *jemSuite) TestRevokeCredentialsInUse(c *gc.C) {
	ctlId := s.addController(c, params.EntityPath{"bob", "controller"})
	err := s.jem.DB.SetACL(testContext, s.jem.DB.Controllers(), ctlId, params.ACL{
		Read: []string{"everyone"},
	})
	c.Assert(err, gc.Equals, nil)
	credPath := credentialPath("dummy", "bob", "cred1")
	err = s.jem.DB.UpdateCredential(testContext, &mongodoc.Credential{
		Path: mongodoc.CredentialPathFromParams(credPath),
		Type: "empty",
	})
	c.Assert(err, gc.Equals, nil)

	id := jemtest.NewIdentity("bob")
	err = s.jem.CreateModel(testContext, id, jem.CreateModelParams{
		Path:           params.EntityPath{"bob", "oldmodel"},
		ControllerPath: ctlId,
		Credential:     credPath,
		Cloud:          "dummy",
	}, nil)
	c.Assert(err, gc.Equals, nil)
	err = s.jem.RevokeCredential(testContext, credPath, 0)
	c.Assert(err, gc.ErrorMatches, `cannot revoke because credential is in use on at least one model`)

	// Try with just the check.
	err = s.jem.RevokeCredential(testContext, credPath, jem.CredentialCheck)
	c.Assert(err, gc.ErrorMatches, `cannot revoke because credential is in use on at least one model`)

	// Try without the check. It should succeed.
	err = s.jem.RevokeCredential(testContext, credPath, jem.CredentialUpdate)
	c.Assert(err, gc.Equals, nil)

	// Try to create another with the credentials that have
	// been revoked. We should fail to do that.
	err = s.jem.CreateModel(testContext, id, jem.CreateModelParams{
		Path:           params.EntityPath{"bob", "newmodel"},
		ControllerPath: ctlId,
		Credential:     credPath,
		Cloud:          "dummy",
	}, nil)
	c.Assert(err, gc.ErrorMatches, `credential dummy/bob/cred1 has been revoked`)

	// Check that the credential really has been revoked on the
	// controller.
	conn, err := s.jem.OpenAPI(testContext, ctlId)
	c.Assert(err, gc.Equals, nil)
	defer conn.Close()
	r, err := cloudapi.NewClient(conn).Credentials(conv.ToCloudCredentialTag(credPath))
	c.Assert(err, gc.Equals, nil)
	c.Assert(r, jc.DeepEquals, []jujuparams.CloudCredentialResult{{
		Error: &jujuparams.Error{
			Message: `credential "cred1" not found`,
			Code:    "not found",
		},
	}})
}

func (s *jemSuite) TestRevokeCredentialsNotInUse(c *gc.C) {
	ctlId := s.addController(c, params.EntityPath{"bob", "controller"})
	err := s.jem.DB.SetACL(testContext, s.jem.DB.Controllers(), ctlId, params.ACL{
		Read: []string{"everyone"},
	})
	c.Assert(err, gc.Equals, nil)
	credPath := credentialPath("dummy", "bob", "cred1")
	mCredPath := mgoCredentialPath("dummy", "bob", "cred1")
	err = s.jem.DB.UpdateCredential(testContext, &mongodoc.Credential{
		Path: mCredPath,
		Type: "empty",
	})
	c.Assert(err, gc.Equals, nil)

	// Sanity check that we can get the credential.
	err = s.jem.DB.GetCredential(testContext, &mongodoc.Credential{Path: mCredPath})
	c.Assert(err, gc.Equals, nil)

	// Try with just the check.
	err = s.jem.RevokeCredential(testContext, credPath, jem.CredentialCheck)
	c.Assert(err, gc.Equals, nil)

	// Check that the credential is still there.
	err = s.jem.DB.GetCredential(testContext, &mongodoc.Credential{Path: mCredPath})
	c.Assert(err, gc.Equals, nil)

	// Try with both the check and the update flag.
	err = s.jem.RevokeCredential(testContext, credPath, 0)
	c.Assert(err, gc.Equals, nil)

	// The credential should be marked as revoked and all
	// the details should be cleater.
	cred := mongodoc.Credential{
		Path: mCredPath,
	}
	err = s.jem.DB.GetCredential(testContext, &cred)
	c.Assert(err, gc.Equals, nil)
	c.Assert(cred, jc.DeepEquals, mongodoc.Credential{
		Id:         "dummy/bob/cred1",
		Path:       mCredPath,
		Revoked:    true,
		Attributes: make(map[string]string),
	})
}

func (s *jemSuite) TestUpdateCredential(c *gc.C) {
	ctlPath := s.addController(c, params.EntityPath{User: "bob", Name: "controller"})
	credPath := credentialPath("dummy", "bob", "cred")
	mCredPath := mgoCredentialPath("dummy", "bob", "cred")
	cred := &mongodoc.Credential{
		Path: mCredPath,
		Type: "empty",
	}
	err := s.jem.DB.UpdateCredential(testContext, cred)
	c.Assert(err, gc.Equals, nil)
	conn, err := s.jem.OpenAPI(testContext, ctlPath)
	c.Assert(err, gc.Equals, nil)
	defer conn.Close()

	_, err = jem.UpdateControllerCredential(s.jem, testContext, conn, ctlPath, cred)
	c.Assert(err, gc.Equals, nil)
	err = s.jem.DB.CredentialAddController(testContext, mCredPath, ctlPath)
	c.Assert(err, gc.Equals, nil)

	// Sanity check it was deployed
	client := cloudapi.NewClient(conn)
	credTag := names.NewCloudCredentialTag("dummy/bob@external/cred")
	creds, err := client.Credentials(credTag)
	c.Assert(err, gc.Equals, nil)
	c.Assert(creds, jc.DeepEquals, []jujuparams.CloudCredentialResult{{
		Result: &jujuparams.CloudCredential{
			AuthType: "empty",
		},
	}})

	_, err = s.jem.UpdateCredential(testContext, &mongodoc.Credential{
		Path: mCredPath,
		Type: "userpass",
		Attributes: map[string]string{
			"username": "cloud-user",
			"password": "cloud-pass",
		},
	}, jem.CredentialUpdate)
	c.Assert(err, gc.Equals, nil)

	// check it was updated on the controller.
	creds, err = client.Credentials(credTag)
	c.Assert(err, gc.Equals, nil)
	c.Assert(creds, jc.DeepEquals, []jujuparams.CloudCredentialResult{{
		Result: &jujuparams.CloudCredential{
			AuthType: "userpass",
			Attributes: map[string]string{
				"username": "cloud-user",
			},
			Redacted: []string{
				"password",
			},
		},
	}})

	// Revoke the credential
	err = s.jem.RevokeCredential(testContext, credPath, jem.CredentialUpdate)
	c.Assert(err, gc.Equals, nil)

	// check it was removed on the controller.
	creds, err = client.Credentials(credTag)
	c.Assert(err, gc.Equals, nil)
	c.Assert(creds, jc.DeepEquals, []jujuparams.CloudCredentialResult{{
		Error: &jujuparams.Error{
			Code:    "not found",
			Message: `credential "cred" not found`,
		},
	}})
}

var credentialTests = []struct {
	path             params.CredentialPath
	expectErrorCause error
}{{
	path: params.CredentialPath{
		Cloud: "dummy",
		User:  "bob",
		Name:  "credential",
	},
}, {
	path: params.CredentialPath{
		Cloud: "dummy",
		User:  "bob-group",
		Name:  "credential",
	},
}, {
	path: params.CredentialPath{
		Cloud: "dummy",
		User:  "alice",
		Name:  "credential",
	},
	expectErrorCause: params.ErrUnauthorized,
}, {
	path: params.CredentialPath{
		Cloud: "dummy",
		User:  "bob",
		Name:  "credential2",
	},
	expectErrorCause: params.ErrNotFound,
}, {
	path: params.CredentialPath{
		Cloud: "dummy",
		User:  "bob-group",
		Name:  "credential2",
	},
	expectErrorCause: params.ErrNotFound,
}, {
	path: params.CredentialPath{
		Cloud: "dummy",
		User:  "alice",
		Name:  "credential2",
	},
	expectErrorCause: params.ErrUnauthorized,
}}

func (s *jemSuite) TestCredential(c *gc.C) {
	creds := []mongodoc.Credential{{
		Path: mongodoc.CredentialPath{
			Cloud: "dummy",
			EntityPath: mongodoc.EntityPath{
				User: "alice",
				Name: "credential",
			},
		},
	}, {
		Path: mongodoc.CredentialPath{
			Cloud: "dummy",
			EntityPath: mongodoc.EntityPath{
				User: "bob",
				Name: "credential",
			},
		},
	}, {
		Path: mongodoc.CredentialPath{
			Cloud: "dummy",
			EntityPath: mongodoc.EntityPath{
				User: "bob-group",
				Name: "credential",
			},
		},
	}}
	for _, cred := range creds {
		cred.Id = cred.Path.String()
		err := s.jem.DB.UpdateCredential(testContext, &cred)
		c.Assert(err, gc.Equals, nil)
	}
	for i, test := range credentialTests {
		c.Logf("test %d. %s", i, test.path)
		ctl := mongodoc.Credential{
			Path: mongodoc.CredentialPathFromParams(test.path),
		}
		err := s.jem.GetCredential(context.Background(), jemtest.NewIdentity("bob", "bob-group"), &ctl)
		if test.expectErrorCause != nil {
			c.Assert(errgo.Cause(err), gc.Equals, test.expectErrorCause)
			continue
		}
		c.Assert(err, gc.Equals, nil)
		c.Assert(ctl.Path.ToParams(), jc.DeepEquals, test.path)
	}
}

var earliestControllerVersionTests = []struct {
	about       string
	controllers []mongodoc.Controller
	expect      version.Number
}{{
	about:  "no controllers",
	expect: version.Number{},
}, {
	about: "one controller",
	controllers: []mongodoc.Controller{{
		Path:    params.EntityPath{"bob", "c1"},
		Public:  true,
		Version: &version.Number{Minor: 1},
		ACL: params.ACL{
			Read: []string{"everyone"},
		},
	}},
	expect: version.Number{Minor: 1},
}, {
	about: "multiple controllers",
	controllers: []mongodoc.Controller{{
		Path:    params.EntityPath{"bob", "c1"},
		Public:  true,
		Version: &version.Number{Minor: 1},
		ACL: params.ACL{
			Read: []string{"everyone"},
		},
	}, {
		Path:    params.EntityPath{"bob", "c2"},
		Public:  true,
		Version: &version.Number{Minor: 2},
		ACL: params.ACL{
			Read: []string{"everyone"},
		},
	}, {
		Path:    params.EntityPath{"bob", "c3"},
		Public:  true,
		Version: &version.Number{Minor: 3},
		ACL: params.ACL{
			Read: []string{"everyone"},
		},
	}},
	expect: version.Number{Minor: 1},
}, {
	about: "non-public controllers ignored",
	controllers: []mongodoc.Controller{{
		Path:    params.EntityPath{"bob", "c1"},
		Version: &version.Number{Minor: 1},
	}, {
		Path:   params.EntityPath{"bob", "c2"},
		Public: true,
		ACL: params.ACL{
			Read: []string{"everyone"},
		},
		Version: &version.Number{Minor: 2},
	}},
	expect: version.Number{Minor: 2},
}}

func (s *jemSuite) TestEarliestControllerVersion(c *gc.C) {
	for i, test := range earliestControllerVersionTests {
		c.Logf("test %d: %v", i, test.about)
		_, err := s.jem.DB.Controllers().RemoveAll(nil)
		c.Assert(err, gc.Equals, nil)
		for _, ctl := range test.controllers {
			err := s.jem.DB.InsertController(testContext, &ctl)
			c.Assert(err, gc.Equals, nil)
		}
		v, err := s.jem.EarliestControllerVersion(testContext, jemtest.NewIdentity("someone"))
		c.Assert(err, gc.Equals, nil)
		c.Assert(v, jc.DeepEquals, test.expect)
	}
}

func (s *jemSuite) TestUpdateMachineInfo(c *gc.C) {
	m := s.bootstrapModel(c, params.EntityPath{"bob", "model-1"})
	ctlPath := params.EntityPath{"bob", "controller"}

	err := s.jem.UpdateMachineInfo(testContext, ctlPath, &jujuparams.MachineInfo{
		ModelUUID: m.UUID,
		Id:        "0",
		Series:    "quantal",
	})
	c.Assert(err, gc.Equals, nil)
	err = s.jem.UpdateMachineInfo(testContext, ctlPath, &jujuparams.MachineInfo{
		ModelUUID: m.UUID,
		Id:        "1",
		Series:    "precise",
	})
	c.Assert(err, gc.Equals, nil)

	docs, err := s.jem.DB.MachinesForModel(testContext, m.UUID)
	c.Assert(err, gc.Equals, nil)
	for i := range docs {
		cleanMachineDoc(&docs[i])
	}
	c.Assert(docs, jc.DeepEquals, []mongodoc.Machine{{
		Id:         ctlPath.String() + " " + m.UUID + " 0",
		Controller: ctlPath.String(),
		Cloud:      "dummy",
		Region:     "dummy-region",
		Info: &jujuparams.MachineInfo{
			ModelUUID: m.UUID,
			Id:        "0",
			Series:    "quantal",
			Config:    map[string]interface{}{},
		},
	}, {
		Id:         ctlPath.String() + " " + m.UUID + " 1",
		Controller: ctlPath.String(),
		Cloud:      "dummy",
		Region:     "dummy-region",
		Info: &jujuparams.MachineInfo{
			ModelUUID: m.UUID,
			Id:        "1",
			Series:    "precise",
			Config:    map[string]interface{}{},
		},
	}})

	// Check that we can update one of the documents.
	err = s.jem.UpdateMachineInfo(testContext, ctlPath, &jujuparams.MachineInfo{
		ModelUUID: m.UUID,
		Id:        "0",
		Series:    "quantal",
		Life:      "dying",
	})
	c.Assert(err, gc.Equals, nil)

	// Check that setting a machine dead removes it.
	err = s.jem.UpdateMachineInfo(testContext, ctlPath, &jujuparams.MachineInfo{
		ModelUUID: m.UUID,
		Id:        "1",
		Series:    "precise",
		Life:      "dead",
	})
	c.Assert(err, gc.Equals, nil)

	docs, err = s.jem.DB.MachinesForModel(testContext, m.UUID)
	c.Assert(err, gc.Equals, nil)
	for i := range docs {
		cleanMachineDoc(&docs[i])
	}
	c.Assert(docs, jc.DeepEquals, []mongodoc.Machine{{
		Id:         ctlPath.String() + " " + m.UUID + " 0",
		Controller: ctlPath.String(),
		Cloud:      "dummy",
		Region:     "dummy-region",
		Info: &jujuparams.MachineInfo{
			ModelUUID: m.UUID,
			Id:        "0",
			Series:    "quantal",
			Config:    map[string]interface{}{},
			Life:      "dying",
		},
	}})
}

func (s *jemSuite) TestUpdateMachineUnknownModel(c *gc.C) {
	ctlPath := params.EntityPath{"bob", "controller"}

	err := s.jem.UpdateMachineInfo(testContext, ctlPath, &jujuparams.MachineInfo{
		ModelUUID: "no-such-uuid",
		Id:        "1",
		Series:    "precise",
	})
	c.Assert(err, gc.Equals, nil)
}

func (s *jemSuite) TestUpdateMachineIncorrectController(c *gc.C) {
	m := s.bootstrapModel(c, params.EntityPath{"bob", "model-1"})
	ctlPath := params.EntityPath{"bob", "controller2"}

	err := s.jem.UpdateMachineInfo(testContext, ctlPath, &jujuparams.MachineInfo{
		ModelUUID: m.UUID,
		Id:        "1",
		Series:    "precise",
	})
	c.Assert(err, gc.Equals, nil)
}

func (s *jemSuite) TestUpdateApplicationInfo(c *gc.C) {
	m := s.bootstrapModel(c, params.EntityPath{"bob", "model-1"})
	ctlPath := params.EntityPath{"bob", "controller"}

	err := s.jem.UpdateApplicationInfo(testContext, ctlPath, &jujuparams.ApplicationInfo{
		ModelUUID: m.UUID,
		Name:      "0",
	})
	c.Assert(err, gc.Equals, nil)
	err = s.jem.UpdateApplicationInfo(testContext, ctlPath, &jujuparams.ApplicationInfo{
		ModelUUID: m.UUID,
		Name:      "1",
	})
	c.Assert(err, gc.Equals, nil)

	docs, err := s.jem.DB.ApplicationsForModel(testContext, m.UUID)
	c.Assert(err, gc.Equals, nil)
	for i := range docs {
		cleanApplicationDoc(&docs[i])
	}
	c.Assert(docs, jc.DeepEquals, []mongodoc.Application{{
		Id:         ctlPath.String() + " " + m.UUID + " 0",
		Controller: ctlPath.String(),
		Cloud:      "dummy",
		Region:     "dummy-region",
		Info: &mongodoc.ApplicationInfo{
			ModelUUID: m.UUID,
			Name:      "0",
		},
	}, {
		Id:         ctlPath.String() + " " + m.UUID + " 1",
		Controller: ctlPath.String(),
		Cloud:      "dummy",
		Region:     "dummy-region",
		Info: &mongodoc.ApplicationInfo{
			ModelUUID: m.UUID,
			Name:      "1",
		},
	}})

	// Check that we can update one of the documents.
	err = s.jem.UpdateApplicationInfo(testContext, ctlPath, &jujuparams.ApplicationInfo{
		ModelUUID: m.UUID,
		Name:      "0",
		Life:      "dying",
	})
	c.Assert(err, gc.Equals, nil)

	// Check that setting an application dead removes it.
	err = s.jem.UpdateApplicationInfo(testContext, ctlPath, &jujuparams.ApplicationInfo{
		ModelUUID: m.UUID,
		Name:      "1",
		Life:      "dead",
	})
	c.Assert(err, gc.Equals, nil)

	docs, err = s.jem.DB.ApplicationsForModel(testContext, m.UUID)
	c.Assert(err, gc.Equals, nil)
	for i := range docs {
		cleanApplicationDoc(&docs[i])
	}
	c.Assert(docs, jc.DeepEquals, []mongodoc.Application{{
		Id:         ctlPath.String() + " " + m.UUID + " 0",
		Controller: ctlPath.String(),
		Cloud:      "dummy",
		Region:     "dummy-region",
		Info: &mongodoc.ApplicationInfo{
			ModelUUID: m.UUID,
			Name:      "0",
			Life:      "dying",
		},
	}})
}

func (s *jemSuite) TestUpdateApplicationUnknownModel(c *gc.C) {
	m := s.bootstrapModel(c, params.EntityPath{"bob", "model-1"})
	ctlPath := params.EntityPath{"bob", "controller"}

	err := s.jem.UpdateApplicationInfo(testContext, ctlPath, &jujuparams.ApplicationInfo{
		ModelUUID: m.UUID,
		Name:      "1",
	})
	c.Assert(err, gc.Equals, nil)
}

func (s *jemSuite) TestUpdateModelCredential(c *gc.C) {
	model := s.bootstrapModel(c, params.EntityPath{User: "bob", Name: "model"})

	credPath := credentialPath("dummy", "bob", "cred2")
	err := s.jem.DB.UpdateCredential(testContext, &mongodoc.Credential{
		Path: mongodoc.CredentialPathFromParams(credPath),
		Type: "empty",
	})

	conn, err := s.jem.OpenAPI(testContext, model.Controller)
	c.Assert(err, gc.Equals, nil)
	defer conn.Close()

	err = s.jem.UpdateModelCredential(testContext, conn, model, &mongodoc.Credential{
		Path: mongodoc.CredentialPathFromParams(credPath),
		Type: "empty",
	})
	c.Assert(err, gc.Equals, nil)
	model1 := mongodoc.Model{Path: model.Path}
	err = s.jem.DB.GetModel(testContext, &model1)
	c.Assert(err, gc.Equals, nil)
	c.Assert(model1.Credential, jc.DeepEquals, mongodoc.CredentialPathFromParams(credPath))
}

func (s *jemSuite) TestWatchAllModelSummaries(c *gc.C) {
	s.addController(c, params.EntityPath{"bob", "controller"})
	ctlPath := params.EntityPath{User: "bob", Name: "controller"}

	pubsub := s.jem.Pubsub()
	summaryChannel := make(chan interface{}, 1)
	handlerFunction := func(_ string, summary interface{}) {
		select {
		case summaryChannel <- summary:
		default:
		}
	}
	cleanup, err := pubsub.Subscribe("deadbeef-0bad-400d-8000-4b1d0d06f00d", handlerFunction)
	c.Assert(err, jc.ErrorIsNil)
	defer cleanup()

	watcherCleanup, err := s.jem.WatchAllModelSummaries(context.Background(), ctlPath)
	c.Assert(err, gc.Equals, nil)
	defer func() {
		err := watcherCleanup()
		if err != nil {
			c.Logf("failed to stop all model summaries watcher: %v", err)
		}
	}()

	select {
	case summary := <-summaryChannel:
		c.Assert(summary, gc.DeepEquals,
			jujuparams.ModelAbstract{
				UUID:       "deadbeef-0bad-400d-8000-4b1d0d06f00d",
				Removed:    false,
				Controller: "",
				Name:       "controller",
				Admins:     []string{"admin"},
				Cloud:      "dummy",
				Region:     "dummy-region",
				Credential: "dummy/admin/cred",
				Size: jujuparams.ModelSummarySize{
					Machines:     0,
					Containers:   0,
					Applications: 0,
					Units:        0,
					Relations:    0,
				},
				Status: "green",
			})
	case <-time.After(time.Second):
		c.Fatal("timed out")
	}
}

func (s *jemSuite) TestGetModel(c *gc.C) {
	model := s.bootstrapModel(c, params.EntityPath{"test", "model"})

	model1 := mongodoc.Model{Path: model.Path}
	err := s.jem.GetModel(testContext, jemtest.NewIdentity("test"), jujuparams.ModelReadAccess, &model1)
	c.Assert(err, gc.Equals, nil)
	c.Assert(model1, jc.DeepEquals, *model)

	_, err = s.jem.DB.Models().FindId(model.Id).Apply(mgo.Change{
		Update: bson.D{{"$unset", bson.D{
			{"cloud", ""},
			{"cloudregion", ""},
			{"credential", ""},
			{"defaultseries", ""},
		}}},
	}, nil)
	c.Assert(err, gc.Equals, nil)

	model2 := mongodoc.Model{UUID: model.UUID}
	err = s.jem.GetModel(testContext, jemtest.NewIdentity("test"), jujuparams.ModelReadAccess, &model2)
	c.Assert(err, gc.Equals, nil)

	c.Assert(model2, gc.DeepEquals, model1)
}

func (s *jemSuite) TestGetModelUnauthorized(c *gc.C) {
	model := s.bootstrapModel(c, params.EntityPath{"test", "model"})

	model1 := mongodoc.Model{Path: model.Path}
	err := s.jem.GetModel(testContext, jemtest.NewIdentity("not-test"), jujuparams.ModelReadAccess, &model1)
	c.Assert(err, gc.ErrorMatches, "unauthorized")
	c.Assert(errgo.Cause(err), gc.Equals, params.ErrUnauthorized)
}

func (s *jemSuite) addController(c *gc.C, path params.EntityPath) params.EntityPath {
	return addController(c, path, s.APIInfo(c), s.jem)
}

func addController(c *gc.C, path params.EntityPath, info *jujuapi.Info, jem *jem.JEM) params.EntityPath {
	hps, err := mongodoc.ParseAddresses(info.Addrs)
	c.Assert(err, gc.Equals, nil)

	ctl := &mongodoc.Controller{
		Path:          path,
		HostPorts:     [][]mongodoc.HostPort{hps},
		CACert:        info.CACert,
		AdminUser:     info.Tag.Id(),
		AdminPassword: info.Password,
		Public:        true,
	}
	err = jem.AddController(testContext, jemtest.NewIdentity(string(path.User), string(jem.ControllerAdmin())), ctl)
	c.Assert(err, gc.Equals, nil)

	return path
}

func (s *jemSuite) bootstrapModel(c *gc.C, path params.EntityPath) *mongodoc.Model {
	return bootstrapModel(c, path, s.APIInfo(c), s.jem)
}

func bootstrapModel(c *gc.C, path params.EntityPath, info *jujuapi.Info, j *jem.JEM) *mongodoc.Model {
	ctlPath := addController(c, params.EntityPath{User: path.User, Name: "controller"}, info, j)
	credPath := credentialPath("dummy", string(path.User), "cred")
	err := j.DB.UpdateCredential(testContext, &mongodoc.Credential{
		Path: mongodoc.CredentialPathFromParams(credPath),
		Type: "empty",
	})
	c.Assert(err, gc.Equals, nil)
	err = j.CreateModel(testContext, jemtest.NewIdentity(string(path.User)), jem.CreateModelParams{
		Path:           path,
		ControllerPath: ctlPath,
		Credential: params.CredentialPath{
			Cloud: "dummy",
			User:  path.User,
			Name:  "cred",
		},
		Cloud: "dummy",
	}, nil)
	c.Assert(err, gc.Equals, nil)
	model := mongodoc.Model{
		Path: path,
	}
	err = j.DB.GetModel(testContext, &model)
	return &model
}

type testUsageSenderAuthorizationClient struct {
	errors []error
}

func (c *testUsageSenderAuthorizationClient) SetErrors(errors []error) {
	c.errors = errors
}

func (c *testUsageSenderAuthorizationClient) GetCredentials(ctx context.Context, applicationUser string) ([]byte, error) {
	var err error
	if len(c.errors) > 0 {
		err, c.errors = c.errors[0], c.errors[1:]
	}
	return []byte("test credentials"), err
}

func credentialPath(cloud, user, name string) params.CredentialPath {
	return params.CredentialPath{
		Cloud: params.Cloud(cloud),
		User:  params.User(user),
		Name:  params.CredentialName(name),
	}
}

func mgoCredentialPath(cloud, user, name string) mongodoc.CredentialPath {
	return mongodoc.CredentialPath{
		Cloud: cloud,
		EntityPath: mongodoc.EntityPath{
			User: user,
			Name: name,
		},
	}
}

// cleanMachineDoc cleans up the machine document so
// that we can use a DeepEqual comparison without worrying
// about non-nil vs nil map comparisons.
func cleanMachineDoc(doc *mongodoc.Machine) {
	if len(doc.Info.AgentStatus.Data) == 0 {
		doc.Info.AgentStatus.Data = nil
	}
	if len(doc.Info.InstanceStatus.Data) == 0 {
		doc.Info.InstanceStatus.Data = nil
	}
}

// cleanApplicationDoc cleans up the application document so
// that we can use a DeepEqual comparison without worrying
// about non-nil vs nil map comparisons.
func cleanApplicationDoc(doc *mongodoc.Application) {
	if len(doc.Info.Status.Data) == 0 {
		doc.Info.Status.Data = nil
	}
}
