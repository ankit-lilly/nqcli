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

	"github.com/ankit-lilly/nqcli/internal/core"
	"golang.org/x/sync/errgroup"
)

const staticSchemaJSON = `{
  "schema_version": "static",
  "root_label": "Study",
  "notes": [
    "Treat this schema as authoritative for vertex labels and edge labels.",
    "Do not invent property keys beyond those listed in 'properties'.",
    "Traversal source is g.",
    "Cross-reference edges (next_encounter, in_epoch, etc.) link siblings; they are NOT parent-child.",
    "Entries in 'id_references' describe semantic ID links carried on properties. They are not materialized graph edges unless the same relationship also appears in 'cross_references'."
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
    "Study": ["studyId", "systemName", "systemVersion", "usdmVersion", "is-dtf-trial", "is-pediatric-study"],
    "StudyVersion": ["versionIdentifier", "rationale", "protocolStatus"],
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
    "Note": ["text"],
    "StudyArm": ["dataOriginDescription", "populationIds"],
    "StudyCell": ["armId", "epochId", "elementIds"],
    "StudyDesignPopulation": ["includesHealthySubjects", "plannedSex", "criterionIds"],
    "EligibilityCriterion": ["identifier", "criterionItemId"],
    "SubjectEnrollment": ["forGeographicScope", "forStudyCohortId", "forStudySiteId"],
    "Quantity": ["value"],
    "Amendment": ["number", "summary", "previousId"],
    "StudySite": [],
    "StudyAmendmentReason": [],
    "ScheduleTimelineExit": [],
    "ExtensionAttribute": []
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
      "roles": { "edgeLabel": "has_role", "childLabel": "StudyRole" },
      "amendments": { "edgeLabel": "has_amendment", "childLabel": "Amendment" },
      "enrollments": { "edgeLabel": "has_enrollment", "childLabel": "SubjectEnrollment" }
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
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" },
      "arms": { "edgeLabel": "has_arm", "childLabel": "StudyArm" },
      "studyCells": { "edgeLabel": "has_study_cell", "childLabel": "StudyCell" },
      "populations": { "edgeLabel": "has_population", "childLabel": "StudyDesignPopulation" },
      "eligibilityCriteria": { "edgeLabel": "has_eligibility_criterion", "childLabel": "EligibilityCriterion" },
      "timePerspective": { "edgeLabel": "has_time_perspective", "childLabel": "Code" }
    },
    "StudyTitle": {
      "type": { "edgeLabel": "has_type", "childLabel": "Code" }
    },
    "Organization": {
      "type": { "edgeLabel": "has_type", "childLabel": "Code" },
      "legalAddress": { "edgeLabel": "has_address", "childLabel": "Address" },
      "managedSites": { "edgeLabel": "has_managed_site", "childLabel": "StudySite" }
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
    },
    "StudyArm": {
      "type": { "edgeLabel": "has_type", "childLabel": "Code" },
      "dataOriginType": { "edgeLabel": "has_data_origin_type", "childLabel": "Code" },
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" }
    },
    "StudyCell": {},
    "StudyDesignPopulation": {
      "plannedSex": { "edgeLabel": "has_planned_sex", "childLabel": "Code" },
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" }
    },
    "EligibilityCriterion": {
      "category": { "edgeLabel": "has_category", "childLabel": "Code" },
      "notes": { "edgeLabel": "has_note", "childLabel": "Note" }
    },
    "StudySite": {
      "country": { "edgeLabel": "has_country", "childLabel": "Code" }
    },
    "SubjectEnrollment": {
      "quantity": { "edgeLabel": "has_quantity", "childLabel": "Quantity" }
    },
    "Quantity": {
      "unit": { "edgeLabel": "has_unit", "childLabel": "AliasCode" }
    },
    "StudyAmendmentReason": {
      "code": { "edgeLabel": "has_code", "childLabel": "Code" }
    },
    "Amendment": {
      "primaryReason": { "edgeLabel": "has_primary_reason", "childLabel": "StudyAmendmentReason" },
      "enrollments": { "edgeLabel": "has_enrollment", "childLabel": "SubjectEnrollment" }
    },
    "TherapeuticArea": {},
    "ExtensionAttribute": {}
  },
  "id_references": {
    "StudyArm.populationIds": { "target": "StudyDesignPopulation", "multi": true },
    "StudyCell.armId": { "target": "StudyArm" },
    "StudyCell.epochId": { "target": "StudyEpoch" },
    "StudyCell.elementIds": { "target": "StudyElement", "multi": true },
    "StudyDesignPopulation.criterionIds": { "target": "EligibilityCriterion", "multi": true },
    "EligibilityCriterion.criterionItemId": { "target": "EligibilityCriterionItem" },
    "SubjectEnrollment.forStudyCohortId": { "target": "StudyCohort" },
    "Amendment.previousId": { "target": "Amendment" }
  },
  "cross_references": {
    "Study->has_latest_version->StudyVersion": "mutable pointer to the most recent version",
    "StudyIdentifier->scoped_by->Organization": "scopeId",
    "Encounter->next_encounter->Encounter": "linked list via previousId",
    "Encounter->scheduled_at->ScheduledActivityInstance": "scheduledAtId",
    "Activity->next_activity->Activity": "linked list via previousId",
    "Activity->parent_of->Activity": "childIds",
    "StudyEpoch->next_epoch->StudyEpoch": "linked list via previousId",
    "ScheduledActivityInstance->in_epoch->StudyEpoch": "epochId",
    "ScheduledActivityInstance->occurs_in->Encounter": "encounterId",
    "ScheduledActivityInstance->includes_activity->Activity": "activityIds",
    "ScheduledActivityInstance->exits_at->ScheduleTimelineExit": "timelineExitId",
    "Timing->relative_from->ScheduledActivityInstance": "relativeFromScheduledInstanceId",
    "Timing->relative_to->ScheduledActivityInstance": "relativeToScheduledInstanceId",
    "Condition->applies_to_context->(any vertex)": "contextIds",
    "Condition->applies_to->(any vertex)": "appliesToIds",
    "SubjectEnrollment->for_study_site->StudySite": "forStudySiteId"
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

func buildGraphSchema(ctx context.Context, svc core.QueryService) (string, error) {
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

func discoverGraphSchema(ctx context.Context, svc core.QueryService) (*graphSchema, error) {
	var vertexLabels, edgeLabels []string
	var edgePatterns []map[string]string
	labels, labelsCtx := errgroup.WithContext(ctx)
	labels.Go(func() error {
		var err error
		vertexLabels, err = queryStringList(labelsCtx, svc, "g.V().label().dedup()")
		if err != nil {
			return fmt.Errorf("discover vertex labels: %w", err)
		}
		return nil
	})
	labels.Go(func() error {
		var err error
		edgeLabels, err = queryStringList(labelsCtx, svc, "g.E().label().dedup()")
		if err != nil {
			return fmt.Errorf("discover edge labels: %w", err)
		}
		return nil
	})
	labels.Go(func() error {
		var err error
		edgePatterns, err = queryEdgePatterns(labelsCtx, svc)
		if err != nil {
			return fmt.Errorf("discover edge patterns: %w", err)
		}
		return nil
	})
	if err := labels.Wait(); err != nil {
		return nil, err
	}

	slices.Sort(vertexLabels)
	slices.Sort(edgeLabels)

	vertexSchemas := make([]labelSchema, len(vertexLabels))
	edgeSchemas := make([]labelSchema, len(edgeLabels))
	details, detailsCtx := errgroup.WithContext(ctx)
	details.SetLimit(4)
	for index := range max(len(vertexLabels), len(edgeLabels)) {
		if index < len(vertexLabels) {
			label := vertexLabels[index]
			details.Go(func() error {
				discovered, err := discoverLabelSchema(detailsCtx, svc, true, label)
				if err == nil {
					vertexSchemas[index] = discovered
				}
				return err
			})
		}
		if index < len(edgeLabels) {
			label := edgeLabels[index]
			details.Go(func() error {
				discovered, err := discoverLabelSchema(detailsCtx, svc, false, label)
				if err == nil {
					edgeSchemas[index] = discovered
				}
				return err
			})
		}
	}
	if err := details.Wait(); err != nil {
		return nil, err
	}

	vertices := make(map[string]labelSchema, len(vertexLabels))
	for index, label := range vertexLabels {
		vertices[label] = vertexSchemas[index]
	}
	edges := make(map[string]labelSchema, len(edgeLabels))
	for index, label := range edgeLabels {
		edges[label] = edgeSchemas[index]
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

func discoverLabelSchema(ctx context.Context, svc core.QueryService, isVertex bool, label string) (labelSchema, error) {
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

func queryEnumCandidates(ctx context.Context, svc core.QueryService, prefix, escapedLabel, prop string) ([]any, error) {
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

func queryStringList(ctx context.Context, svc core.QueryService, query string) ([]string, error) {
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

func queryAnyList(ctx context.Context, svc core.QueryService, query string) ([]any, error) {
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

func queryEdgePatterns(ctx context.Context, svc core.QueryService) ([]map[string]string, error) {
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

func queryCount(ctx context.Context, svc core.QueryService, query string) (int64, error) {
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

func executeGremlin(ctx context.Context, svc core.QueryService, query string) (any, error) {
	result, err := svc.ExecuteQuery(ctx, query, "gremlin", core.QueryOpts{SkipFormatting: true, MaxResponseBytes: 8 << 20})
	if err != nil {
		return nil, err
	}
	var payload any
	if err := json.Unmarshal([]byte(result.Content), &payload); err != nil {
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
