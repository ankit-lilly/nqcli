package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ankit-lilly/nqcli/internal/core"
	"github.com/charmbracelet/log"
)

type spyService struct {
	called    bool
	lastQuery string
	lastType  string
	lastOpts  core.QueryOpts
	lastCtx   context.Context
	result    core.QueryResult
	err       error
}

func (s *spyService) ExecuteQuery(ctx context.Context, query, queryType string, opts core.QueryOpts) (core.QueryResult, error) {
	s.called = true
	s.lastCtx = ctx
	s.lastQuery = query
	s.lastType = queryType
	s.lastOpts = opts
	if s.result.Content != "" || s.result.Raw != "" || s.err != nil {
		return s.result, s.err
	}
	return core.QueryResult{Content: `{"ok":true}`, Processed: "processed", Raw: "raw"}, nil
}

func TestExplorerServesIndexWithoutRedirect(t *testing.T) {
	t.Parallel()
	srv := New(&spyService{}, log.NewWithOptions(io.Discard, log.Options{}))

	for _, path := range []string{"/explorer/", "/explorer/schema/view"} {
		recorder := httptest.NewRecorder()
		srv.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))
		if recorder.Code != http.StatusOK {
			t.Errorf("GET %s status = %d, want %d", path, recorder.Code, http.StatusOK)
		}
		if location := recorder.Header().Get("Location"); location != "" {
			t.Errorf("GET %s redirected to %q", path, location)
		}
		if cache := recorder.Header().Get("Cache-Control"); cache != "no-cache" {
			t.Errorf("GET %s Cache-Control = %q, want no-cache", path, cache)
		}
	}
}

func TestExplorerAssetsAreImmutable(t *testing.T) {
	t.Parallel()
	srv := New(&spyService{}, log.NewWithOptions(io.Discard, log.Options{}))
	assets, err := fs.Glob(explorerDistFS, "webui_dist/assets/*")
	if err != nil || len(assets) == 0 {
		t.Fatalf("find embedded asset: %v", err)
	}
	recorder := httptest.NewRecorder()
	path := "/explorer/" + strings.TrimPrefix(assets[0], "webui_dist/")
	srv.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, path, nil))

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if cache := recorder.Header().Get("Cache-Control"); cache != "public, max-age=31536000, immutable" {
		t.Fatalf("Cache-Control = %q", cache)
	}
}

