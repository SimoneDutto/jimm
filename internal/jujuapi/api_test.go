// Copyright 2016 Canonical Ltd.

package jujuapi_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	jc "github.com/juju/testing/checkers"
	"github.com/juju/testing/httptesting"
	gc "gopkg.in/check.v1"

	"github.com/CanonicalLtd/jem"
	"github.com/CanonicalLtd/jem/internal/apitest"
	"github.com/CanonicalLtd/jem/params"
)

type apiSuite struct {
	apitest.Suite
}

var _ = gc.Suite(&apiSuite{})

func (s *apiSuite) TestGUI(c *gc.C) {
	s.AssertAddController(c, params.EntityPath{User: "bob", Name: "controller-1"}, true)
	cred := s.AssertUpdateCredential(c, "bob", "dummy", "cred1", "empty")
	_, uuid := s.CreateModel(c, params.EntityPath{"bob", "gui-model"}, params.EntityPath{"bob", "controller-1"}, cred)
	jemSrv := s.NewServer(c, s.Session, s.IDMSrv, jem.ServerParams{
		GUILocation: "https://jujucharms.com.test",
	})
	defer jemSrv.Close()
	AssertRedirect(c, RedirectParams{
		Handler:        jemSrv,
		Method:         "GET",
		URL:            fmt.Sprintf("/gui/%s", uuid),
		ExpectCode:     http.StatusMovedPermanently,
		ExpectLocation: "https://jujucharms.com.test/u/bob/gui-model",
	})
}

func (s *apiSuite) TestGUINotFound(c *gc.C) {
	httptesting.AssertJSONCall(c, httptesting.JSONCallParams{
		URL:          fmt.Sprintf("/gui/%s", "000000000000-0000-0000-0000-00000000"),
		Handler:      s.JEMSrv,
		ExpectStatus: http.StatusNotFound,
		ExpectBody: params.Error{
			Code:    params.ErrNotFound,
			Message: "no GUI location specified",
		},
	})
}

func (s *apiSuite) TestGUIModelNotFound(c *gc.C) {
	jemSrv := s.NewServer(c, s.Session, s.IDMSrv, jem.ServerParams{
		GUILocation: "https://jujucharms.com.test",
	})
	defer jemSrv.Close()
	httptesting.AssertJSONCall(c, httptesting.JSONCallParams{
		URL:          fmt.Sprintf("/gui/%s", "000000000000-0000-0000-0000-00000000"),
		Handler:      jemSrv,
		ExpectStatus: http.StatusNotFound,
		ExpectBody: params.Error{
			Code:    params.ErrNotFound,
			Message: `model "000000000000-0000-0000-0000-00000000" not found`,
		},
	})
}

type RedirectParams struct {
	Method         string
	URL            string
	Handler        http.Handler
	ExpectCode     int
	ExpectLocation string
}

// AssertRedirect checks that a handler returns a redirect
func AssertRedirect(c *gc.C, p RedirectParams) {
	if p.Method == "" {
		p.Method = "GET"
	}
	req, err := http.NewRequest(p.Method, p.URL, nil)
	c.Assert(err, jc.ErrorIsNil)
	rr := httptest.NewRecorder()
	p.Handler.ServeHTTP(rr, req)
	if p.ExpectCode == 0 {
		c.Assert(300 <= rr.Code && rr.Code < 400, gc.Equals, true, gc.Commentf("Expected redirect status (3XX), got %d", rr.Code))
	} else {
		c.Assert(rr.Code, gc.Equals, p.ExpectCode)
	}
	c.Assert(rr.HeaderMap.Get("Location"), gc.Equals, p.ExpectLocation)
}