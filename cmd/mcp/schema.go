package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

const staticSchemaJSON = `{
  "schema_version": "static",
  "root_label": "Study",
  "notes": [
    "Treat this schema as authoritative for vertex labels and edge labels.",
    "Do not invent property keys beyond those listed in 'properties'.",
    "Traversal source is g."
  ],
  "properties": {
    "all_vertices": [
      "id",
      "instanceType",
      "name",
      "label",
      "description",
      "createdAt",
      "updatedAt",
      "extensionAttributes"
    ],
    "StudyVersion": ["versionIdentifier", "rationale"],
    "StudyIdentifier": ["text", "scopeId"],
    "USDMSource": [
      "createdBy",
      "authorId",
      "openLabel",
      "sentBy",
      "sentAt",
      "sourceName",
      "updatedBy"
    ],
    "Collaborator": ["email"],
    "TherapeuticArea": ["code", "decode", "codeSystem", "codeSystemVersion"],
    "Indication": ["isRareDisease"],
    "StudyDesign": ["rationale", "compound"],
    "StudyTitle": ["text"],
    "Organization": ["identifierScheme", "identifier"],
    "Address": ["text", "lines", "city", "district", "state", "postalCode"],
    "Code": ["code", "codeSystem", "codeSystemVersion", "decode"],
    "AliasCode": ["standardCodeAliases"],
    "Encounter": ["previousId", "nextId", "scheduledAtId"],
    "Activity": ["previousId", "nextId", "timelineId", "childIds", "biomedicalConceptIds", "bcCategoryIds"],
    "StudyEpoch": ["previousId", "nextId"],
    "ScheduleTimeline": ["mainTimeline", "entryCondition", "entryId", "plannedDuration"],
    "Timing": [
      "value",
      "valueLabel",
      "relativeFromScheduledInstanceId",
      "relativeToScheduledInstanceId",
      "windowLower",
      "windowUpper",
      "windowLabel"
    ],
    "ScheduledActivityInstance": [
      "defaultConditionId",
      "defaultCondition",
      "epochId",
      "timelineId",
      "timelineExitId",
      "encounterId",
      "activityIds"
    ],
    "Condition": ["text", "dictionaryId", "contextIds", "appliesToIds"],
    "StudyRoleRelationship": [
      "biomedicalConceptIds",
      "bcCategoryIds",
      "scheduledActivityInstanceIds"
    ],
    "BiomedicalConcept": ["reference", "synonyms"],
    "BiomedicalConceptCategory": ["members"],
    "BiomedicalConceptProperty": ["isRequired", "isEnabled", "datatype"],
    "ResponseCode": ["isEnabled"],
    "Note": ["text"]
  },
  "known_instance_types": {
    "StudyDesign": ["InterventionalStudyDesign", "ObservationalStudyDesign"]
  },
  "schema": {
    "Study": {
      "versions": { "edgeLabel": "has_version", "childLabel": "StudyVersion" }
    },
    "StudyVersion": {
      "studyIdentifiers": { "edgeLabel": "has_identifier", "childLabel": "StudyIdentifier" },
      "studyDesigns": { "edgeLabel": "has_design", "childLabel": "StudyDesign" },
      "titles": { "edgeLabel": "has_title", "childLabel": "StudyTitle" },
      "organizations": { "edgeLabel": "has_organization", "childLabel": "Organization" },
      "conditions": { "edgeLabel": "has_condition", "childLabel": "Condition" },
      "sourceVersion": { "edgeLabel": "has_source_version", "childLabel": "USDMSource" },
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" },
      "bcCategories": {
        "edgeLabel": "has_biomedical_category",
        "childLabel": "BiomedicalConceptCategory"
      },
      "biomedicalConcepts": { "edgeLabel": "has_biomedical_concept", "childLabel": "BiomedicalConcept" },
      "roles": { "edgeLabel": "has_role", "childLabel": "StudyRole" }
    },
    "USDMSource": {
      "collaborators": { "edgeLabel": "has_collaborator", "childLabel": "Collaborator" }
    },
    "Collaborator": {},
    "StudyIdentifier": {},
    "StudyDesign": {
      "studyType": { "edgeLabel": "has_study_type", "childLabel": "Code" },
      "studyPhase": { "edgeLabel": "has_phase", "childLabel": "AliasCode" },
      "encounters": { "edgeLabel": "has_encounter", "childLabel": "Encounter" },
      "activities": { "edgeLabel": "has_activity", "childLabel": "Activity" },
      "epochs": { "edgeLabel": "has_epoch", "childLabel": "StudyEpoch" },
      "scheduleTimelines": { "edgeLabel": "has_timeline", "childLabel": "ScheduleTimeline" },
      "intentTypes": { "edgeLabel": "has_intent_type", "childLabel": "Code" },
      "subTypes": { "edgeLabel": "has_sub_type", "childLabel": "Code" },
      "model": { "edgeLabel": "has_model", "childLabel": "Code" },
      "therapeuticAreas": { "edgeLabel": "has_therapeutic_area", "childLabel": "TherapeuticArea" },
      "indications": { "edgeLabel": "has_indication", "childLabel": "Indication" },
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" }
    },
    "StudyTitle": {
      "type": { "edgeLabel": "has_type", "childLabel": "Code" }
    },
    "Organization": {
      "type": { "edgeLabel": "has_type", "childLabel": "Code" },
      "legalAddress": { "edgeLabel": "has_address", "childLabel": "Address" }
    },
    "StudyRole": {
      "code": { "edgeLabel": "has_code", "childLabel": "Code" },
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" }
    },
    "Address": {
      "country": { "edgeLabel": "located_in", "childLabel": "Code" }
    },
    "Encounter": {
      "type": { "edgeLabel": "has_type", "childLabel": "Code" },
      "environmentalSettings": { "edgeLabel": "has_setting", "childLabel": "Code" },
      "contactModes": { "edgeLabel": "has_contact_mode", "childLabel": "Code" },
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" }
    },
    "Activity": {
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" }
    },
    "StudyEpoch": {
      "type": { "edgeLabel": "has_type", "childLabel": "Code" },
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" }
    },
    "ScheduleTimeline": {
      "exits": { "edgeLabel": "has_exit", "childLabel": "ScheduleTimelineExit" },
      "timings": { "edgeLabel": "has_timing", "childLabel": "Timing" },
      "instances": { "edgeLabel": "has_instance", "childLabel": "ScheduledActivityInstance" },
      "studyRoleRelationships": {
        "edgeLabel": "has_role_relationship",
        "childLabel": "StudyRoleRelationship"
      }
    },
    "Timing": {
      "type": { "edgeLabel": "has_type", "childLabel": "Code" },
      "relativeToFrom": { "edgeLabel": "has_relative_type", "childLabel": "Code" }
    },
    "ScheduledActivityInstance": {},
    "AliasCode": {
      "standardCode": { "edgeLabel": "has_standard_code", "childLabel": "Code" }
    },
    "Condition": {
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" }
    },
    "Code": {
      "extensionAttributes": { "edgeLabel": "has_extension_attribute", "childLabel": "ExtensionAttribute" }
    },
    "Indication": {
      "codes": { "edgeLabel": "has_code", "childLabel": "Code" },
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" }
    },
    "ScheduleTimelineExit": {},
    "BiomedicalConceptCategory": {
      "code": { "edgeLabel": "has_code", "childLabel": "AliasCode" },
      "members": { "edgeLabel": "has_member", "childLabel": "BiomedicalConcept" },
      "children": { "edgeLabel": "has_child_category", "childLabel": "BiomedicalConceptCategory" },
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" }
    },
    "BiomedicalConcept": {
      "code": { "edgeLabel": "has_code", "childLabel": "AliasCode" },
      "properties": { "edgeLabel": "has_property", "childLabel": "BiomedicalConceptProperty" },
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" }
    },
    "BiomedicalConceptProperty": {
      "code": { "edgeLabel": "has_code", "childLabel": "AliasCode" },
      "responseCodes": { "edgeLabel": "has_response_code", "childLabel": "ResponseCode" },
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" }
    },
    "ResponseCode": {
      "code": { "edgeLabel": "has_code", "childLabel": "Code" }
    },
    "Note": {
      "codes": { "edgeLabel": "has_code", "childLabel": "Code" }
    },
    "StudyRoleRelationship": {
      "code": { "edgeLabel": "has_code", "childLabel": "Code" },
      "biomedicalConceptIds": {
        "edgeLabel": "has_biomedical_concept",
        "childLabel": "BiomedicalConcept"
      },
      "bcCategoryIds": {
        "edgeLabel": "has_biomedical_category",
        "childLabel": "BiomedicalConceptCategory"
      },
      "scheduledActivityInstanceIds": {
        "edgeLabel": "targets_instance",
        "childLabel": "ScheduledActivityInstance"
      },
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" }
    }
  }
}`

