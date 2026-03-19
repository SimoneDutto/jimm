# versioner

Queries the GitHub releases API for `juju/juju` and writes a flat, sorted JSON list of stable releases (from `3.6.5` onwards) to `internal/jimm/juju/juju-releases.json`.

That file is embedded at compile time via `//go:embed` in `internal/jimm/juju/supportedversions.go` and served by the `SupportedVersions` API endpoint.

## Usage

```
make generate-juju-releases
```

Run this whenever you want to refresh the release list. The output file must be committed — CI verifies it is up to date.

## No go.mod

This tool lives inside the main JIMM module and shares its dependencies. It only adds `go-github` (and its single transitive dep `go-querystring`), which are declared in the root `go.mod`.