func TestQueriesEndpointInvokesService(t *testing.T) {
	t.Parallel()

	type ctxKey struct{}

	spy := &spyService{}
	logger := log.NewWithOptions(io.Discard, log.Options{})
	srv := New(spy, logger)

	req := httptest.NewRequest(http.MethodPost, "/queries", strings.NewReader(`{"type":"gremlin","query":"g.V()"}`))
	req = req.WithContext(context.WithValue(req.Context(), ctxKey{}, "http-request"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if !spy.called {
		t.Fatalf("expected ExecuteQuery to be called")
	}
	if spy.lastQuery != "g.V()" {
		t.Fatalf("expected query 'g.V()', got %q", spy.lastQuery)
	}
	if spy.lastType != "gremlin" {
		t.Fatalf("expected type 'gremlin', got %q", spy.lastType)
	}
	if got := spy.lastCtx.Value(ctxKey{}); got != "http-request" {
		t.Fatalf("expected request context to propagate, got %v", got)
	}
}

func TestQueriesEndpointPassesSerializerOpts(t *testing.T) {
	t.Parallel()

	spy := &spyService{}
	logger := log.NewWithOptions(io.Discard, log.Options{})
	srv := New(spy, logger)

	body := `{"type":"gremlin","query":"g.V()","serializer":"application/vnd.gremlin-v2.0+json"}`
	req := httptest.NewRequest(http.MethodPost, "/queries", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if spy.lastOpts.Serializer != "application/vnd.gremlin-v2.0+json" {
		t.Fatalf("expected serializer %q, got %q", "application/vnd.gremlin-v2.0+json", spy.lastOpts.Serializer)
	}
}

func TestGremlinEndpointWrapsInNeptuneEnvelope(t *testing.T) {
	t.Parallel()

	spy := &spyService{}
	logger := log.NewWithOptions(io.Discard, log.Options{})
	srv := New(spy, logger)

	req := httptest.NewRequest(http.MethodPost, "/gremlin", strings.NewReader(`{"query":"g.V().limit(1)"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var env struct {
		RequestID string `json:"requestId"`
		Status    struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"status"`
		Result struct {
			Data json.RawMessage `json:"data"`
		} `json:"result"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&env); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if env.RequestID == "" {
		t.Fatal("expected non-empty requestId")
	}
	if env.Status.Code != 200 {
		t.Fatalf("expected status code 200, got %d", env.Status.Code)
	}
	if env.Result.Data == nil {
		t.Fatal("expected non-nil result.data")
	}
}

func TestGremlinEndpointUsesGraphSONV3Serializer(t *testing.T) {
	t.Parallel()

	spy := &spyService{}
	logger := log.NewWithOptions(io.Discard, log.Options{})
	srv := New(spy, logger)

	req := httptest.NewRequest(http.MethodPost, "/gremlin", strings.NewReader(`{"query":"g.V()"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if spy.lastOpts.Serializer != "application/vnd.gremlin-v3.0+json" {
		t.Fatalf("expected serializer %q, got %q", "application/vnd.gremlin-v3.0+json", spy.lastOpts.Serializer)
	}
	if !spy.lastOpts.SkipFormatting {
		t.Fatal("expected SkipFormatting to be true")
	}
}

func TestGremlinEndpointReturnsHTTPErrorWhenQueryFails(t *testing.T) {
	t.Parallel()

	spy := &spyService{err: errors.New("bad traversal")}
	logger := log.NewWithOptions(io.Discard, log.Options{})
	srv := New(spy, logger)

	req := httptest.NewRequest(http.MethodPost, "/gremlin", strings.NewReader(`{"query":"g.invalid()"}`))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, rec.Code)
	}
	var response queryErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if response.DetailedMessage != "bad traversal" {
		t.Fatalf("expected query error message, got %q", response.DetailedMessage)
	}
}

func TestOpenCypherEndpointInvokesService(t *testing.T) {
	t.Parallel()

	spy := &spyService{result: core.QueryResult{Content: `{"data":{"results":[]},"meta":{"source":"api"}}`}}
	logger := log.NewWithOptions(io.Discard, log.Options{})
	srv := New(spy, logger)

	req := httptest.NewRequest(http.MethodPost, "/openCypher", strings.NewReader(`{"query":"MATCH (n) RETURN n"}`))
	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if spy.lastType != "openCypher" {
		t.Fatalf("expected type openCypher, got %q", spy.lastType)
	}
	if got := strings.TrimSpace(rec.Body.String()); got != `{"results":[]}` {
		t.Fatalf("unexpected response body %q", got)
	}
}

func TestLoggerEndpointReturns200(t *testing.T) {
	t.Parallel()

	spy := &spyService{}
	logger := log.NewWithOptions(io.Discard, log.Options{})
	srv := New(spy, logger)

	req := httptest.NewRequest(http.MethodPost, "/logger", nil)
	req.Header.Set("level", "info")
	req.Header.Set("message", "test log")
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestDefaultConnectionReturnsJSON(t *testing.T) {
	t.Parallel()

	spy := &spyService{}
	logger := log.NewWithOptions(io.Discard, log.Options{})
	srv := New(spy, logger)

	req := httptest.NewRequest(http.MethodGet, "/defaultConnection", nil)
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var config struct {
		APIBaseURL  string `json:"apiBaseUrl"`
		QueryEngine string `json:"queryEngine"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&config); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if config.APIBaseURL != "http://example.com" {
		t.Fatalf("expected API base URL %q, got %q", "http://example.com", config.APIBaseURL)
	}
	if config.QueryEngine != "gremlin" {
		t.Fatalf("expected query engine 'gremlin', got %q", config.QueryEngine)
	}
}

func TestDefaultConnectionReturnsConfiguredQueryEngine(t *testing.T) {
	t.Parallel()

	logger := log.NewWithOptions(io.Discard, log.Options{})
	srv := NewWithOptions(&spyService{}, logger, Options{QueryEngine: "openCypher"})
	rec := httptest.NewRecorder()

	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/defaultConnection", nil))

	var config struct {
		QueryEngine string `json:"queryEngine"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&config); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if config.QueryEngine != "openCypher" {
		t.Fatalf("expected query engine 'openCypher', got %q", config.QueryEngine)
	}
}
