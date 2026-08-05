package mcp

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

func (s *stubQueryService) ExecuteQueryCtx(_ context.Context, _ string, _ string) (string, string, error) {
	s.execCalls++
	if s.execErr != nil {
		return "", "", s.execErr
	}
	return "[]", "", nil
}

func TestBuildGraphSchemaStaticDefault(t *testing.T) {
	t.Setenv(schemaSourceEnvVar, "")

	svc := &stubQueryService{}
	got, err := buildGraphSchema(context.Background(), svc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(got) != strings.TrimSpace(staticSchemaJSON) {
		t.Fatalf("expected static schema to be returned")
	}
	if svc.execCalls != 0 {
		t.Fatalf("expected no discovery queries, got %d", svc.execCalls)
	}
}

func TestBuildGraphSchemaDynamicFallbackOnError(t *testing.T) {
	t.Setenv(schemaSourceEnvVar, schemaSourceDynamic)

	svc := &stubQueryService{execErr: errors.New("boom")}
	got, err := buildGraphSchema(context.Background(), svc)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if strings.TrimSpace(got) != strings.TrimSpace(staticSchemaJSON) {
		t.Fatalf("expected static schema fallback to be returned")
	}
	if svc.execCalls == 0 {
		t.Fatalf("expected discovery queries to be attempted")
	}
}
