package filecache

import (
	"errors"
	"reflect"
	"testing"
	"time"

	graphschema "github.com/ankit-lilly/nqcli/internal/schema"
)

func TestStoreRoundTrip(t *testing.T) {
	t.Parallel()
	store := New(t.TempDir())
	now := time.Now().UTC().Truncate(time.Nanosecond)
	want := graphschema.Snapshot{
		Status:     graphschema.StatusReady,
		LastUpdate: &now,
		Vertices: []graphschema.Vertex{{
			Type:       "Study",
			Attributes: []graphschema.Attribute{{Name: "name", DataType: "String"}},
		}},
		Edges:           []graphschema.Edge{},
		EdgeConnections: []graphschema.EdgeConnection{},
	}
	if err := store.Save("dsoadev", want); err != nil {
		t.Fatal(err)
	}
	got, err := store.Load("dsoadev")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("loaded snapshot = %#v, want %#v", got, want)
	}
}

func TestStoreReportsCacheMiss(t *testing.T) {
	t.Parallel()
	store := New(t.TempDir())
	if _, err := store.Load("missing"); !errors.Is(err, graphschema.ErrCacheMiss) {
		t.Fatalf("Load error = %v, want cache miss", err)
	}
}
