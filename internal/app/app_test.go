package app

import (
	"context"
	"fmt"
	"testing"
)

type stubClient struct {
	response string
	err      error
	lastQ    string
	lastType string
	lastCtx  context.Context
}

func (s *stubClient) ExecuteQueryCtx(ctx context.Context, query, queryType string) (string, error) {
	s.lastCtx = ctx
	s.lastQ = query
	s.lastType = queryType
	return s.response, s.err
}

func newServiceWithStub(response string, err error) (*AppService, *stubClient) {
	stub := &stubClient{response: response, err: err}
	svc := NewAppService(stub)
	return svc, stub
}

func TestExecuteQuery_EmptyQuery(t *testing.T) {
	t.Parallel()
	svc, _ := newServiceWithStub("", nil)

	_, _, err := svc.ExecuteQuery("", "gremlin")
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

func TestExecuteQuery_WhitespaceOnly(t *testing.T) {
	t.Parallel()
	svc, _ := newServiceWithStub("", nil)

	_, _, err := svc.ExecuteQuery("   \t\n  ", "gremlin")
	if err == nil {
		t.Fatal("expected error for whitespace-only query")
	}
}

func TestExecuteQuery_NeptuneError(t *testing.T) {
	t.Parallel()
	svc, _ := newServiceWithStub("", fmt.Errorf("connection refused"))

	_, _, err := svc.ExecuteQuery("g.V()", "gremlin")
	if err == nil {
		t.Fatal("expected error when neptune fails")
	}
}

func TestExecuteQuery_SuccessfulResponseExtraction(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		response      string
		wantProcessed string
		wantRaw       string
	}{
		{
			name:          "full nested response with data.executeQuery containing JSON",
			response:      `{"data":{"executeQuery":"{\"data\":[{\"id\":\"1\",\"label\":\"Study\"}]}"}}`,
			wantProcessed: "[\n  {\n    \"id\": \"1\",\n    \"label\": \"Study\"\n  }\n]",
			wantRaw:       `{"data":{"executeQuery":"{\"data\":[{\"id\":\"1\",\"label\":\"Study\"}]}"}}`,
		},
		{
			name:          "executeQuery contains plain string (not JSON)",
			response:      `{"data":{"executeQuery":"hello world"}}`,
			wantProcessed: "hello world",
			wantRaw:       `{"data":{"executeQuery":"hello world"}}`,
		},
		{
			name:          "response missing data key",
			response:      `{"errors":[{"message":"bad query"}]}`,
			wantProcessed: `{"errors":[{"message":"bad query"}]}`,
			wantRaw:       `{"errors":[{"message":"bad query"}]}`,
		},
		{
			name:          "data present but no executeQuery",
			response:      `{"data":{"other":"value"}}`,
			wantProcessed: `{"data":{"other":"value"}}`,
			wantRaw:       `{"data":{"other":"value"}}`,
		},
		{
			name:          "executeQuery is null",
			response:      `{"data":{"executeQuery":null}}`,
			wantProcessed: `{"data":{"executeQuery":null}}`,
			wantRaw:       `{"data":{"executeQuery":null}}`,
		},
		{
			name:          "executeQuery contains JSON without data wrapper",
			response:      `{"data":{"executeQuery":"[1,2,3]"}}`,
			wantProcessed: "[\n  1,\n  2,\n  3\n]",
			wantRaw:       `{"data":{"executeQuery":"[1,2,3]"}}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc, _ := newServiceWithStub(tt.response, nil)

			processed, raw, err := svc.ExecuteQuery("g.V()", "gremlin")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if processed != tt.wantProcessed {
				t.Errorf("processed:\ngot:  %q\nwant: %q", processed, tt.wantProcessed)
			}
			if raw != tt.wantRaw {
				t.Errorf("raw:\ngot:  %q\nwant: %q", raw, tt.wantRaw)
			}
		})
	}
}

func TestExecuteQuery_InvalidJSON(t *testing.T) {
	t.Parallel()
	svc, _ := newServiceWithStub("not valid json", nil)

	_, _, err := svc.ExecuteQuery("g.V()", "gremlin")
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestExecuteQuery_PassesQueryTypeThrough(t *testing.T) {
	t.Parallel()
	svc, stub := newServiceWithStub(`{"data":{"executeQuery":"ok"}}`, nil)

	_, _, err := svc.ExecuteQuery("MATCH (n) RETURN n", "cypher")
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

func TestExecuteQueryCtx_PassesContextThrough(t *testing.T) {
	t.Parallel()

	type ctxKey struct{}

	svc, stub := newServiceWithStub(`{"data":{"executeQuery":"ok"}}`, nil)
	ctx := context.WithValue(context.Background(), ctxKey{}, "request-123")

	_, _, err := svc.ExecuteQueryCtx(ctx, "g.V()", "gremlin")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := stub.lastCtx.Value(ctxKey{}); got != "request-123" {
		t.Fatalf("expected context value to propagate, got %v", got)
	}
}
