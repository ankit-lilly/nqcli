package neptune

import "testing"

func TestNeptuneHTTPSURL(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		endpoint string
		port     int32
		want     string
	}{
		{
			name:     "host only",
			endpoint: "example.cluster.us-east-1.neptune.amazonaws.com",
			port:     8182,
			want:     "https://example.cluster.us-east-1.neptune.amazonaws.com:8182",
		},
		{
			name:     "keeps existing scheme",
			endpoint: "https://example.cluster.us-east-1.neptune.amazonaws.com",
			port:     8182,
			want:     "https://example.cluster.us-east-1.neptune.amazonaws.com:8182",
		},
		{
			name:     "keeps existing port",
			endpoint: "https://example.cluster.us-east-1.neptune.amazonaws.com:9999",
			port:     8182,
			want:     "https://example.cluster.us-east-1.neptune.amazonaws.com:9999",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if got := neptuneHTTPSURL(tc.endpoint, tc.port); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
