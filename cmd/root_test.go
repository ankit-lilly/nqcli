package cmd

import (
	"context"
	"os"
	"testing"
)

type spyQueryService struct {
	executeCalls      int
	executeQueryCalls int
	lastQuery         string
	lastQueryType     string
}

func (s *spyQueryService) Execute(path, queryType string) (string, string, error) {
	s.executeCalls++
	s.lastQuery = path
	s.lastQueryType = queryType
	return "{}", "", nil
}

func (s *spyQueryService) ExecuteQuery(query, queryType string) (string, string, error) {
	s.executeQueryCalls++
	s.lastQuery = query
	s.lastQueryType = queryType
	return "{}", "", nil
}

func TestRootCommandInlineQueryCallsExecuteQuery(t *testing.T) {
	spy := &spyQueryService{}
	origFactory := newQueryService
	newQueryService = func(ctx context.Context) (queryService, error) { return spy, nil }
	t.Cleanup(func() {
		newQueryService = origFactory
		rootCmd.SetArgs(nil)
	})

	rootCmd.SetArgs([]string{"--type", "cypher", "g.V()"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() returned error: %v", err)
	}

	if spy.executeQueryCalls != 1 {
		t.Fatalf("expected ExecuteQuery to be called once, got %d", spy.executeQueryCalls)
	}
	if spy.executeCalls != 0 {
		t.Fatalf("expected Execute not to be called, got %d", spy.executeCalls)
	}
	if spy.lastQuery != "g.V()" {
		t.Fatalf("expected query 'g.V()', got %q", spy.lastQuery)
	}
	if spy.lastQueryType != "cypher" {
		t.Fatalf("expected query type 'cypher', got %q", spy.lastQueryType)
	}
}

func TestRootCommandFileArgumentCallsExecute(t *testing.T) {
	spy := &spyQueryService{}
	origFactory := newQueryService
	newQueryService = func(ctx context.Context) (queryService, error) { return spy, nil }
	t.Cleanup(func() {
		newQueryService = origFactory
		rootCmd.SetArgs(nil)
	})

	tmpFile, err := os.CreateTemp(t.TempDir(), "query-*.txt")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	tmpFile.Close()

	rootCmd.SetArgs([]string{"--type", "gremlin", tmpFile.Name()})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("rootCmd.Execute() returned error: %v", err)
	}

	if spy.executeCalls != 1 {
		t.Fatalf("expected Execute to be called once, got %d", spy.executeCalls)
	}
	if spy.executeQueryCalls != 0 {
		t.Fatalf("expected ExecuteQuery not to be called, got %d", spy.executeQueryCalls)
	}
	if spy.lastQuery != tmpFile.Name() {
		t.Fatalf("expected query path %q, got %q", tmpFile.Name(), spy.lastQuery)
	}
	if spy.lastQueryType != "gremlin" {
		t.Fatalf("expected query type 'gremlin', got %q", spy.lastQueryType)
	}
}
