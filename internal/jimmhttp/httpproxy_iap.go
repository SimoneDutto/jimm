// Copyright 2024 Canonical.

package jimmhttp

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/go-chi/chi/v5"

	"github.com/canonical/jimm/v3/internal/jimm"
	"github.com/canonical/jimm/v3/internal/middleware"
	"github.com/canonical/jimm/v3/internal/rpc"
)

// IAPProxyHandler is an handler that provides proxying capabilities.
// It uses the uuid in the path to proxy requests to model's controller.
type IAPProxyHandler struct {
	Router *chi.Mux
	jimm   *jimm.JIMM
}

// NewIAPProxyHandler creates a proxy http handler.
func NewIAPProxyHandler(jimm *jimm.JIMM) *IAPProxyHandler {
	return &IAPProxyHandler{Router: chi.NewRouter(), jimm: jimm}
}

// Routes returns the grouped routers routes with group specific middlewares.
func (hph *IAPProxyHandler) Routes() chi.Router {
	hph.SetupMiddleware()
	hph.Router.HandleFunc("/", hph.ProxyHTTP)
	return hph.Router
}

// SetupMiddleware applies authn and authz middlewares.
func (hph *IAPProxyHandler) SetupMiddleware() {
	hph.Router.Use(func(h http.Handler) http.Handler {
		return middleware.AuthenticateViaCookie(h, hph.jimm)
	})
}

// ProxyHTTP extracts the model uuid from the path to proxy the request to the right controller.
func (hph *IAPProxyHandler) ProxyHTTP(w http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	endpoint := hph.jimm.GetUnitEndpoint(ctx, "14daf0de-407a-4ffd-8219-09ee9a2d32c9", "nginx/0")
	url := url.URL{}
	url.Scheme = "http"
	url.Host = fmt.Sprintf("%s:%d", endpoint, 80)
	err := rpc.DoRequest(ctx, w, req, rpc.HttpOptions{
		TLSConfig: nil,
		URL:       url,
	})
	if err != nil {
		writeError(ctx, w, http.StatusGatewayTimeout, err, "Gateway timeout")
	}
}
