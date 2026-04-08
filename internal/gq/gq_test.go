package gq

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ankit-lilly/nqcli/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

func TestExecuteGraphQLPostsSignedPayload(t *testing.T) {
	t.Parallel()

	var received GraphQLPayload
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if auth := r.Header.Get("Authorization"); !strings.Contains(auth, "AWS4-HMAC-SHA256") {
			t.Fatalf("expected signed request, got Authorization=%q", auth)
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if err := json.Unmarshal(body, &received); err != nil {
			t.Fatalf("unmarshal request body: %v", err)
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"trials":[]}}`))
	}))
	defer server.Close()

	client, err := NewClient(
		&config.Config{URL: server.URL},
		aws.Config{
			Region:      "us-east-1",
			Credentials: credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
		},
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	raw, err := client.ExecuteGraphQL("query { trials { trialAlias } }", map[string]any{"limit": 5})
	if err != nil {
		t.Fatalf("ExecuteGraphQL: %v", err)
	}
	if raw != `{"data":{"trials":[]}}` {
		t.Fatalf("unexpected response: %s", raw)
	}
	if received.Query != "query { trials { trialAlias } }" {
		t.Fatalf("unexpected query: %q", received.Query)
	}

	var variables map[string]any
	encoded, err := json.Marshal(received.Variables)
	if err != nil {
		t.Fatalf("marshal variables: %v", err)
	}
	if err := json.Unmarshal(encoded, &variables); err != nil {
		t.Fatalf("unmarshal variables: %v", err)
	}
	if variables["limit"] != float64(5) {
		t.Fatalf("unexpected variables: %#v", variables)
	}
}
