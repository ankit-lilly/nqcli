package core

import (
	"context"
	"fmt"
	"testing"
)

type stubClient struct {
	response QueryPayload
	err      error
	lastQ    string
	lastType string
	lastOpts QueryOpts
	lastCtx  context.Context
}

func (s *stubClient) ExecuteQuery(ctx context.Context, query, queryType string, opts QueryOpts) (QueryPayload, error) {
	s.lastCtx = ctx
	s.lastQ = query
	s.lastType = queryType
	s.lastOpts = opts
	return s.response, s.err
}

func newServiceWithStub(response QueryPayload, err error) (*Service, *stubClient) {
	stub := &stubClient{response: response, err: err}
	svc := NewService(stub)
	return svc, stub
}

func TestExecuteQuery_EmptyQuery(t *testing.T) {
	t.Parallel()
	svc, _ := newServiceWithStub(QueryPayload{}, nil)

	_, err := svc.ExecuteQuery(context.Background(), "", "gremlin", QueryOpts{})
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestExecuteQuery_WhitespaceOnly(t *testing.T) {
	t.Parallel()
	svc, _ := newServiceWithStub(QueryPayload{}, nil)

	_, err := svc.ExecuteQuery(context.Background(), "   \t\n  ", "gremlin", QueryOpts{})
	if err == nil {
		t.Fatal("expected error for whitespace-only query")
	}
}

func TestExecuteQuery_NeptuneError(t *testing.T) {
	t.Parallel()
	svc, _ := newServiceWithStub(QueryPayload{Raw: `{"errors":["boom"]}`}, fmt.Errorf("connection refused"))

	result, err := svc.ExecuteQuery(context.Background(), "g.V()", "gremlin", QueryOpts{})
	if err == nil {
		t.Fatal("expected error when neptune fails")
	}
	if result.Raw != `{"errors":["boom"]}` {
		t.Fatalf("expected raw error payload to be preserved, got %q", result.Raw)
	}
}

func TestExecuteQuery_PassesOptsThrough(t *testing.T) {
	t.Parallel()
	svc, stub := newServiceWithStub(QueryPayload{Content: `[]`, Raw: `{"data":[]}`}, nil)

	opts := QueryOpts{Serializer: "application/vnd.gremlin-v2.0+json"}
	_, err := svc.ExecuteQuery(context.Background(), "g.V()", "gremlin", opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.lastOpts.Serializer != opts.Serializer {
		t.Fatalf("expected serializer %q, got %q", opts.Serializer, stub.lastOpts.Serializer)
	}
}

func TestExecuteQuery_SuccessfulResponseFormatting(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		response      QueryPayload
		wantProcessed string
		wantContent   string
		wantRaw       string
	}{
		{
			name:          "formats JSON object",
			response:      QueryPayload{Content: `{"id":"1","label":"Study"}`, Raw: `{"data":{"executeQuery":"{\"id\":\"1\",\"label\":\"Study\"}"}}`},
			wantProcessed: "{\n  \"id\": \"1\",\n  \"label\": \"Study\"\n}",
			wantContent:   `{"id":"1","label":"Study"}`,
			wantRaw:       `{"data":{"executeQuery":"{\"id\":\"1\",\"label\":\"Study\"}"}}`,
		},
		{
			name:          "formats JSON array",
			response:      QueryPayload{Content: `[{"id":"1","label":"Study"}]`, Raw: `{"data":[{"id":"1","label":"Study"}]}`},
			wantProcessed: "[\n  {\n    \"id\": \"1\",\n    \"label\": \"Study\"\n  }\n]",
			wantContent:   `[{"id":"1","label":"Study"}]`,
			wantRaw:       `{"data":[{"id":"1","label":"Study"}]}`,
		},
		{
			name:          "passes through plain text",
			response:      QueryPayload{Content: "hello world", Raw: `{"data":{"executeQuery":"hello world"}}`},
			wantProcessed: "hello world",
			wantContent:   "hello world",
			wantRaw:       `{"data":{"executeQuery":"hello world"}}`,
		},
		{
			name:          "preserves full JSON response when payload is an object",
			response:      QueryPayload{Content: `{"errors":[{"message":"bad query"}]}`, Raw: `{"errors":[{"message":"bad query"}]}`},
			wantProcessed: "{\n  \"errors\": [\n    {\n      \"message\": \"bad query\"\n    }\n  ]\n}",
			wantContent:   `{"errors":[{"message":"bad query"}]}`,
			wantRaw:       `{"errors":[{"message":"bad query"}]}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _ := newServiceWithStub(tt.response, nil)

			result, err := svc.ExecuteQuery(context.Background(), "g.V()", "gremlin", QueryOpts{})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if result.Processed != tt.wantProcessed {
				t.Errorf("processed:\ngot:  %q\nwant: %q", result.Processed, tt.wantProcessed)
			}
			if result.Content != tt.wantContent {
				t.Errorf("content:\ngot:  %q\nwant: %q", result.Content, tt.wantContent)
			}
			if result.Raw != tt.wantRaw {
				t.Errorf("raw:\ngot:  %q\nwant: %q", result.Raw, tt.wantRaw)
			}
		})
	}
}

func TestExecuteQuery_InvalidJSON(t *testing.T) {
	t.Parallel()
	svc, _ := newServiceWithStub(QueryPayload{Content: "not valid json", Raw: "not valid json"}, nil)

	result, err := svc.ExecuteQuery(context.Background(), "g.V()", "gremlin", QueryOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != "not valid json" {
		t.Errorf("expected content passthrough, got %q", result.Content)
	}
	if result.Processed != "not valid json" {
		t.Errorf("expected raw passthrough, got %q", result.Processed)
	}
	if result.Raw != "not valid json" {
		t.Errorf("expected raw to match response, got %q", result.Raw)
	}
}

func TestExecuteQuery_FallsBackToRawWhenContentEmpty(t *testing.T) {
	t.Parallel()
	svc, _ := newServiceWithStub(QueryPayload{Raw: `{"data":{"keep":true}}`}, nil)

	result, err := svc.ExecuteQuery(context.Background(), "g.V()", "gremlin", QueryOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Content != `{"data":{"keep":true}}` {
		t.Fatalf("expected raw fallback content, got %q", result.Content)
	}
	if result.Processed != "{\n  \"data\": {\n    \"keep\": true\n  }\n}" {
		t.Fatalf("expected formatted raw fallback, got %q", result.Processed)
	}
}

func TestExecuteQuery_PassesQueryTypeThrough(t *testing.T) {
	t.Parallel()
	svc, stub := newServiceWithStub(QueryPayload{Content: "ok", Raw: `{"data":{"executeQuery":"ok"}}`}, nil)

	_, err := svc.ExecuteQuery(context.Background(), "MATCH (n) RETURN n", "cypher", QueryOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.lastType != "cypher" {
		t.Fatalf("expected query type 'cypher', got %q", stub.lastType)
	}
	if stub.lastQ != "MATCH (n) RETURN n" {
		t.Fatalf("expected query to be passed through, got %q", stub.lastQ)
	}
}

func TestExecuteQuery_PassesContextThrough(t *testing.T) {
	t.Parallel()

	type ctxKey struct{}

	svc, stub := newServiceWithStub(QueryPayload{Content: "ok", Raw: `{"data":{"executeQuery":"ok"}}`}, nil)
	ctx := context.WithValue(context.Background(), ctxKey{}, "request-123")

	_, err := svc.ExecuteQuery(ctx, "g.V()", "gremlin", QueryOpts{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := stub.lastCtx.Value(ctxKey{}); got != "request-123" {
		t.Fatalf("expected context value to propagate, got %v", got)
	}
}

func TestSkipFormattingPreservesContentAndDefaultPrettyOutput(t *testing.T) {
	svc, _ := newServiceWithStub(QueryPayload{Content: `{"value":1}`, Raw: `{"data":{"value":1}}`}, nil)
	pretty, err := svc.ExecuteQuery(context.Background(), "g.V()", "gremlin", QueryOpts{})
	if err != nil || pretty.Processed != "{\n  \"value\": 1\n}" {
		t.Fatalf("default output changed: %+v, %v", pretty, err)
	}
	compact, err := svc.ExecuteQuery(context.Background(), "g.V()", "gremlin", QueryOpts{SkipFormatting: true})
	if err != nil || compact.Processed != "" || compact.Content != pretty.Content || compact.Raw != pretty.Raw {
		t.Fatalf("compact output: %+v, %v", compact, err)
	}
}
