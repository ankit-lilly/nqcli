package server

import (
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
}

func (s *spyExecutor) ExecuteQuery(query, queryType string) (string, string, error) {
	s.called = true
	s.lastQuery = query
	s.lastType = queryType
	return "processed", "raw", nil
}

func TestQueriesEndpointInvokesExecutor(t *testing.T) {
	t.Parallel()

	executor := &spyExecutor{}
	logger := log.NewWithOptions(io.Discard, log.Options{})
	srv := New(executor, logger)

	req := httptest.NewRequest(http.MethodPost, "/queries", strings.NewReader(`{"type":"gremlin","query":"g.V()"}`))
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
}
