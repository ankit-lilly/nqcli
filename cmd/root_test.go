package cmd

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/ankit-lilly/nqcli/internal/core"
)

type spyQueryService struct {
	calls         int
	lastQuery     string
	lastQueryType string
	lastOpts      core.QueryOpts
	lastCtx       context.Context
}

func (s *spyQueryService) ExecuteQuery(ctx context.Context, query, queryType string, opts core.QueryOpts) (core.QueryResult, error) {
	s.calls++
	s.lastCtx = ctx
	s.lastQuery = query
	s.lastQueryType = queryType
	s.lastOpts = opts
	return core.QueryResult{Content: "{}", Processed: "{}", Raw: "{}"}, nil
}

func TestReadQueryAcceptsLongInlineTraversal(t *testing.T) {
	t.Parallel()
	query := "g.V().hasLabel(" + strings.Repeat(`"Study",`, 1000) + `"Study")`
	got, err := readQuery([]string{query})
	if err != nil {
		t.Fatal(err)
	}
	if got != query {
		t.Fatal("long inline traversal changed")
	}
}

func TestRootCommandInlineQuery(t *testing.T) {
	type ctxKey struct{}

	spy := &spyQueryService{}
	origFactory := newQueryService
	newQueryService = func(ctx context.Context) (core.QueryService, error) { return spy, nil }
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

	if spy.calls != 1 {
		t.Fatalf("expected ExecuteQuery to be called once, got %d", spy.calls)
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

func TestRootCommandFileArgument(t *testing.T) {
	type ctxKey struct{}

	spy := &spyQueryService{}
	origFactory := newQueryService
	newQueryService = func(ctx context.Context) (core.QueryService, error) { return spy, nil }
	t.Cleanup(func() {
		newQueryService = origFactory
		rootCmd.SetArgs(nil)
		rootCmd.SetContext(context.Background())
	})

	tmpFile, err := os.CreateTemp(t.TempDir(), "query-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	if _, err := tmpFile.WriteString("g.V().count()"); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	rootCmd.SetContext(context.WithValue(context.Background(), ctxKey{}, "query-file"))
	rootCmd.SetArgs([]string{"--type", "gremlin", tmpFile.Name()})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() returned error: %v", err)
	}

	if spy.calls != 1 {
		t.Fatalf("expected ExecuteQuery to be called once, got %d", spy.calls)
	}
	if spy.lastQuery != "g.V().count()" {
		t.Fatalf("expected query content 'g.V().count()', got %q", spy.lastQuery)
	}
	if spy.lastQueryType != "gremlin" {
		t.Fatalf("expected query type 'gremlin', got %q", spy.lastQueryType)
	}
	if got := spy.lastCtx.Value(ctxKey{}); got != "query-file" {
		t.Fatalf("expected command context to propagate, got %v", got)
	}
}
