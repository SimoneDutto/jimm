// Copyright 2025 Canonical.

package jujuclistore

import (
	"bytes"
	"context"
	"html/template"
	"io"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/juju/clock"
	jujuerrrors "github.com/juju/errors"
	"github.com/juju/retry"
	"github.com/juju/zaputil/zapctx"
	"go.uber.org/zap"
)

// launchPadURL is the base URL for the Juju binary downloads.
const launchPadURL = "https://launchpad.net/juju"

// launchPadTemplate is the template for constructing the download URL for Juju binaries.
const launchPadTemplate = "{{.BaseURL}}/{{.VersionWithMinor}}/{{.VersionWithPatch}}/+download/juju-{{.VersionWithPatch}}-{{.Os}}-{{.Arch}}.tar.xz"

var retryRequestError = jujuerrrors.New("retry request error")

// JujuCLIStoreConfig holds the configuration for the Juju binary fetcher.
type JujuCLIStoreConfig struct {
	BaseURL string // Base URL for the Juju binary downloads. Example: "https://launchpad.net/juju"
	Dir     string // Directory to store the downloaded binaries. Defaults to the system's temp directory.

	MaxEntries int // Maximum number of entries to keep in the directory. Defaults to 2.
}

// JujuBinarySpec defines the specifications for a Juju binary to be downloaded.
type JujuBinarySpec struct {
	VersionWithMinor string // Version with minor version number, e.g., "3.6"
	VersionWithPatch string // Version with patch version number, e.g., "3.6.2"
	Os               string // Operating system, e.g., "linux"
	Arch             string // Architecture, e.g., "amd64"
}

type jujuCLIStore struct {
	config   JujuCLIStoreConfig
	template template.Template

	entries map[string]*Binary // Map to keep track of downloaded entries
	lock    sync.Mutex         // Protects the entries map
}

// NewJujuCLIStore creates a new jujuCLIFetcher instance with the provided configuration.
// If the BaseURL is not provided, it defaults to the launchpad URL.
func NewJujuCLIStore(cfg JujuCLIStoreConfig) (*jujuCLIStore, error) {
	if cfg.BaseURL == "" {
		// Default to the launchpad URL if no base URL is provided.
		cfg.BaseURL = launchPadURL
	}
	tmpl, err := template.New("URL").Parse(launchPadTemplate)
	if err != nil {
		return nil, err
	}
	if cfg.MaxEntries <= 0 {
		// Default to 2 entries if MaxEntries is not set or is less than or
		// equal to zero.
		cfg.MaxEntries = 2
	}
	return &jujuCLIStore{
		config:   cfg,
		template: *tmpl,
		entries:  make(map[string]*Binary, cfg.MaxEntries),
		lock:     sync.Mutex{},
	}, nil
}

// Binary represents a downloaded Juju binary.
// It contains the full path to the binary file.
// It also provides a method to mark the binary as done, which can be used to indicate that the binary
// is no longer used.
type Binary struct {
	FullPath string

	done atomic.Bool
}

func (b *Binary) Done() {
	b.done.Store(true)
}

// Get downloads the Juju binary specified by the JujuBinarySpec.
// It returns a Binary instance containing the full path to the downloaded binary.
// If the download fails, it returns an error.
// It retries the download on server errors or rate limiting.
// The retry logic uses exponential backoff.
// The context can be used to cancel the operation.
func (p *jujuCLIStore) Get(ctx context.Context, spec JujuBinarySpec) (*Binary, error) {
	var buf bytes.Buffer
	err := p.template.Execute(&buf, map[string]string{
		"BaseURL":          p.config.BaseURL,
		"VersionWithMinor": spec.VersionWithMinor,
		"VersionWithPatch": spec.VersionWithPatch,
		"Os":               spec.Os,
		"Arch":             spec.Arch,
	})
	if err != nil {
		return nil, err
	}
	url := buf.String()
	p.lock.Lock()
	defer p.lock.Unlock()
	binary, ok := p.entries[url]
	if ok {
		return binary, nil
	}
	err = p.freeSpace(ctx)
	if err != nil {
		return nil, err
	}
	var file *os.File
	err = retry.Call(retry.CallArgs{
		Func: func() error {
			file, err = p.downloadFile(ctx, url)
			if err != nil {
				return err
			}
			return nil
		},
		IsFatalError: func(err error) bool {
			return !jujuerrrors.Is(err, retryRequestError)
		},
		BackoffFunc: retry.DoubleDelay,
		Attempts:    10,
		Delay:       time.Second,
		Clock:       clock.WallClock,
		Stop:        ctx.Done(),
	})
	if err != nil {
		return nil, err
	}
	binary = &Binary{
		FullPath: file.Name(),
	}
	p.entries[url] = binary
	return binary, nil
}

// freeSpace checks if the entries map has reached the maximum number of entries.
// If it has, it deletes a random entry from the map.
// The randomness is guaranteed by the fact iterating over a map in Go does not guarantee order.
func (p *jujuCLIStore) freeSpace(ctx context.Context) error {
	if len(p.entries) < p.config.MaxEntries {
		return nil
	}
	deletion := false
	for key, binary := range p.entries {
		if binary.done.Load() {
			err := os.Remove(binary.FullPath)
			if err != nil {
				zapctx.Error(ctx, "failed to remove binary", zap.Error(err), zap.String("path", binary.FullPath))
			} else {
				delete(p.entries, key)
				deletion = true
				break
			}
		}
	}
	if !deletion {
		// This shouldn't happen because we bootstrap a controller at the time, so there should always be at least one
		// binary that can be deleted.
		return jujuerrrors.Errorf("no entries to delete, max entries limit reached: %d", p.config.MaxEntries)
	}
	return nil
}

// downloadFile downloads the file from the specified URL and returns a file handle.
// It retries on server errors or rate limiting.
// If the download fails, it returns an error.
// The file is created in the directory specified in the configuration.
// The context can be used to cancel the operation.
func (p *jujuCLIStore) downloadFile(ctx context.Context, downloadUrl string) (*os.File, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadUrl, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if (resp.StatusCode >= 500 && resp.StatusCode < 600) || resp.StatusCode == http.StatusTooManyRequests {
		return nil, retryRequestError
	}
	if resp.StatusCode != http.StatusOK {
		return nil, jujuerrrors.Errorf("request failed with status %d", resp.StatusCode)
	}
	file, err := os.CreateTemp(p.config.Dir, "juju-*")
	if err != nil {
		return nil, err
	}
	_, err = io.Copy(file, resp.Body)
	if err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, err
	}
	return file, nil
}
