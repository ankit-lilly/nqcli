package mcp

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/ankit-lilly/nqcli/internal/core"
)

type stubQueryService struct {
	execCalls int
	execErr   error
	result    core.QueryResult
}

func (s *stubQueryService) ExecuteQuery(_ context.Context, _ string, _ string, _ core.QueryOpts) (core.QueryResult, error) {
	s.execCalls++
	if s.execErr != nil {
		return core.QueryResult{}, s.execErr
	}
	if s.result == (core.QueryResult{}) {
		return core.QueryResult{Content: "[]", Processed: "[]", Raw: "[]"}, nil
	}
	return s.result, nil
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

func TestStaticSchemaDocumentsSemanticIDReferences(t *testing.T) {
	var payload struct {
		Notes        []string                  `json:"notes"`
		IDReferences map[string]map[string]any `json:"id_references"`
	}

	if err := json.Unmarshal([]byte(staticSchemaJSON), &payload); err != nil {
		t.Fatalf("static schema must stay valid JSON: %v", err)
	}

	if len(payload.Notes) == 0 || !strings.Contains(payload.Notes[len(payload.Notes)-1], "id_references") {
		t.Fatalf("expected notes to explain id_references")
	}

	tests := map[string]string{
		"StudyArm.populationIds":               "StudyDesignPopulation",
		"StudyCell.armId":                      "StudyArm",
		"StudyCell.epochId":                    "StudyEpoch",
		"StudyCell.elementIds":                 "StudyElement",
		"StudyDesignPopulation.criterionIds":   "EligibilityCriterion",
		"EligibilityCriterion.criterionItemId": "EligibilityCriterionItem",
		"SubjectEnrollment.forStudyCohortId":   "StudyCohort",
		"Amendment.previousId":                 "Amendment",
	}

	for key, wantTarget := range tests {
		ref, ok := payload.IDReferences[key]
		if !ok {
			t.Fatalf("expected %s to be documented in id_references", key)
		}
		gotTarget, _ := ref["target"].(string)
		if gotTarget != wantTarget {
			t.Fatalf("expected %s target %q, got %q", key, wantTarget, gotTarget)
		}
	}
}

func TestExecuteGremlinUsesCanonicalContent(t *testing.T) {
	svc := &stubQueryService{
		result: core.QueryResult{
			Content:   `[{"id":"1"}]`,
			Processed: `not json`,
			Raw:       `{"data":[{"id":"1"}]}`,
		},
	}

	payload, err := executeGremlin(context.Background(), svc, "g.V()")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	items, ok := payload.([]any)
	if !ok || len(items) != 1 {
		t.Fatalf("expected one-item payload, got %#v", payload)
	}
}
