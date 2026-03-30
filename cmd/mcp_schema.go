package cmd

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
	schemaSourceStatic   = "static"
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

func buildGraphSchema(_ context.Context, appService queryService) (string, error) {
	staticSchema := strings.TrimSpace(staticSchemaJSON)
	mode := strings.ToLower(strings.TrimSpace(os.Getenv(schemaSourceEnvVar)))
	if mode == "" {
		mode = schemaSourceStatic
	}

	if mode != schemaSourceDynamic && staticSchema != "" {
		return staticSchema, nil
	}

	dynamicSchema, err := discoverGraphSchema(appService)
	if err == nil {
		payload, marshalErr := json.MarshalIndent(dynamicSchema, "", "  ")
		if marshalErr != nil {
			return "", marshalErr
		}
		return string(payload), nil
	}

	if staticSchema != "" {
		return staticSchema, nil
	}

	return "", err
}

func discoverGraphSchema(appService queryService) (*graphSchema, error) {
	vertexLabels, err := queryStringList(appService, "g.V().label().dedup()")
	if err != nil {
		return nil, fmt.Errorf("discover vertex labels: %w", err)
	}
	edgeLabels, err := queryStringList(appService, "g.E().label().dedup()")
	if err != nil {
		return nil, fmt.Errorf("discover edge labels: %w", err)
	}
	edgePatterns, err := queryEdgePatterns(appService)
	if err != nil {
		return nil, fmt.Errorf("discover edge patterns: %w", err)
	}

	slices.Sort(vertexLabels)
	slices.Sort(edgeLabels)

	vertices := make(map[string]labelSchema, len(vertexLabels))
	for _, label := range vertexLabels {
		props, propErr := queryStringList(
			appService,
			fmt.Sprintf("g.V().hasLabel('%s').properties().key().dedup()", escapeGremlinString(label)),
		)
		if propErr != nil {
			return nil, fmt.Errorf("discover vertex properties for %s: %w", label, propErr)
		}
		propInfos, enumErr := buildPropertyInfos(appService, true, label, props)
		if enumErr != nil {
			return nil, fmt.Errorf("analyze vertex properties for %s: %w", label, enumErr)
		}
		count, countErr := queryCount(
			appService,
			fmt.Sprintf("g.V().hasLabel('%s').count()", escapeGremlinString(label)),
		)
		if countErr != nil {
			return nil, fmt.Errorf("count vertices for %s: %w", label, countErr)
		}
		vertices[label] = labelSchema{
			Count:      count,
			Properties: propInfos,
		}
	}

	edges := make(map[string]labelSchema, len(edgeLabels))
	for _, label := range edgeLabels {
		props, propErr := queryStringList(
			appService,
			fmt.Sprintf("g.E().hasLabel('%s').properties().key().dedup()", escapeGremlinString(label)),
		)
		if propErr != nil {
			return nil, fmt.Errorf("discover edge properties for %s: %w", label, propErr)
		}
		propInfos, enumErr := buildPropertyInfos(appService, false, label, props)
		if enumErr != nil {
			return nil, fmt.Errorf("analyze edge properties for %s: %w", label, enumErr)
		}
		count, countErr := queryCount(
			appService,
			fmt.Sprintf("g.E().hasLabel('%s').count()", escapeGremlinString(label)),
		)
		if countErr != nil {
			return nil, fmt.Errorf("count edges for %s: %w", label, countErr)
		}
		edges[label] = labelSchema{
			Count:      count,
			Properties: propInfos,
		}
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

func buildPropertyInfos(appService queryService, isVertex bool, label string, props []string) ([]propertyInfo, error) {
	slices.Sort(props)
	infos := make([]propertyInfo, 0, len(props))
	for _, prop := range props {
		values, err := queryEnumCandidates(appService, isVertex, label, prop)
		if err != nil {
			return nil, err
		}

		info := propertyInfo{Name: prop}
		if len(values) > 0 {
			sampleLimit := enumSampleValueLimit
			if len(values) < sampleLimit {
				sampleLimit = len(values)
			}
			info.SampleValues = values[:sampleLimit]
		}
		if len(values) > 0 && len(values) <= enumReturnLimit {
			info.Enum = values
		}
		infos = append(infos, info)
	}
	return infos, nil
}

func queryEnumCandidates(appService queryService, isVertex bool, label, prop string) ([]any, error) {
	prefix := "g.V()"
	if !isVertex {
		prefix = "g.E()"
	}
	query := fmt.Sprintf(
		"%s.hasLabel('%s').values('%s').dedup().limit(%d)",
		prefix,
		escapeGremlinString(label),
		escapeGremlinString(prop),
		enumSampleLimit+1,
	)
	values, err := queryAnyList(appService, query)
	if err != nil {
		return nil, err
	}
	if len(values) > enumSampleLimit {
		return nil, nil
	}
	return values, nil
}

func queryStringList(appService queryService, query string) ([]string, error) {
	raw, err := executeGremlin(appService, query)
	if err != nil {
		return nil, err
	}
	return asStringSlice(raw)
}

func queryAnyList(appService queryService, query string) ([]any, error) {
	raw, err := executeGremlin(appService, query)
	if err != nil {
		return nil, err
	}
	return asAnySlice(raw)
}

func queryEdgePatterns(appService queryService) ([]map[string]string, error) {
	raw, err := executeGremlin(appService, "g.E().project('out','label','in').by(outV().label()).by(label()).by(inV().label()).dedup()")
	if err != nil {
		return nil, err
	}
	anyList, err := asAnySlice(raw)
	if err != nil {
		return nil, err
	}
	patterns := make([]map[string]string, 0, len(anyList))
	for _, item := range anyList {
		m, ok := item.(map[string]any)
		if !ok {
			continue
		}
		outVal, outOk := m["out"].(string)
		labelVal, labelOk := m["label"].(string)
		inVal, inOk := m["in"].(string)
		if outOk && labelOk && inOk {
			patterns = append(patterns, map[string]string{
				"out":   outVal,
				"label": labelVal,
				"in":    inVal,
			})
		}
	}
	return patterns, nil
}

func queryCount(appService queryService, query string) (int64, error) {
	raw, err := executeGremlin(appService, query)
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

func executeGremlin(appService queryService, query string) (any, error) {
	prettyJSON, _, err := appService.ExecuteQuery(query, "gremlin")
	if err != nil {
		return nil, err
	}
	var payload any
	if err := json.Unmarshal([]byte(prettyJSON), &payload); err != nil {
		return nil, fmt.Errorf("parse gremlin response: %w", err)
	}
	return payload, nil
}

func asAnySlice(value any) ([]any, error) {
	if value == nil {
		return nil, nil
	}
	switch v := value.(type) {
	case []any:
		return v, nil
	case []string:
		out := make([]any, len(v))
		for i, item := range v {
			out[i] = item
		}
		return out, nil
	default:
		return nil, fmt.Errorf("expected list, got %T", value)
	}
}

func asStringSlice(value any) ([]string, error) {
	list, err := asAnySlice(value)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(list))
	for _, item := range list {
		str, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("expected string item, got %T", item)
		}
		out = append(out, str)
	}
	return out, nil
}

func escapeGremlinString(value string) string {
	if value == "" {
		return value
	}
	return strings.ReplaceAll(value, "'", "\\'")
}