const (
	schemaSourceEnvVar   = "NQ_MCP_SCHEMA_SOURCE"
	schemaSourceDynamic  = "dynamic"
	enumSampleLimit      = 10
	enumReturnLimit      = 10
	enumSampleValueLimit = 5
)

type graphSchema struct {
	SchemaVersion string                 `json:"schema_version"`
	GeneratedAt   string                 `json:"generated_at"`
	VertexLabels  []string               `json:"vertex_labels"`
	EdgeLabels    []string               `json:"edge_labels"`
	EdgePatterns  []map[string]string    `json:"edge_patterns"`
	Vertices      map[string]labelSchema `json:"vertices"`
	Edges         map[string]labelSchema `json:"edges"`
}

type labelSchema struct {
	Count      int64          `json:"count,omitempty"`
	Properties []propertyInfo `json:"properties,omitempty"`
}

type propertyInfo struct {
	Name         string `json:"name"`
	Enum         []any  `json:"enum,omitempty"`
	SampleValues []any  `json:"sample_values,omitempty"`
}

func buildGraphSchema(ctx context.Context, svc QueryService) (string, error) {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv(schemaSourceEnvVar)))
	if mode != schemaSourceDynamic {
		return staticSchemaJSON, nil
	}

	dynamicSchema, err := discoverGraphSchema(ctx, svc)
	if err != nil {
		return staticSchemaJSON, nil
	}

	payload, err := json.MarshalIndent(dynamicSchema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func discoverGraphSchema(ctx context.Context, svc QueryService) (*graphSchema, error) {
	vertexLabels, err := queryStringList(ctx, svc, "g.V().label().dedup()")
	if err != nil {
		return nil, fmt.Errorf("discover vertex labels: %w", err)
	}
	edgeLabels, err := queryStringList(ctx, svc, "g.E().label().dedup()")
	if err != nil {
		return nil, fmt.Errorf("discover edge labels: %w", err)
	}
	edgePatterns, err := queryEdgePatterns(ctx, svc)
	if err != nil {
		return nil, fmt.Errorf("discover edge patterns: %w", err)
	}

	slices.Sort(vertexLabels)
	slices.Sort(edgeLabels)

	vertices := make(map[string]labelSchema, len(vertexLabels))
	for _, label := range vertexLabels {
		ls, err := discoverLabelSchema(ctx, svc, true, label)
		if err != nil {
			return nil, err
		}
		vertices[label] = ls
	}

	edges := make(map[string]labelSchema, len(edgeLabels))
	for _, label := range edgeLabels {
		ls, err := discoverLabelSchema(ctx, svc, false, label)
		if err != nil {
			return nil, err
		}
		edges[label] = ls
	}

	return &graphSchema{
		SchemaVersion: "dynamic",
		GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		VertexLabels:  vertexLabels,
		EdgeLabels:    edgeLabels,
		EdgePatterns:  edgePatterns,
		Vertices:      vertices,
		Edges:         edges,
	}, nil
}

func discoverLabelSchema(ctx context.Context, svc QueryService, isVertex bool, label string) (labelSchema, error) {
	prefix := "g.E()"
	if isVertex {
		prefix = "g.V()"
	}
	escaped := escapeGremlinString(label)

	props, err := queryStringList(ctx, svc,
		fmt.Sprintf("%s.hasLabel('%s').properties().key().dedup()", prefix, escaped))
	if err != nil {
		return labelSchema{}, fmt.Errorf("discover properties for %s: %w", label, err)
	}

	slices.Sort(props)
	infos := make([]propertyInfo, 0, len(props))
	for _, prop := range props {
		values, err := queryEnumCandidates(ctx, svc, prefix, escaped, prop)
		if err != nil {
			return labelSchema{}, fmt.Errorf("analyze properties for %s: %w", label, err)
		}
		info := propertyInfo{Name: prop}
		if len(values) > 0 {
			sampleLimit := min(enumSampleValueLimit, len(values))
			info.SampleValues = values[:sampleLimit]
			if len(values) <= enumReturnLimit {
				info.Enum = values
			}
		}
		infos = append(infos, info)
	}

	count, err := queryCount(ctx, svc,
		fmt.Sprintf("%s.hasLabel('%s').count()", prefix, escaped))
	if err != nil {
		return labelSchema{}, fmt.Errorf("count for %s: %w", label, err)
	}

	return labelSchema{Count: count, Properties: infos}, nil
}

func queryEnumCandidates(ctx context.Context, svc QueryService, prefix, escapedLabel, prop string) ([]any, error) {
	query := fmt.Sprintf(
		"%s.hasLabel('%s').values('%s').dedup().limit(%d)",
		prefix, escapedLabel, escapeGremlinString(prop), enumSampleLimit+1,
	)
	values, err := queryAnyList(ctx, svc, query)
	if err != nil {
		return nil, err
	}
	if len(values) > enumSampleLimit {
		return nil, nil
	}
	return values, nil
}

func queryStringList(ctx context.Context, svc QueryService, query string) ([]string, error) {
	raw, err := executeGremlin(ctx, svc, query)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}
	switch v := raw.(type) {
	case []string:
		return v, nil
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			s, ok := item.(string)
			if !ok {
				return nil, fmt.Errorf("expected string item, got %T", item)
			}
			out = append(out, s)
		}
		return out, nil
	default:
		return nil, fmt.Errorf("expected list, got %T", raw)
	}
}

