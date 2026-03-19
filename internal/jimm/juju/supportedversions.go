// Copyright 2026 Canonical.

package juju

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/canonical/jimm/v3/internal/errors"
	"github.com/canonical/jimm/v3/pkg/api/params"
	"github.com/juju/version/v2"
)

//go:embed juju-releases.json
var jujuReleasesData []byte

// SupportedVersions returns a list of supported Juju versions. The contextualVersion parameter can be used
// to filter.
func (j *JujuManager) SupportedVersions(ctx context.Context, contextualVersion *string) (params.SupportedJujuVersionsResponse, error) {
	var parsedContextualVersion version.Number
	var err error
	if contextualVersion != nil {
		parsedContextualVersion, err = version.Parse(*contextualVersion)
		if err != nil {
			return params.SupportedJujuVersionsResponse{}, fmt.Errorf("invalid contextual version %q: %w", *contextualVersion, err)
		}
	}

	releases, err := loadReleases()
	if err != nil {
		return params.SupportedJujuVersionsResponse{}, errors.E(err)
	}
	versionElems := make([]params.VersionElem, 0)
	for _, release := range releases {
		parsedVersion, err := version.Parse(release.Version)
		if err != nil {
			return params.SupportedJujuVersionsResponse{}, errors.E(fmt.Errorf("invalid release version %q in embedded juju releases: %w", release.Version, err))
		}
		if contextualVersion != nil {
			// if the release version is less than or equal to the contextual version, skip it
			if parsedVersion.Compare(parsedContextualVersion) != 1 {
				continue
			}
		}
		versionElems = append(versionElems, params.VersionElem{
			Version:       release.Version,
			Date:          release.Date,
			LinkToRelease: release.LinkToRelease,
		})
	}

	return params.SupportedJujuVersionsResponse{Versions: versionElems}, nil
}

func loadReleases() ([]params.VersionElem, error) {
	var releases []params.VersionElem
	if err := json.Unmarshal(jujuReleasesData, &releases); err != nil {
		return nil, fmt.Errorf("failed to unmarshal embedded juju releases: %w", err)
	}
	return releases, nil
}
