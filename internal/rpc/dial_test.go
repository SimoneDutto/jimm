// Copyright 2026 Canonical.

package rpc

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestDialAllHelperReturnsAfterFirstSuccessfulDial(t *testing.T) {
	fastURL := "wss://fast.example/api"
	slowURL := "wss://slow.example/api"
	slowCancelled := make(chan struct{}, 1)

	dialer := &fakeWebsocketDialer{
		responses: map[string]fakeDialResponse{
			fastURL: {conn: &websocket.Conn{}},
			slowURL: {
				delay:     200 * time.Millisecond,
				err:       errors.New("slow dial failed"),
				cancelled: slowCancelled,
			},
		},
	}

	conn, err := dialAllHelper(context.Background(), dialer, []string{slowURL, fastURL}, nil)

	if err != nil {
		t.Fatalf("dialAllHelper returned error: %v", err)
	}
	if conn == nil {
		t.Fatalf("expected winning connection")
	}

	select {
	case <-slowCancelled:
	case <-time.After(100 * time.Millisecond):
		t.Fatalf("expected slow loser dial to be cancelled")
	}
}

func TestDialAllHelperReturnsAllDialErrors(t *testing.T) {
	firstURL := "wss://first.example/api"
	secondURL := "wss://second.example/api"
	firstErr := errors.New("first dial failed")
	secondErr := errors.New("second dial failed")

	dialer := &fakeWebsocketDialer{
		responses: map[string]fakeDialResponse{
			firstURL:  {err: firstErr},
			secondURL: {err: secondErr},
		},
	}

	conn, err := dialAllHelper(context.Background(), dialer, []string{firstURL, secondURL}, nil)

	if conn != nil {
		t.Fatalf("expected no connection, got %v", conn)
	}
	if err == nil {
		t.Fatalf("expected error")
	}
	if !errors.Is(err, firstErr) {
		t.Fatalf("expected error to include first dial error, got %v", err)
	}
	if !errors.Is(err, secondErr) {
		t.Fatalf("expected error to include second dial error, got %v", err)
	}
	if !strings.Contains(err.Error(), firstURL) {
		t.Fatalf("expected error to include first URL, got %v", err)
	}
	if !strings.Contains(err.Error(), secondURL) {
		t.Fatalf("expected error to include second URL, got %v", err)
	}
}

type fakeWebsocketDialer struct {
	responses map[string]fakeDialResponse
}

type fakeDialResponse struct {
	conn      *websocket.Conn
	err       error
	delay     time.Duration
	cancelled chan<- struct{}
}

func (d *fakeWebsocketDialer) DialWebsocket(ctx context.Context, url string, _ http.Header) (*websocket.Conn, error) {
	response := d.responses[url]
	if response.delay > 0 {
		timer := time.NewTimer(response.delay)
		defer timer.Stop()

		select {
		case <-timer.C:
		case <-ctx.Done():
			if response.cancelled != nil {
				select {
				case response.cancelled <- struct{}{}:
				default:
				}
			}
			<-timer.C
		}
	}
	return response.conn, response.err
}
