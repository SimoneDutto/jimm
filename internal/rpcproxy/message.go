// Copyright 2025 Canonical.

package rpcproxy

import (
	"encoding/json"
	"time"

	"github.com/canonical/jimm/v3/internal/telemetry"
)

// A message encodes a single message sent, or received, over an RPC
// connection. It contains the union of fields in a request or response
// message.
type message struct {
	start      time.Time
	span       *telemetry.TrackedSpan `json:"-"`
	RequestID  uint64                 `json:"request-id,omitempty"`
	Type       string                 `json:"type,omitempty"`
	Version    int                    `json:"version,omitempty"`
	ID         string                 `json:"id,omitempty"`
	Request    string                 `json:"request,omitempty"`
	Params     json.RawMessage        `json:"params,omitempty"`
	Error      string                 `json:"error,omitempty"`
	ErrorCode  string                 `json:"error-code,omitempty"`
	ErrorInfo  map[string]any         `json:"error-info,omitempty"`
	Response   json.RawMessage        `json:"response,omitempty"`
	TraceID    string                 `json:"trace-id,omitempty"`
	SpanID     string                 `json:"span-id,omitempty"`
	TraceFlags int                    `json:"trace-flags,omitempty"`
}

// finishSpan safely finishes the message's trace span (if any) and records
// the supplied error.  Safe to call on a nil receiver.
func (m *message) finishSpan(err error) {
	if m == nil {
		return
	}
	if m.span != nil {
		m.span.Finish(err)
		m.span = nil
	}
}
