package server

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
)

type spyExecutor struct {
	called    bool
	lastQuery string
	lastType  string
	lastCtx   context.Context
}

func (s *spyExecutor) ExecuteQueryCtx(ctx context.Context, query, queryType string) (string, string, error) {
	s.called = true
	s.lastCtx = ctx
	s.lastQuery = query
	s.lastType = queryType
	return "processed", "raw", nil
}

func TestQueriesEndpointInvokesExecutor(t *testing.T) {
	t.Parallel()

	type ctxKey struct{}

	executor := &spyExecutor{}
	logger := log.NewWithOptions(io.Discard, log.Options{})
	srv := New(executor, logger)

	req := httptest.NewRequest(http.MethodPost, "/queries", strings.NewReader(`{"type":"gremlin","query":"g.V()"}`))
	req = req.WithContext(context.WithValue(req.Context(), ctxKey{}, "http-request"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !executor.called {
		t.Fatalf("expected ExecuteQuery to be called")
	}
	if executor.lastQuery != "g.V()" {
		t.Fatalf("expected query 'g.V()', got %q", executor.lastQuery)
	}
	if executor.lastType != "gremlin" {
		t.Fatalf("expected type 'gremlin', got %q", executor.lastType)
	}
	if got := executor.lastCtx.Value(ctxKey{}); got != "http-request" {
		t.Fatalf("expected request context to propagate, got %v", got)
	}
}
