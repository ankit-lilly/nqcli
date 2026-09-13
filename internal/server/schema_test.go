package server

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ankit-lilly/nqcli/internal/core"
)

type schemaTestService struct{}

func (schemaTestService) ExecuteQuery(_ context.Context, query, _ string, _ core.QueryOpts) (core.QueryResult, error) {
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
		{
			name:    "wrapped group count map",
			content: `[{"Study":1113,"StudyVersion":12445}]`,
			target:  func() any { return &map[string]any{} },
			want:    map[string]any{"Study": float64(1113), "StudyVersion": float64(12445)},
		},
		{
			name:    "unwrapped group count map",
			content: `{"Study":1113}`,
			target:  func() any { return &map[string]any{} },
			want:    map[string]any{"Study": float64(1113)},
		},
		{
			name:    "wrapped folded properties",
			content: `[["id","name"]]`,
			target:  func() any { return &[]string{} },
			want:    []string{"id", "name"},
		},
		{
			name:    "wrapped folded edge connections",
			content: `[[{"sourceVertexType":"Study","targetVertexType":"StudyVersion"}]]`,
			target:  func() any { return &[]map[string]any{} },
			want: []map[string]any{{
				"sourceVertexType": "Study",
				"targetVertexType": "StudyVersion",
			}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			target := tt.target()
			if err := decodeQueryJSON(tt.content, target); err != nil {
				t.Fatalf("decodeQueryJSON() error = %v", err)
			}
			if got := reflect.ValueOf(target).Elem().Interface(); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("decodeQueryJSON() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestDecodeQueryJSONRejectsMultipleTraversersForSingletonTarget(t *testing.T) {
	t.Parallel()

	var target map[string]any
	if err := decodeQueryJSON(`[{"Study":1},{"Study":2}]`, &target); err == nil {
		t.Fatal("decodeQueryJSON() expected an error")
	}
}

func TestSchemaDiscoveryAcceptsNeptuneTraverserLists(t *testing.T) {
	t.Parallel()

	snapshot, err := (schemaDiscovery{service: schemaTestService{}}).discover(context.Background())
	if err != nil {
		t.Fatalf("discover() error = %v", err)
	}
	if len(snapshot.Vertices) != 2 || len(snapshot.Edges) != 1 {
		t.Fatalf("unexpected schema sizes: %d vertices, %d edges", len(snapshot.Vertices), len(snapshot.Edges))
	}
	if len(snapshot.EdgeConnections) != 1 {
		t.Fatalf("expected one edge connection, got %d", len(snapshot.EdgeConnections))
	}
	connection := snapshot.EdgeConnections[0]
	if connection.SourceVertexType != "Study" || connection.TargetVertexType != "StudyVersion" {
		t.Fatalf("unexpected edge connection: %#v", connection)
	}
}
