// Copyright 2015 Canonical Ltd.

package jemcmd_test

import (
	gc "gopkg.in/check.v1"
	"gopkg.in/errgo.v1"

	"github.com/CanonicalLtd/jem/params"
)

type addServerSuite struct {
	commonSuite
}

var _ = gc.Suite(&addServerSuite{})

func (s *addServerSuite) TestAddServer(c *gc.C) {
	s.idmSrv.AddUser("bob")
	s.idmSrv.SetDefaultUser("bob")
	client := s.jemClient("bob")
	_, err := client.GetJES(&params.GetJES{
		EntityPath: params.EntityPath{
			User: "bob",
			Name: "foo",
		},
	})
	c.Assert(errgo.Cause(err), gc.Equals, params.ErrNotFound)
	stdout, stderr, code := run(c, c.MkDir(), "add-server", "bob/foo")
	c.Assert(code, gc.Equals, 0, gc.Commentf("stderr: %s", stderr))
	c.Assert(stdout, gc.Equals, "")
	c.Assert(stderr, gc.Equals, "")
	_, err = client.GetJES(&params.GetJES{
		EntityPath: params.EntityPath{
			User: "bob",
			Name: "foo",
		},
	})
	c.Assert(err, gc.IsNil)
}

var addServerErrorTests = []struct {
	about        string
	args         []string
	expectStderr string
	expectCode   int
}{{
	about:        "too few arguments",
	args:         []string{},
	expectStderr: "got 0 arguments, want 1",
	expectCode:   2,
}, {
	about:        "too many arguments",
	args:         []string{"a", "b", "c"},
	expectStderr: "got 3 arguments, want 1",
	expectCode:   2,
}, {
	about:        "invalid server name",
	args:         []string{"a"},
	expectStderr: `invalid entity path "a": wrong number of parts in entity path`,
	expectCode:   2,
}, {
	about:        "invalid name checked by server",
	args:         []string{"bad!name/foo"},
	expectStderr: `invalid entity path "bad!name/foo": invalid user name "bad!name"`,
	expectCode:   2,
}}

func (s *addServerSuite) TestAddServerError(c *gc.C) {
	for i, test := range addServerErrorTests {
		c.Logf("test %d: %s", i, test.about)
		stdout, stderr, code := run(c, c.MkDir(), "add-server", test.args...)
		c.Assert(code, gc.Equals, test.expectCode, gc.Commentf("stderr: %s", stderr))
		c.Assert(stderr, gc.Matches, "(error:|ERROR) "+test.expectStderr+"\n")
		c.Assert(stdout, gc.Equals, "")
	}
}
