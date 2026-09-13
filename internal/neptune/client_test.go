package neptune

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/ankit-lilly/nqcli/internal/config"
	"github.com/ankit-lilly/nqcli/internal/core"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

func TestExecuteREST_ExtractsDataEnvelope(t *testing.T) {
	t.Parallel()

	const rawResponse = `{"data":[{"id":"1","label":"Study"}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(rawResponse))
	}))
	defer server.Close()

	client, err := NewClient(
		&config.Config{URL: server.URL},
		aws.Config{
			Region:      "us-east-1",
			Credentials: credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
		},
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	result, err := client.ExecuteQuery(context.Background(), "g.V()", "gremlin", core.QueryOpts{})
	if err != nil {
		t.Fatalf("executeREST: %v", err)
	}
	if result.Content != `[{"id":"1","label":"Study"}]` {
		t.Fatalf("expected unwrapped REST payload, got %q", result.Content)
	}
	if result.Raw != rawResponse {
		t.Fatalf("expected raw response to be preserved, got %q", result.Raw)
	}
}

func TestExecuteREST_SetsAcceptHeaderWhenSerializerProvided(t *testing.T) {
	t.Parallel()

	var receivedAccept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	// Force REST backend by using an execute-api hostname workaround:
	// httptest uses localhost, so detectBackend won't pick REST.
	// We test executeREST directly instead.
	client, err := NewClient(
		&config.Config{URL: server.URL},
		aws.Config{
			Region:      "us-east-1",
			Credentials: credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
		},
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	opts := core.QueryOpts{Serializer: "application/vnd.gremlin-v2.0+json"}
	_, err = client.ExecuteQuery(context.Background(), "g.V()", "gremlin", opts)
	if err != nil {
		t.Fatalf("executeREST: %v", err)
	}
	if receivedAccept != opts.Serializer {
		t.Fatalf("expected Accept=%q, got %q", opts.Serializer, receivedAccept)
	}
}

func TestExecuteREST_NoAcceptHeaderWhenSerializerEmpty(t *testing.T) {
	t.Parallel()

	var receivedAccept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	client, err := NewClient(
		&config.Config{URL: server.URL},
		aws.Config{
			Region:      "us-east-1",
			Credentials: credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
		},
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	_, err = client.ExecuteQuery(context.Background(), "g.V()", "gremlin", core.QueryOpts{})
	if err != nil {
		t.Fatalf("executeREST: %v", err)
	}
	if receivedAccept != "" {
		t.Fatalf("expected no Accept header, got %q", receivedAccept)
	}
}

func TestExecuteQuery_PrefersDirectGremlin(t *testing.T) {
	t.Parallel()

	var directCalled bool
	var receivedAccept string
	directServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		directCalled = true
		receivedAccept = r.Header.Get("Accept")
		if r.URL.Path != "/gremlin" {
			t.Fatalf("expected direct /gremlin path, got %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requestId":"1","status":{"code":200},"result":{"data":{"@type":"g:List","@value":[]}}}`))
	}))
	defer directServer.Close()

	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("REST fallback should not be called when direct succeeds")
	}))
	defer restServer.Close()

	client, err := NewClient(
		&config.Config{URL: restServer.URL, DirectURL: directServer.URL},
		aws.Config{
			Region:      "us-east-1",
			Credentials: credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
		},
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	result, err := client.ExecuteQuery(context.Background(), "g.V()", "gremlin", core.QueryOpts{})
	if err != nil {
		t.Fatalf("ExecuteQuery: %v", err)
	}
	if !directCalled {
		t.Fatal("expected direct Neptune endpoint to be called")
	}
	if result.Content != `[]` {
		t.Fatalf("expected direct result.data to be normalized to plain JSON, got %q", result.Content)
	}
	if receivedAccept != defaultGremlinAccept {
		t.Fatalf("expected default direct Accept=%q, got %q", defaultGremlinAccept, receivedAccept)
	}
}