func queryAnyList(ctx context.Context, svc QueryService, query string) ([]any, error) {
	raw, err := executeGremlin(ctx, svc, query)
	if err != nil {
		return nil, err
	}
	if raw == nil {
		return nil, nil
	}
	switch v := raw.(type) {
	case []any:
		return v, nil
	case []string:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = item
		}
		return out, nil
	default:
		return nil, fmt.Errorf("expected list, got %T", raw)
	}
}

func queryEdgePatterns(ctx context.Context, svc QueryService) ([]map[string]string, error) {
	raw, err := executeGremlin(ctx, svc, "g.E().project('out','label','in').by(outV().label()).by(label()).by(inV().label()).dedup()")
	if err != nil {
		return nil, err
	}
	anyList, ok := raw.([]any)
	if !ok {
		return nil, fmt.Errorf("expected list, got %T", raw)
	}
	patterns := make([]map[string]string, 0, len(anyList))
	for _, item := range anyList {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		outVal, _ := m["out"].(string)
		labelVal, _ := m["label"].(string)
		inVal, _ := m["in"].(string)
		if outVal != "" && labelVal != "" && inVal != "" {
			patterns = append(patterns, map[string]string{
				"out":   outVal,
				"label": labelVal,
				"in":    inVal,
			})
		}
	}
	return patterns, nil
}

