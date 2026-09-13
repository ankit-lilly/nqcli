package desktop

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ankit-lilly/nqcli/internal/core"
)

// Values are always quoted as literals, never interpolated as Gremlin source.
func gremlinString(s string) string {
	return "'" + strings.NewReplacer("\\", "\\\\", "'", "\\'", "\n", "\\n", "\r", "\\r", "\t", "\\t").Replace(s) + "'"
}

type StudySearch struct {
	Search string `json:"search"`
}
type StudySelection struct {
	StudyID string `json:"studyId"`
}
type SoASelection struct {
	VersionID string `json:"versionId"`
	DesignID  string `json:"designId"`
}
type DataResponse struct {
	Data    any    `json:"data"`
	Error   string `json:"error,omitempty"`
	Warning string `json:"warning,omitempty"`
}

const latestVersion = "coalesce(out('has_latest_version'),out('has_version').order().by('createdAt',desc).limit(1))"

func (d *DesktopService) ListStudies(req StudySearch) DataResponse {
	q := "g.V().hasLabel('Study')"
	if strings.TrimSpace(req.Search) != "" {
		q += ".has('name',containing(" + gremlinString(strings.TrimSpace(req.Search)) + "))"
	}
	// No database-wide sort on every keystroke. Search explicitly; narrow to find more.
	q += ".limit(51).project('id','name').by(id).by(coalesce(values('name'),constant('Unnamed study')))"
	return d.dataQuery(q, 50)
}

func (d *DesktopService) ListVersions(req StudySelection) DataResponse {
	if req.StudyID == "" {
		return DataResponse{Error: "Select a study"}
	}
	q := "g.V(" + gremlinString(req.StudyID) + ").hasLabel('Study').project('latest','versions').by(" + latestVersion + ".id().fold()).by(out('has_version').order().by('createdAt',desc).limit(201).project('id','name','status','amendment','designs').by(id).by(coalesce(values('versionIdentifier'),constant('Unknown version'))).by(coalesce(values('protocolStatus'),constant('Unknown'))).by(out('has_amendment').values('name').fold()).by(out('has_design').limit(21).project('id','name').by(id).by(coalesce(values('name'),values('instanceType'),constant('Study design'))).fold()).fold())"
	return d.dataQuery(q, 0)
}

// Each collection has a sentinel element so the UI can explicitly report incomplete data.
// Project only the fields needed for a matrix; no repeated full graph paths or pretty JSON.
func soaQuery(req SoASelection) string {
	return "g.V(" + gremlinString(req.VersionID) + ").hasLabel('StudyVersion').as('v').out('has_design').hasId(" + gremlinString(req.DesignID) + ").project('visits','activities','labs','instances','timings','conditions')" +
		".by(out('has_encounter').limit(1001).project('id','sourceId','name','previousId','nextId','modality').by(id).by(coalesce(values('id'),id())).by(coalesce(values('label'),values('name'),constant('Visit'))).by(coalesce(in('next_encounter').id(),values('previousId'),constant(''))).by(coalesce(out('next_encounter').id(),values('nextId'),constant(''))).by(out('has_contact_mode').values('decode').fold()).fold())" +
		".by(out('has_activity').limit(1001).project('id','name').by(id).by(coalesce(values('label'),values('name'),constant('Activity'))).fold())" +
		".by(select('v').out('has_biomedical_concept').limit(1001).project('id','name').by(id).by(coalesce(values('label'),values('name'),constant('Lab concept'))).fold())" +
		".by(out('has_timeline').out('has_instance').limit(2001).project('id','visits','activities','epochs').by(id).by(out('occurs_in').id().fold()).by(out('includes_activity').id().fold()).by(out('in_epoch').values('name').fold()).fold())" +
		".by(out('has_timeline').out('has_timing').limit(2001).project('id','value','windowLower','windowUpper','windowLabel','from','to','type','relativeType').by(id).by(coalesce(values('valueLabel'),values('value'),constant(''))).by(coalesce(values('windowLower'),constant(''))).by(coalesce(values('windowUpper'),constant(''))).by(coalesce(values('windowLabel'),constant(''))).by(out('relative_from').id().fold()).by(out('relative_to').id().fold()).by(out('has_type').values('decode').fold()).by(out('has_relative_type').values('decode').fold()).fold())" +
		".by(select('v').out('has_condition').limit(501).project('id','text','activities','instances').by(id).by(coalesce(values('text'),constant(''))).by(out('applies_to').id().fold()).by(out('applies_to_context').id().fold()).fold())"
}

func (d *DesktopService) GetSoA(req SoASelection) DataResponse {
	if req.VersionID == "" || req.DesignID == "" {
		return DataResponse{Error: "Select a version and study design"}
	}
	return d.dataQuery(soaQuery(req), 0)
}

func (d *DesktopService) dataQuery(query string, limit int) DataResponse {
	d.mu.RLock()
	svc, ctx := d.svc, d.ctx
	d.mu.RUnlock()
	result, err := svc.ExecuteQuery(ctx, query, "gremlin", core.QueryOpts{SkipFormatting: true, MaxResponseBytes: 8 << 20})
	if err != nil {
		return DataResponse{Error: err.Error()}
	}
	content := result.Content
	if content == "" {
		content = result.Raw
	}
	value, err := decodeData(content)
	if err != nil {
		return DataResponse{Error: err.Error()}
	}
	warning := ""
	if list, ok := value.([]any); ok && limit > 0 && len(list) > limit {
		value = list[:limit]
		warning = "Showing the first 50 matches. Narrow your search to find more studies."
	}
	return DataResponse{Data: value, Warning: warning}
}

// Accept typed GraphSON as well as the REST endpoint's plain JSON data envelope.
func decodeData(raw string) (any, error) {
	var value any
	decoder := json.NewDecoder(strings.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		return nil, fmt.Errorf("invalid result: %w", err)
	}
	return normalizeData(value), nil
}
func normalizeData(value any) any {
	switch v := value.(type) {
	case []any:
		for i := range v {
			v[i] = normalizeData(v[i])
		}
		return v
	case map[string]any:
		if typ, ok := v["@type"].(string); ok {
			inner := v["@value"]
			if typ == "g:Map" {
				result := map[string]any{}
				if pairs, ok := inner.([]any); ok {
					for i := 0; i+1 < len(pairs); i += 2 {
						result[fmt.Sprint(normalizeData(pairs[i]))] = normalizeData(pairs[i+1])
					}
				}
				return result
			}
			return normalizeData(inner)
		}
		if data, ok := v["data"]; ok {
			return normalizeData(data)
		}
		for k, item := range v {
			v[k] = normalizeData(item)
		}
		return v
	default:
		return value
	}
}
