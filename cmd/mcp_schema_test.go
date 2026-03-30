package cmd

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type stubQueryService struct {
	execCalls int
	execErr   error
}

func (s *stubQueryService) Execute(_ string, _ string) (string, string, error) {
	return "", "", errors.New("not implemented")
}

func (s *stubQueryService) ExecuteQuery(_ string, _ string) (string, string, error) {
	s.execCalls++
	if s.execErr != nil {
		return "", "", s.execErr
	}
	return "[]", "", nil
}

func TestBuildGraphSchemaStaticDefault(t *testing.T) {
	t.Setenv(schemaSourceEnvVar, "")

	service := &stubQueryService{}
	got, err := buildGraphSchema(context.Background(), service)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := strings.TrimSpace(staticSchemaJSON)
	if strings.TrimSpace(got) != want {
		t.Fatalf("expected static schema to be returned")
	}
	if service.execCalls != 0 {
		t.Fatalf("expected no discovery queries, got %d", service.execCalls)
	}
}

func TestBuildGraphSchemaDynamicFallbackOnError(t *testing.T) {
	t.Setenv(schemaSourceEnvVar, schemaSourceDynamic)

	service := &stubQueryService{execErr: errors.New("boom")}
	got, err := buildGraphSchema(context.Background(), service)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := strings.TrimSpace(staticSchemaJSON)
	if strings.TrimSpace(got) != want {
		t.Fatalf("expected static schema fallback to be returned")
	}
	if service.execCalls == 0 {
		t.Fatalf("expected discovery queries to be attempted")
	}
}
