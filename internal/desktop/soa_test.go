package desktop

import (
	"context"
	"os"
	"strings"
	"testing"

	"github.com/ankit-lilly/nqcli/internal/core"
)

func TestDecodeSoAGraphSON(t *testing.T) {
	raw := `{"@type":"g:List","@value":[{"@type":"g:Map","@value":["id","instance-1","activities",{"@type":"g:List","@value":["a","b"]}]}]}`
	got, err := decodeData(raw)
	if err != nil {
		t.Fatal(err)
	}
	entries := got.([]any)
	item := entries[0].(map[string]any)
	if item["id"] != "instance-1" || len(item["activities"].([]any)) != 2 {
		t.Fatalf("lost fields: %#v", got)
	}
}
func TestGremlinLiteral(t *testing.T) {
	if got := gremlinString("a'b\\c\n"); got != "'a\\'b\\\\c\\n'" {
		t.Fatalf("unsafe literal: %q", got)
	}
}
func TestSoAQueryScopesToVersionAndDesign(t *testing.T) {
	spy := &spyService{result: core.QueryResult{Content: "[]"}}
	service := newTestService(spy)
	response := service.GetSoA(SoASelection{VersionID: "v'1", DesignID: "d1"})
	if response.Error != "" {
		t.Fatal(response.Error)
	}
	if !strings.HasPrefix(spy.lastQuery, "g.V('v\\'1').hasLabel('StudyVersion').as('v').out('has_design').hasId('d1')") {
		t.Fatal(spy.lastQuery)
	}
	if !spy.lastOpts.SkipFormatting || spy.lastOpts.MaxResponseBytes == 0 {
		t.Fatal("SoA must use bounded compact results")
	}
}
func TestLifecycleAndCancelQueries(t *testing.T) {
	service := NewDesktopService(&spyService{}, "dev", nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	service.Startup(ctx)
	old := service.ctx
	service.CancelQueries()
	if old.Err() == nil {
		t.Fatal("active request was not cancelled")
	}
	if service.ctx.Err() != nil {
		t.Fatal("subsequent requests cannot run")
	}
	cancel()
	if service.ctx.Err() == nil {
		t.Fatal("application shutdown must cancel subsequent requests")
	}
}
func TestBoundGraphKeepsConnectedEdges(t *testing.T) {
	result, warning := boundGraph([]GraphElement{{Group: "nodes", Data: map[string]any{"id": "a"}}, {Group: "edges", Data: map[string]any{"id": "edge", "source": "a", "target": "missing"}}})
	if len(result) != 1 || warning == "" {
		t.Fatal("dangling edge was not removed")
	}
}

// Optional read-only query fixture used for manual MCP validation; no network in tests.
func TestWriteSoAQuery(t *testing.T) {
	path := os.Getenv("NQ_SOA_QUERY_OUTPUT")
	if path == "" {
		t.Skip("optional query export")
	}
	if err := os.WriteFile(path, []byte(soaQuery(SoASelection{VersionID: os.Getenv("NQ_VERSION_ID"), DesignID: os.Getenv("NQ_DESIGN_ID")})), 0600); err != nil {
		t.Fatal(err)
	}
}
func TestPlainEnvelopeAndTypedVertexProperties(t *testing.T) {
	props, err := parseValueMap(`{"data":{"@type":"g:List","@value":[{"@type":"g:Map","@value":["name",{"@type":"g:List","@value":["Trial"]}]}]}}`)
	if err != nil || props["name"] != "Trial" {
		t.Fatalf("%#v: %v", props, err)
	}
}

type waitingService struct{ started chan struct{} }

func (s *waitingService) ExecuteQuery(ctx context.Context, _ string, _ string, _ core.QueryOpts) (core.QueryResult, error) {
	close(s.started)
	<-ctx.Done()
	return core.QueryResult{}, ctx.Err()
}
func TestCancelReachesBackendRequest(t *testing.T) {
	backend := &waitingService{started: make(chan struct{})}
	service := NewDesktopService(backend, "dev", nil)
	done := make(chan QueryResponse, 1)
	go func() { done <- service.ExecuteQuery(QueryRequest{Query: "g.V()"}) }()
	<-backend.started
	service.CancelQueries()
	if response := <-done; response.Error == "" {
		t.Fatal("cancelled request returned success")
	}
}
