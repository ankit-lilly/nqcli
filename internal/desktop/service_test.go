package desktop

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/ankit-lilly/nqcli/internal/core"
)

type spyService struct {
	lastQuery string
	lastType  string
	lastOpts  core.QueryOpts
	result    core.QueryResult
	err       error
}

func (s *spyService) ExecuteQuery(_ context.Context, query, queryType string, opts core.QueryOpts) (core.QueryResult, error) {
	s.lastQuery = query
	s.lastType = queryType
	s.lastOpts = opts
	return s.result, s.err
}

func newTestService(spy *spyService) *DesktopService {
	svc := NewDesktopService(spy, "test-profile", nil)
	svc.ctx = context.Background()
	return svc
}

func TestExecuteQuery_PassesTypeAndQuery(t *testing.T) {
	t.Parallel()
	spy := &spyService{result: core.QueryResult{Processed: "{}", Raw: "{}"}}
	svc := newTestService(spy)

	resp := svc.ExecuteQuery(QueryRequest{Query: "g.V()", Type: "cypher"})

	if spy.lastQuery != "g.V()" {
		t.Fatalf("expected query 'g.V()', got %q", spy.lastQuery)
	}
	if spy.lastType != "cypher" {
		t.Fatalf("expected type 'cypher', got %q", spy.lastType)
	}
	if resp.Error != "" {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
}

func TestExecuteQuery_DefaultsToGremlin(t *testing.T) {
	t.Parallel()
	spy := &spyService{result: core.QueryResult{Processed: "{}", Raw: "{}"}}
	svc := newTestService(spy)

	svc.ExecuteQuery(QueryRequest{Query: "g.V()"})

	if spy.lastType != "gremlin" {
		t.Fatalf("expected default type 'gremlin', got %q", spy.lastType)
	}
}

func TestExecuteQuery_ReturnsError(t *testing.T) {
	t.Parallel()
	spy := &spyService{err: fmt.Errorf("connection failed")}
	svc := newTestService(spy)

	resp := svc.ExecuteQuery(QueryRequest{Query: "g.V()", Type: "gremlin"})

	if resp.Error == "" {
		t.Fatal("expected error in response")
	}
}

func TestExecuteQuery_PassesSerializer(t *testing.T) {
	t.Parallel()
	spy := &spyService{result: core.QueryResult{Processed: "{}", Raw: "{}"}}
	svc := newTestService(spy)

	svc.ExecuteQuery(QueryRequest{Query: "g.V()", Type: "gremlin", Serializer: "application/json"})

	if spy.lastOpts.Serializer != "application/json" {
		t.Fatalf("expected serializer 'application/json', got %q", spy.lastOpts.Serializer)
	}
}

func TestExecuteGraphQuery_UsesDefaultSerializer(t *testing.T) {
	t.Parallel()
	spy := &spyService{result: core.QueryResult{Content: "[]", Raw: "[]"}}
	svc := newTestService(spy)

	svc.ExecuteGraphQuery(QueryRequest{Query: "g.V()"})

	if spy.lastOpts.Serializer != "" {
		t.Fatalf("expected empty serializer, got %q", spy.lastOpts.Serializer)
	}
}

func TestExecuteGraphQuery_ParsesElements(t *testing.T) {
	t.Parallel()
	graphson := `[{"@type":"g:Vertex","@value":{"id":"v1","label":"Study"}}]`
	spy := &spyService{result: core.QueryResult{Content: graphson, Raw: graphson}}
	svc := newTestService(spy)

	resp := svc.ExecuteGraphQuery(QueryRequest{Query: "g.V()"})

	if resp.Error != "" {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
	if len(resp.Elements) != 1 {
		t.Fatalf("expected 1 element, got %d", len(resp.Elements))
	}
	if resp.Elements[0].Data["label"] != "Study" {
		t.Fatalf("expected label 'Study', got %v", resp.Elements[0].Data["label"])
	}
}

func TestExecuteGraphQuery_ReturnsErrorOnFailure(t *testing.T) {
	t.Parallel()
	spy := &spyService{err: fmt.Errorf("timeout")}
	svc := newTestService(spy)

	resp := svc.ExecuteGraphQuery(QueryRequest{Query: "g.V()"})

	if resp.Error == "" {
		t.Fatal("expected error in response")
	}
}

func TestExpandVertex_ExecutesGremlinAndParsesGraph(t *testing.T) {
	t.Parallel()
	raw := `[
		{"id":"v2","label":"StudyVersion"},
		{"id":"e1","label":"has_version","outV":"v1","inV":"v2"}
	]`
	spy := &spyService{result: core.QueryResult{Content: raw}}
	svc := newTestService(spy)

	resp := svc.ExpandVertex(ExpandVertexRequest{
		ID:                "v1",
		Type:              "gremlin",
		Direction:         "out",
		Relationship:      "has_version",
		NeighborLabel:     "StudyVersion",
		Limit:             12,
		ExcludedVertexIDs: []string{"v1", "already-visible"},
	})

	if resp.Error != "" {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
	if len(resp.Elements) != 2 {
		t.Fatalf("expected two graph elements, got %d", len(resp.Elements))
	}
	if spy.lastType != "gremlin" {
		t.Fatalf("expected gremlin query, got %q", spy.lastType)
	}
	for _, want := range []string{
		"g.V('v1').as('start').out('has_version')",
		".hasLabel('StudyVersion')",
		".hasId('already-visible')",
		".range(0,12)",
		".outE('has_version')",
	} {
		if !strings.Contains(spy.lastQuery, want) {
			t.Errorf("query %q does not contain %q", spy.lastQuery, want)
		}
	}
}

func TestExpandVertex_ExecutesOpenCypher(t *testing.T) {
	t.Parallel()
	spy := &spyService{result: core.QueryResult{Content: `{"results":[]}`}}
	svc := newTestService(spy)

	resp := svc.ExpandVertex(ExpandVertexRequest{
		ID:            `vertex"1`,
		Type:          "cypher",
		Direction:     "in",
		Relationship:  "has`version",
		NeighborLabel: "StudyVersion",
		Limit:         500,
	})

	if resp.Error != "" {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
	if spy.lastType != "cypher" {
		t.Fatalf("expected cypher query, got %q", spy.lastType)
	}
	for _, want := range []string{
		"MATCH (source)<-[edge:`has``version`]-(neighbor:`StudyVersion`)",
		`ID(source) = "vertex\"1"`,
		"LIMIT 100",
	} {
		if !strings.Contains(spy.lastQuery, want) {
			t.Errorf("query %q does not contain %q", spy.lastQuery, want)
		}
	}
}

func TestExpandVertex_RejectsInvalidRequest(t *testing.T) {
	t.Parallel()
	svc := newTestService(&spyService{})

	for _, req := range []ExpandVertexRequest{
		{},
		{ID: "v1", Type: "sparql"},
		{ID: "v1", Direction: "sideways"},
	} {
		if resp := svc.ExpandVertex(req); resp.Error == "" {
			t.Fatalf("expected validation error for %#v", req)
		}
	}
}

func TestGetProfile_ReturnsCurrentProfile(t *testing.T) {
	t.Parallel()
	spy := &spyService{}
	svc := NewDesktopService(spy, "dsoadev", nil)
	svc.ctx = context.Background()

	info := svc.GetProfile()

	if info.Profile != "dsoadev" {
		t.Fatalf("expected profile 'dsoadev', got %q", info.Profile)
	}
	if info.Env != "dev" {
		t.Fatalf("expected env 'dev', got %q", info.Env)
	}
}

func TestSwitchProfile_ChangesService(t *testing.T) {
	t.Parallel()
	spy1 := &spyService{result: core.QueryResult{Processed: "first"}}
	spy2 := &spyService{result: core.QueryResult{Processed: "second"}}
	factory := func(_ context.Context, profile string) (core.QueryService, error) {
		if profile == "dsoaqa" {
			return spy2, nil
		}
		return spy1, nil
	}
	svc := NewDesktopService(spy1, "dsoadev", factory)
	svc.ctx = context.Background()

	resp := svc.SwitchProfile(ProfileRequest{Profile: "dsoaqa"})

	if resp.Error != "" {
		t.Fatalf("unexpected error: %s", resp.Error)
	}
	if resp.Profile != "dsoaqa" {
		t.Fatalf("expected profile 'dsoaqa', got %q", resp.Profile)
	}
	if resp.Env != "qa" {
		t.Fatalf("expected env 'qa', got %q", resp.Env)
	}
}
