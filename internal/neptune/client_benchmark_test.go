package neptune

import (
	"strings"
	"testing"

	"github.com/ankit-lilly/nqcli/internal/core"
)

func BenchmarkNormalizeDirectGremlinPayload(b *testing.B) {
	const vertex = `{"@type":"g:Vertex","@value":{"id":"v1","label":"Study"}}`
	var data strings.Builder
	data.Grow(1 << 20)
	data.WriteString(`{"requestId":"1","result":{"data":{"@type":"g:List","@value":[`)
	first := true
	for data.Len()+len(vertex)+4 < 1<<20 {
		if !first {
			data.WriteByte(',')
		}
		data.WriteString(vertex)
		first = false
	}
	data.WriteString(`]}}}`)
	payload := []byte(data.String())

	b.ReportAllocs()
	b.SetBytes(int64(len(payload)))
	for b.Loop() {
		result := normalizeDirectGremlinPayload(payload, core.QueryOpts{})
		if len(result) == 0 {
			b.Fatal("empty normalized payload")
		}
	}
}

func BenchmarkUnwrapSingleDataEnvelope(b *testing.B) {
	payload := []byte(`{"data":[` + strings.Repeat(`{"value":123},`, 32768) + `null]}`)
	b.ReportAllocs()
	b.SetBytes(int64(len(payload)))
	for b.Loop() {
		if _, ok := unwrapSingleDataEnvelope(payload); !ok {
			b.Fatal("envelope was not unwrapped")
		}
	}
}
