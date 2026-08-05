package cmd

import (
	"context"
	"os"
	"testing"
)

type spyQueryService struct {
	executeCtxCalls      int
	executeQueryCtxCalls int
	lastQuery            string
	lastQueryType        string
	lastCtx              context.Context
}

func (s *spyQueryService) ExecuteCtx(ctx context.Context, path, queryType string) (string, string, error) {
	s.executeCtxCalls++
	s.lastCtx = ctx
	s.lastQuery = path
	s.lastQueryType = queryType
	return "{}", "", nil
}

func (s *spyQueryService) ExecuteQueryCtx(ctx context.Context, query, queryType string) (string, string, error) {
	s.executeQueryCtxCalls++
	s.lastCtx = ctx
	s.lastQuery = query
	s.lastQueryType = queryType
	return "{}", "", nil
}

func TestRootCommandInlineQueryCallsExecuteQueryCtx(t *testing.T) {
	type ctxKey struct{}

	spy := &spyQueryService{}
	origFactory := newQueryService
	newQueryService = func(ctx context.Context) (queryService, error) { return spy, nil }
	t.Cleanup(func() {
		newQueryService = origFactory
		rootCmd.SetArgs(nil)
		rootCmd.SetContext(context.Background())
	})

	rootCmd.SetContext(context.WithValue(context.Background(), ctxKey{}, "inline-query"))
	rootCmd.SetArgs([]string{"--type", "cypher", "g.V()"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() returned error: %v", err)
	}

	if spy.executeQueryCtxCalls != 1 {
		t.Fatalf("expected ExecuteQueryCtx to be called once, got %d", spy.executeQueryCtxCalls)
	}
	if spy.executeCtxCalls != 0 {
		t.Fatalf("expected ExecuteCtx not to be called, got %d", spy.executeCtxCalls)
	}
	if spy.lastQuery != "g.V()" {
		t.Fatalf("expected query 'g.V()', got %q", spy.lastQuery)
	}
	if spy.lastQueryType != "cypher" {
		t.Fatalf("expected query type 'cypher', got %q", spy.lastQueryType)
	}
	if got := spy.lastCtx.Value(ctxKey{}); got != "inline-query" {
		t.Fatalf("expected command context to propagate, got %v", got)
	}
}

func TestRootCommandFileArgumentCallsExecuteCtx(t *testing.T) {
	type ctxKey struct{}

	spy := &spyQueryService{}
	origFactory := newQueryService
	newQueryService = func(ctx context.Context) (queryService, error) { return spy, nil }
	t.Cleanup(func() {
		newQueryService = origFactory
		rootCmd.SetArgs(nil)
		rootCmd.SetContext(context.Background())
	})

	tmpFile, err := os.CreateTemp(t.TempDir(), "query-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()

	rootCmd.SetContext(context.WithValue(context.Background(), ctxKey{}, "query-file"))
	rootCmd.SetArgs([]string{"--type", "gremlin", tmpFile.Name()})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() returned error: %v", err)
	}

	if spy.executeCtxCalls != 1 {
		t.Fatalf("expected ExecuteCtx to be called once, got %d", spy.executeCtxCalls)
	}
	if spy.executeQueryCtxCalls != 0 {
		t.Fatalf("expected ExecuteQueryCtx not to be called, got %d", spy.executeQueryCtxCalls)
	}
	if spy.lastQuery != tmpFile.Name() {
		t.Fatalf("expected query path %q, got %q", tmpFile.Name(), spy.lastQuery)
	}
	if spy.lastQueryType != "gremlin" {
		t.Fatalf("expected query type 'gremlin', got %q", spy.lastQueryType)
	}
	if got := spy.lastCtx.Value(ctxKey{}); got != "query-file" {
		t.Fatalf("expected command context to propagate, got %v", got)
	}
}
