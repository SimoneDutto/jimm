// Copyright 2026 Canonical.

package rpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	stderrors "errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"sync/atomic"

	"github.com/gorilla/websocket"
	"github.com/juju/juju/core/network"
	"github.com/juju/names/v5"
	"github.com/juju/zaputil/zapctx"
	"go.uber.org/zap"

	"github.com/canonical/jimm/v3/internal/dbmodel"
	jimmerrors "github.com/canonical/jimm/v3/internal/errors"
)

// A Dialer is used to create client connections to an RPC URL.
type Dialer struct {
	// TLSConfig is used to configure TLS for the client connection.
	TLSConfig *tls.Config
}

type websocketDialer interface {
	DialWebsocket(context.Context, string, http.Header) (*websocket.Conn, error)
}

// Dial establishes a new client RPC connection to the given URL.
func (d Dialer) Dial(ctx context.Context, url string, headers http.Header) (*Client, error) {
	conn, err := d.DialWebsocket(ctx, url, headers)
	if err != nil {
		return nil, err
	}
	return NewClient(conn), nil
}

// DialWebsocket dials a url and returns a websocket.
func (d Dialer) DialWebsocket(ctx context.Context, url string, headers http.Header) (*websocket.Conn, error) {

	dialer := websocket.Dialer{
		TLSClientConfig: d.TLSConfig,
	}
	conn, resp, err := dialer.DialContext(ctx, url, headers)
	if err != nil {
		return nil, fmt.Errorf("basic dial failed: %w", err)
	}
	defer resp.Body.Close()
	return conn, nil
}

// GetAddressesAndTLSConfig returns the addresses and TLS configuration for the given controller.
func GetAddressesAndTLSConfig(ctx context.Context, ctl *dbmodel.Controller) ([]string, *tls.Config) {
	var tlsConfig *tls.Config
	if ctl.CACertificate != "" {
		cp := x509.NewCertPool()
		ok := cp.AppendCertsFromPEM([]byte(ctl.CACertificate))
		if !ok {
			zapctx.Warn(ctx, "no CA certificates added")
		}
		tlsConfig = &tls.Config{
			RootCAs:    cp,
			ServerName: ctl.TLSHostname,
			MinVersion: tls.VersionTLS12,
		}
	}
	var addrs []string
	if ctl.PublicAddress != "" {
		addrs = append(addrs, ctl.PublicAddress)
	}

	for _, hps := range ctl.Addresses {
		for _, hp := range hps {
			if maybeReachable(network.Scope(hp.Scope)) {
				var ip string
				if hp.Type == string(network.IPv6Address) {
					ip = fmt.Sprintf("[%s]:%d", hp.Value, hp.Port)
				} else {
					ip = fmt.Sprintf("%s:%d", hp.Value, hp.Port)
				}
				addrs = append(addrs, ip)
			}
		}
	}
	return addrs, tlsConfig
}

// Dial connects to the controller/model and returns a raw websocket
// that can be used as is.
// It accepts the endpoints to dial, normally /api or /commands.
func Dial(ctx context.Context, ctl *dbmodel.Controller, modelTag names.ModelTag, finalPath string, headers http.Header, attrs url.Values) (*websocket.Conn, error) {
	addrs, tlsConfig := GetAddressesAndTLSConfig(ctx, ctl)
	dialer := Dialer{
		TLSConfig: tlsConfig,
	}
	var websocketUrls []string
	for _, addr := range addrs {
		websocketUrls = append(websocketUrls, websocketURL(addr, modelTag, finalPath, attrs))
	}
	zapctx.Debug(ctx, "Dialling all URLs", zap.Any("urls", websocketUrls))
	conn, err := dialAll(ctx, &dialer, websocketUrls, headers)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// maybeReachable decides what kinds of links JIMM should try to connect via.
// Local IPs like localhost for example are excluded but public IPs and Cloud local IPs are potentially reachable.
func maybeReachable(scope network.Scope) bool {
	switch scope {
	case network.ScopeCloudLocal:
		return true
	case network.ScopePublic:
		return true
	case "":
		return true
	default:
		return false
	}
}

func websocketURL(s string, mt names.ModelTag, finalPath string, attrs url.Values) string {
	u := url.URL{
		Scheme: "wss",
		Host:   s,
	}
	if mt.Id() != "" {
		u.Path = path.Join(u.Path, "model", mt.Id())
	}
	if finalPath == "" {
		u.Path = path.Join(u.Path, "api")
	} else {
		u.Path = path.Join(u.Path, finalPath)
	}
	u.RawQuery = attrs.Encode()
	return u.String()
}

// dialAll simultaneously dials all given urls and returns the first
// connection.
func dialAll(ctx context.Context, dialer websocketDialer, urls []string, headers http.Header) (*websocket.Conn, error) {
	if len(urls) == 0 {
		return nil, jimmerrors.New("no urls to dial")
	}
	conn, err := dialAllHelper(ctx, dialer, urls, headers)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

type dialResult struct {
	url  string
	conn *websocket.Conn
	err  error
}

// dialAllHelper simultaneously dials all given urls and returns the first successful websocket connection.
func dialAllHelper(ctx context.Context, dialer websocketDialer, urls []string, headers http.Header) (*websocket.Conn, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	results := make(chan dialResult, len(urls))
	var winnerSelected atomic.Bool
	for _, url := range urls {
		zapctx.Info(ctx, "dialing", zap.String("url", url))
		url := url
		go func() {
			conn, err := dialer.DialWebsocket(ctx, url, headers)
			if err == nil {
				if !winnerSelected.CompareAndSwap(false, true) {
					if conn != nil {
						conn.Close()
					}
					return
				}
			}
			results <- dialResult{url: url, conn: conn, err: err}
		}()
	}

	var errs []error
	for i := 0; i < len(urls); i++ {
		result := <-results
		if result.err != nil {
			errs = append(errs, fmt.Errorf("dial %q: %w", result.url, result.err))
			continue
		}

		return result.conn, nil
	}
	return nil, stderrors.Join(errs...)
}