func queryCount(ctx context.Context, svc QueryService, query string) (int64, error) {
	raw, err := executeGremlin(ctx, svc, query)
	if err != nil {
		return 0, err
	}
	switch v := raw.(type) {
	case float64:
		return int64(v), nil
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case json.Number:
		return v.Int64()
	case string:
		parsed, parseErr := strconv.ParseInt(v, 10, 64)
		if parseErr == nil {
			return parsed, nil
		}
	}
	return 0, fmt.Errorf("unexpected count type %T", raw)
}

func executeGremlin(ctx context.Context, svc QueryService, query string) (any, error) {
	prettyJSON, _, err := svc.ExecuteQueryCtx(ctx, query, "gremlin")
	if err != nil {
		return nil, err
	}
	var payload any
	if err := json.Unmarshal([]byte(prettyJSON), &payload); err != nil {
		return nil, fmt.Errorf("parse gremlin response: %w", err)
	}
	return payload, nil
}

// escapeGremlinString prevents injection in single-quoted Gremlin string literals.
func escapeGremlinString(value string) string {
	if value == "" {
		return value
	}
	var b strings.Builder
	b.Grow(len(value))
	for _, r := range value {
		switch r {
		case '\\':
			b.WriteString(`\\`)
		case '\'':
			b.WriteString(`\'`)
		case '"':
			b.WriteString(`\"`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	return b.String()
}
