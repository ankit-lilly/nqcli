package core

import (
	"strings"
	"testing"
)

func BenchmarkFormatOutput(b *testing.B) {
	payload := benchmarkJSONArray(1 << 20)
	b.ReportAllocs()
	b.SetBytes(int64(len(payload)))
	for b.Loop() {
		result := formatOutput(payload)
		if len(result) == 0 {
			b.Fatal("empty formatted output")
		}
	}
}

func benchmarkJSONArray(targetBytes int) string {
	const row = `{"id":"4b99f24c-8de2-45de-a3d2-d72a64ca88bf","label":"StudyVersion","count":123456789,"active":true}`
	var builder strings.Builder
	builder.Grow(targetBytes + len(row))
	builder.WriteByte('[')
	for builder.Len()+len(row)+2 < targetBytes {
		if builder.Len() > 1 {
			builder.WriteByte(',')
		}
		builder.WriteString(row)
	}
	builder.WriteByte(']')
	return builder.String()
}
