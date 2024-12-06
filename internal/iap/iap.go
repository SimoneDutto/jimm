// Copyright 2024 Canonical.
package iap

import (
	"net/http"

	"github.com/canonical/jimm/v3/internal/jimm"
)

func Handler(j jimm.JIMM) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		endpoint := j.GetUnitEndpoint(ctx, "14daf0de-407a-4ffd-8219-09ee9a2d32c9", "postgresql/0")
		w.Write([]byte(endpoint))
	})
}
