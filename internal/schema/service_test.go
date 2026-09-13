package schema

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ankit-lilly/nqcli/internal/core"
)

type testQueryService struct{}

func (testQueryService) ExecuteQuery(_ context.Context, query, _ string, _ core.QueryOpts) (core.QueryResult, error) {
	var content string
	switch {
	case query == "g.V().label().groupCount()":
		content = `[{"Study":1,"StudyVersion":2}]`
	case query == "g.E().label().groupCount()":
		content = `[{"has_version":1}]`
	case strings.HasPrefix(query, "g.V()"):
		content = `[["id","name"]]`
	case strings.HasPrefix(query, "g.E().hasLabel") && strings.Contains(query, ".properties()"):
		content = `[["createdAt"]]`
	case strings.HasPrefix(query, "g.E().hasLabel"):
		content = `[[{"sourceVertexType":"Study","targetVertexType":"StudyVersion"}]]`
	default:
		return core.QueryResult{}, fmt.Errorf("unexpected query %q", query)
	}
	return core.QueryResult{Content: content}, nil
}

func TestDecodeQueryJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		content string
		target  func() any
		want    any
	}{
		{"wrapped group count", `[{"Study":1113}]`, func() any { return &map[string]any{} }, map[string]any{"Study": float64(1113)}},
		{"unwrapped group count", `{"Study":1113}`, func() any { return &map[string]any{} }, map[string]any{"Study": float64(1113)}},
		{"wrapped properties", `[["id","name"]]`, func() any { return &[]string{} }, []string{"id", "name"}},
		{"wrapped connections", `[[{"sourceVertexType":"Study","targetVertexType":"StudyVersion"}]]`, func() any { return &[]map[string]any{} }, []map[string]any{{"sourceVertexType": "Study", "targetVertexType": "StudyVersion"}}},
		{"API envelope", `{"data":[["id","name"]],"meta":{}}`, func() any { return &[]string{} }, []string{"id", "name"}},
		{"empty API envelope", `{"data":[[]],"meta":{}}`, func() any { return &[]string{} }, []string{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			target := test.target()
			if err := decodeQueryJSON(test.content, target); err != nil {
				t.Fatal(err)
			}
			if got := reflect.ValueOf(target).Elem().Interface(); !reflect.DeepEqual(got, test.want) {
				t.Fatalf("got %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestDecodeQueryJSONRejectsMultipleTraversers(t *testing.T) {
	t.Parallel()
	var target map[string]any
	if decodeQueryJSON(`[{"Study":1},{"Study":2}]`, &target) == nil {
		t.Fatal("expected an error")
	}
}

func TestDiscoveryAcceptsNeptuneTraverserLists(t *testing.T) {
	t.Parallel()
	snapshot, err := (discovery{query: testQueryService{}}).discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Vertices) != 2 || len(snapshot.Edges) != 1 || len(snapshot.EdgeConnections) != 1 {
		t.Fatalf("unexpected schema: %#v", snapshot)
	}
	connection := snapshot.EdgeConnections[0]
	if connection.SourceVertexType != "Study" || connection.TargetVertexType != "StudyVersion" {
		t.Fatalf("unexpected connection: %#v", connection)
	}
}

func TestServiceRevalidatesAndCachesByConnection(t *testing.T) {
	t.Parallel()
	service := NewService(nil)
	started := service.Revalidate(context.Background(), "dev", testQueryService{}, false)
	if started.Status != StatusRunning {
		t.Fatalf("start status = %q, want %q", started.Status, StatusRunning)
	}
	if !service.Wait(context.Background(), "dev") {
		t.Fatal("refresh did not complete")
	}
	ready := service.Get("dev")
	if ready.Status != StatusReady || len(ready.Vertices) != 2 {
		t.Fatalf("unexpected cached snapshot: %#v", ready)
	}
	if got := service.Revalidate(context.Background(), "dev", testQueryService{}, false); got.Status != StatusReady {
		t.Fatalf("fresh cache status = %q, want %q", got.Status, StatusReady)
	}
}
