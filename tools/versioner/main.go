// Copyright 2026 Canonical.

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/google/go-github/v69/github"
	"github.com/juju/version/v2"
)

const (
	outputPath          = "internal/jimm/juju/juju-releases.json"
	minSupportedVersion = "3.6.5"
)

// blacklistedVersions lists specific releases that should be excluded from the output
// even if they otherwise meet the stable release criteria.
var blacklistedVersions = []version.Number{
	version.MustParse("4.0.0"),
	version.MustParse("4.0.1"),
	version.MustParse("4.0.2"),
}

type githubRelease struct {
	TagName     string    `json:"tag_name"`
	HTMLURL     string    `json:"html_url"`
	Draft       bool      `json:"draft"`
	Prerelease  bool      `json:"prerelease"`
	PublishedAt time.Time `json:"published_at"`
}

type outputRelease struct {
	Version string `json:"version"`
	Date    string `json:"date"`
	Link    string `json:"link-to-release"`
}

func main() {
	releases, err := fetchAllReleases()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error fetching releases: %v\n", err)
		os.Exit(1)
	}

	filtered := filterStableFinalReleases(releases)
	output := sortReleases(filtered)

	if err := writeOutputJSON(outputPath, output); err != nil {
		fmt.Fprintf(os.Stderr, "error writing output: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("wrote %s\n", outputPath)
}

func fetchAllReleases() ([]githubRelease, error) {
	ctx := context.Background()
	client := github.NewClient(nil)

	all := make([]githubRelease, 0, 256)
	opts := &github.ListOptions{PerPage: 100, Page: 1}

	for {
		repoReleases, resp, err := client.Repositories.ListReleases(ctx, "juju", "juju", opts)
		if err != nil {
			return nil, err
		}

		for _, release := range repoReleases {
			if release == nil {
				continue
			}
			all = append(all, githubRelease{
				TagName:     release.GetTagName(),
				HTMLURL:     release.GetHTMLURL(),
				Draft:       release.GetDraft(),
				Prerelease:  release.GetPrerelease(),
				PublishedAt: release.GetPublishedAt().Time,
			})
		}

		if resp == nil || resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return all, nil
}

func filterStableFinalReleases(releases []githubRelease) []githubRelease {
	minVersion := version.MustParse(minSupportedVersion)

	filtered := make([]githubRelease, 0, len(releases))
	for _, release := range releases {
		if release.Draft || release.Prerelease {
			continue
		}
		v, ok := stableVersion(release.TagName)
		if !ok {
			continue
		}
		if v.Compare(minVersion) < 0 {
			continue
		}
		if slices.Contains(blacklistedVersions, v) {
			continue
		}
		filtered = append(filtered, release)
	}
	return filtered
}

func sortReleases(releases []githubRelease) []outputRelease {
	type keyedRelease struct {
		Version version.Number
		Release outputRelease
	}

	keyed := make([]keyedRelease, 0, len(releases))
	for _, release := range releases {
		v, ok := stableVersion(release.TagName)
		if !ok {
			continue
		}
		keyed = append(keyed, keyedRelease{
			Version: v,
			Release: outputRelease{
				Version: v.String(),
				Date:    release.PublishedAt.UTC().Format("2006-01-02"),
				Link:    release.HTMLURL,
			},
		})
	}

	sort.Slice(keyed, func(i, j int) bool {
		return keyed[i].Version.Compare(keyed[j].Version) > 0
	})

	output := make([]outputRelease, len(keyed))
	for i := range keyed {
		output[i] = keyed[i].Release
	}

	return output
}

func stableVersion(tag string) (version.Number, bool) {
	tag = strings.TrimPrefix(strings.TrimSpace(tag), "v")
	v, err := version.Parse(tag)
	if err != nil {
		return version.Number{}, false
	}
	if v.Tag != "" {
		return version.Number{}, false
	}
	return v, true
}

func writeOutputJSON(path string, data []outputRelease) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	bytes = append(bytes, '\n')
	return os.WriteFile(path, bytes, 0644)
}
