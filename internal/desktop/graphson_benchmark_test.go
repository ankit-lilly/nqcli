package desktop

import (
	"strconv"
	"strings"
	"testing"
)

func BenchmarkParseGraphSON(b *testing.B) {
	payload := benchmarkGraphSON(500)
	b.ReportAllocs()
	b.SetBytes(int64(len(payload)))
	for b.Loop() {
		elements, err := ParseGraphSON(payload)
		if err != nil {
			b.Fatal(err)
		}
		if len(elements) != 999 {
			b.Fatalf("got %d elements", len(elements))
		}
	}
}

func benchmarkGraphSON(nodes int) string {
	var builder strings.Builder
	builder.Grow(nodes * 240)
	builder.WriteByte('[')
	for index := range nodes {
		if index > 0 {
			builder.WriteByte(',')
		}
		id := strconv.Itoa(index)
		builder.WriteString(`{"@type":"g:Vertex","@value":{"id":"v`)
		builder.WriteString(id)
		builder.WriteString(`","label":"StudyVersion","properties":{"name":[{"@type":"g:VertexProperty","@value":{"value":"Version `)
		builder.WriteString(id)
		builder.WriteString(`"}}]}}}`)
		if index > 0 {
			builder.WriteString(`,{"@type":"g:Edge","@value":{"id":"e`)
			builder.WriteString(id)
			builder.WriteString(`","label":"has_version","outV":"v0","inV":"v`)
			builder.WriteString(id)
			builder.WriteString(`"}}`)
		}
	}
	builder.WriteByte(']')
	return builder.String()
}
