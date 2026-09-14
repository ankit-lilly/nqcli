package schema

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

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
	case strings.HasPrefix(query, "g.V().hasLabel("):
		content = `[["id","name"]]`
	case strings.HasPrefix(query, "g.E().hasLabel(") && strings.Contains(query, ".properties()"):
		content = `[["createdAt"]]`
	case strings.HasPrefix(query, "g.E().hasLabel("):
		content = `[[{"sourceVertexType":"Study","targetVertexType":"StudyVersion"}]]`
	default:
		return core.QueryResult{}, fmt.Errorf("unexpected query %q", query)
	}
	return core.QueryResult{Content: content}, nil
}

type countingQueryService struct {
	calls atomic.Int64
}

func (s *countingQueryService) ExecuteQuery(_ context.Context, query, _ string, _ core.QueryOpts) (core.QueryResult, error) {
	s.calls.Add(1)
	if strings.Contains(query, "groupCount") {
		counts := make(map[string]int, 13)
		prefix := "Vertex"
		if strings.HasPrefix(query, "g.E()") {
			prefix = "Edge"
		}
		for index := range 13 {
			counts[fmt.Sprintf("%s%d", prefix, index)] = 1
		}
		payload, _ := json.Marshal([]any{counts})
		return core.QueryResult{Content: string(payload)}, nil
	}
	return core.QueryResult{Content: `[[]]`}, nil
}

type blockingQueryService struct {
	started chan struct{}
}

func (s blockingQueryService) ExecuteQuery(ctx context.Context, _ string, _ string, _ core.QueryOpts) (core.QueryResult, error) {
	select {
	case s.started <- struct{}{}:
	default:
	}
	<-ctx.Done()
	return core.QueryResult{}, ctx.Err()
}

type memoryStore struct {
	snapshots map[string]Snapshot
	saves     int
}

func (s *memoryStore) Load(key string) (Snapshot, error) {
	snapshot, ok := s.snapshots[key]
	if !ok {
		return Snapshot{}, ErrCacheMiss
	}
	return snapshot, nil
}

func (s *memoryStore) Save(key string, snapshot Snapshot) error {
	s.snapshots[key] = snapshot
	s.saves++
	return nil
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

func TestDiscoveryBoundsConcurrentPerLabelQueries(t *testing.T) {
	t.Parallel()
	query := &countingQueryService{}
	snapshot, err := (discovery{query: query}).discover(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Vertices) != 13 || len(snapshot.Edges) != 13 {
		t.Fatalf("unexpected schema sizes: %d vertices, %d edges", len(snapshot.Vertices), len(snapshot.Edges))
	}
	// Two count queries, 13 vertex-property queries, and 13 each for edge properties and connections.
	if got := query.calls.Load(); got != 41 {
		t.Fatalf("query count = %d, want 41", got)
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

func TestServiceHydratesFreshPersistentSnapshot(t *testing.T) {
	t.Parallel()
	now := time.Now()
	store := &memoryStore{snapshots: map[string]Snapshot{
		"dev": {Status: StatusReady, LastUpdate: &now, Vertices: []Vertex{{Type: "Study"}}},
	}}
	service := NewService(nil, WithStore(store))

	got := service.Revalidate(context.Background(), "dev", testQueryService{}, false)
	if got.Status != StatusReady || len(got.Vertices) != 1 || got.Vertices[0].Type != "Study" {
		t.Fatalf("unexpected hydrated snapshot: %#v", got)
	}
}

func TestServicePersistsCompletedRefresh(t *testing.T) {
	t.Parallel()
	store := &memoryStore{snapshots: make(map[string]Snapshot)}
	service := NewService(nil, WithStore(store))
	service.Revalidate(context.Background(), "dev", testQueryService{}, true)
	if !service.Wait(context.Background(), "dev") {
		t.Fatal("refresh did not complete")
	}
	if store.saves != 1 || store.snapshots["dev"].Status != StatusReady {
		t.Fatalf("completed snapshot was not persisted: %#v", store)
	}
}

func TestServiceCloseCancelsAndWaitsForRefresh(t *testing.T) {
	t.Parallel()
	service := NewService(nil)
	started := make(chan struct{}, 1)
	service.Revalidate(context.Background(), "dev", blockingQueryService{started: started}, true)

	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("schema refresh did not start")
	}

	closed := make(chan struct{})
	go func() {
		service.Close()
		close(closed)
	}()
	select {
	case <-closed:
	case <-time.After(time.Second):
		t.Fatal("Close did not wait for the canceled refresh")
	}

	service.Close()
	if got := service.Revalidate(context.Background(), "dev", testQueryService{}, true); got.Status == StatusRunning {
		t.Fatal("closed service started another refresh")
	}
}