func TestExecuteQuery_PassesExplicitSerializerToDirectGremlin(t *testing.T) {
	t.Parallel()

	var receivedAccept string
	directServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAccept = r.Header.Get("Accept")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"requestId":"1","status":{"code":200},"result":{"data":{"@type":"g:List","@value":[]}}}`))
	}))
	defer directServer.Close()

	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("REST fallback should not be called when direct succeeds")
	}))
	defer restServer.Close()

	client, err := NewClient(
		&config.Config{URL: restServer.URL, DirectURL: directServer.URL},
		aws.Config{
			Region:      "us-east-1",
			Credentials: credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
		},
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	opts := core.QueryOpts{Serializer: "application/vnd.gremlin-v3.0+json"}
	result, err := client.ExecuteQuery(context.Background(), "g.V()", "gremlin", opts)
	if err != nil {
		t.Fatalf("ExecuteQuery: %v", err)
	}
	if receivedAccept != opts.Serializer {
		t.Fatalf("expected Accept=%q, got %q", opts.Serializer, receivedAccept)
	}
	if result.Content != `{"@type":"g:List","@value":[]}` {
		t.Fatalf("expected explicit serializer to preserve GraphSON, got %q", result.Content)
	}
}

func TestExecuteQuery_FallsBackToRESTWhenDirectGremlinFails(t *testing.T) {
	t.Parallel()

	var directCalls atomic.Int32
	directServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		directCalls.Add(1)
		http.Error(w, "direct unavailable", http.StatusBadGateway)
	}))
	defer directServer.Close()

	var restCalls atomic.Int32
	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		restCalls.Add(1)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"@type":"g:List","@value":[]}}`))
	}))
	defer restServer.Close()

	client, err := NewClient(
		&config.Config{URL: restServer.URL, DirectURL: directServer.URL},
		aws.Config{
			Region:      "us-east-1",
			Credentials: credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
		},
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	result, err := client.ExecuteQuery(context.Background(), "g.V()", "gremlin", core.QueryOpts{})
	if err != nil {
		t.Fatalf("ExecuteQuery: %v", err)
	}
	if directCalls.Load() != 1 {
		t.Fatalf("expected one direct Neptune probe, got %d", directCalls.Load())
	}
	if restCalls.Load() != 1 {
		t.Fatalf("expected REST fallback to be called once, got %d", restCalls.Load())
	}
	if result.Content != `{"@type":"g:List","@value":[]}` {
		t.Fatalf("expected REST result.data to be normalized, got %q", result.Content)
	}

	if _, err := client.ExecuteQuery(context.Background(), "g.V()", "gremlin", core.QueryOpts{}); err != nil {
		t.Fatalf("second ExecuteQuery: %v", err)
	}
	if directCalls.Load() != 1 {
		t.Fatalf("expected failed direct endpoint to be bypassed, got %d calls", directCalls.Load())
	}
	if restCalls.Load() != 2 {
		t.Fatalf("expected two REST calls, got %d", restCalls.Load())
	}
}

func TestExecuteQuery_DoesNotFallbackAfterDirectGremlinSucceeds(t *testing.T) {
	t.Parallel()

	var directCalls atomic.Int32
	directServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if directCalls.Add(1) == 1 {
			_, _ = w.Write([]byte(`{"result":{"data":{"@type":"g:List","@value":[]}}}`))
			return
		}
		http.Error(w, "invalid traversal", http.StatusBadRequest)
	}))
	defer directServer.Close()

	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("REST fallback should not run after the direct backend is selected")
	}))
	defer restServer.Close()

	client, err := NewClient(
		&config.Config{URL: restServer.URL, DirectURL: directServer.URL},
		aws.Config{
			Region:      "us-east-1",
			Credentials: credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
		},
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	if _, err := client.ExecuteQuery(context.Background(), "g.V()", "gremlin", core.QueryOpts{}); err != nil {
		t.Fatalf("first ExecuteQuery: %v", err)
	}
	if _, err := client.ExecuteQuery(context.Background(), "g.invalid()", "gremlin", core.QueryOpts{}); err == nil {
		t.Fatal("second ExecuteQuery expected direct query error")
	}
}

func TestExecuteQuery_ConvertsDirectGraphSONToPlainJSONWhenSerializerEmpty(t *testing.T) {
	t.Parallel()

	const graphson = `{
		"requestId": "1",
		"status": {"code": 200},
		"result": {
			"data": {
				"@type": "g:List",
				"@value": [{
					"@type": "g:Map",
					"@value": [
						{"@type": "g:T", "@value": "id"},
						"v1",
						{"@type": "g:T", "@value": "label"},
						"Study",
						"count",
						{"@type": "g:Int64", "@value": 42},
						"properties",
						{"@type": "g:Map", "@value": [
							"name",
							{"@type": "g:VertexProperty", "@value": {"value": "ABC"}}
						]}
					]
				}]
			}
		}
	}`

	directServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(graphson))
	}))
	defer directServer.Close()

	restServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("REST fallback should not be called when direct succeeds")
	}))
	defer restServer.Close()

	client, err := NewClient(
		&config.Config{URL: restServer.URL, DirectURL: directServer.URL},
		aws.Config{
			Region:      "us-east-1",
			Credentials: credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
		},
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	result, err := client.ExecuteQuery(context.Background(), "g.V()", "gremlin", core.QueryOpts{})
	if err != nil {
		t.Fatalf("ExecuteQuery: %v", err)
	}

	want := `[{"count":42,"id":"v1","label":"Study","properties":{"name":"ABC"}}]`
	if result.Content != want {
		t.Fatalf("expected plain JSON %s, got %s", want, result.Content)
	}
}

func TestReadResponseLimit(t *testing.T) {
	for _, tc := range []struct {
		text      string
		limit     int64
		wantError bool
	}{{"1234", 4, false}, {"12345", 4, true}, {"12345", 0, false}} {
		b, err := readResponse(strings.NewReader(tc.text), tc.limit)
		if (err != nil) != tc.wantError {
			t.Fatalf("limit %d: %v", tc.limit, err)
		}
		if err == nil && string(b) != tc.text {
			t.Fatal("response changed")
		}
	}
}
