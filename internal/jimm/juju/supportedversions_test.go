// Copyright 2026 Canonical.

package juju_test

import (
	"context"
	"testing"

	qt "github.com/frankban/quicktest"
	"github.com/juju/version/v2"

	"github.com/canonical/jimm/v3/internal/jimm/juju"
)

func TestSupportedVersions_NoContextualVersion(t *testing.T) {
	c := qt.New(t)

	manager := &juju.JujuManager{}
	resp, err := manager.SupportedVersions(context.Background(), nil)
	c.Assert(err, qt.IsNil)
	c.Assert(len(resp.Versions) > 0, qt.IsTrue)
	for _, v := range resp.Versions {
		c.Assert(v.Version, qt.Not(qt.Equals), "")
		c.Assert(v.Date, qt.Not(qt.Equals), "")
		c.Assert(v.LinkToRelease, qt.Not(qt.Equals), "")
	}
}

func TestSupportedVersions_WithContextualVersion(t *testing.T) {
	c := qt.New(t)

	manager := &juju.JujuManager{}
	contextualVersion := "3.6.5"
	contextualVersionParsed := version.MustParse(contextualVersion)

	resp, err := manager.SupportedVersions(context.Background(), &contextualVersion)
	c.Assert(err, qt.IsNil)
	c.Assert(len(resp.Versions) > 0, qt.IsTrue)
	for _, v := range resp.Versions {
		parsed := version.MustParse(v.Version)
		c.Assert(parsed.Compare(contextualVersionParsed) > 0, qt.IsTrue,
			qt.Commentf("expected version %q to be greater than contextual version %q", v.Version, contextualVersion))
	}
}

func TestSupportedVersions_InvalidContextualVersion(t *testing.T) {
	c := qt.New(t)

	manager := &juju.JujuManager{}
	contextualVersion := "not-a-version"
	_, err := manager.SupportedVersions(context.Background(), &contextualVersion)
	c.Assert(err, qt.ErrorMatches, `invalid contextual version.*`)
}
