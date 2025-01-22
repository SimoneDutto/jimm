// Copyright 2025 Canonical.
package jimmhttp

import (
	"encoding/json"
	"net/http"

	"github.com/juju/zaputil/zapctx"
	"go.uber.org/zap"
)

// WriteFingerprints writes a map as JSON to the response.
func WriteFingerprints(m map[string]string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		b, err := json.Marshal(m)
		if err != nil {
			zapctx.Error(ctx, "failed to marshal map", zap.Error(err))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, err = w.Write(b)
		if err != nil {
			zapctx.Error(ctx, "failed to write response", zap.Error(err))
		}
	}
}
