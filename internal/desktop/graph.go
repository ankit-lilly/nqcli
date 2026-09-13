package desktop

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/ankit-lilly/nqcli/internal/core"
)

const (
	defaultExpansionLimit = 10
	maxExpansionLimit     = 100
)

type ExpandVertexRequest struct {
	ID                string   `json:"id"`
	Type              string   `json:"type"`
	Direction         string   `json:"direction"`
	Relationship      string   `json:"relationship"`
	NeighborLabel     string   `json:"neighborLabel"`
	Limit             int      `json:"limit"`
	ExcludedVertexIDs []string `json:"excludedVertexIds"`
}

type NeighborSummaryRequest struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Direction string `json:"direction"`
}

type NeighborOption struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type NeighborSummaryResponse struct {
	Nodes         []NeighborOption `json:"nodes"`
	Relationships []NeighborOption `json:"relationships"`
	Error         string           `json:"error,omitempty"`
}

func (d *DesktopService) ExpandVertex(req ExpandVertexRequest) GraphResponse {
	query, queryType, err := buildExpandVertexQuery(req)
	if err != nil {
		return GraphResponse{Error: err.Error()}
	}

	d.mu.RLock()
	svc := d.svc
	ctx := d.ctx
	d.mu.RUnlock()

	result, err := svc.ExecuteQuery(ctx, query, queryType, core.QueryOpts{
		SkipFormatting:   true,
		MaxResponseBytes: 8 << 20,
	})
	if err != nil {
		return GraphResponse{Error: err.Error()}
	}

	content := result.Content
	if content == "" {
		content = result.Raw
	}
	elements, err := ParseGraphSON(content)
	if err != nil {
		return GraphResponse{Error: fmt.Sprintf("graph parse error: %v", err)}
	}
	elements, warning := boundGraphWithExternal(elements, map[string]bool{strings.TrimSpace(req.ID): true})
	return GraphResponse{Elements: elements, Warning: warning}
}

func (d *DesktopService) GetNeighborSummary(req NeighborSummaryRequest) NeighborSummaryResponse {
	query, queryType, err := buildNeighborSummaryQuery(req)
	if err != nil {
		return NeighborSummaryResponse{Error: err.Error()}
	}

	d.mu.RLock()
	svc := d.svc
	ctx := d.ctx
	d.mu.RUnlock()

	result, err := svc.ExecuteQuery(ctx, query, queryType, core.QueryOpts{
		SkipFormatting:   true,
		MaxResponseBytes: 2 << 20,
	})
	if err != nil {
		return NeighborSummaryResponse{Error: err.Error()}
	}

	content := result.Content
	if content == "" {
		content = result.Raw
	}
	summary, err := parseNeighborSummary(content)
	if err != nil {
		return NeighborSummaryResponse{Error: fmt.Sprintf("neighbor summary parse error: %v", err)}
	}
	return summary
}

func buildExpandVertexQuery(req ExpandVertexRequest) (string, string, error) {
	req.ID = strings.TrimSpace(req.ID)
	if req.ID == "" {
		return "", "", fmt.Errorf("vertex ID is required")
	}

	queryType := strings.ToLower(strings.TrimSpace(req.Type))
	if queryType == "" {
		queryType = "gremlin"
	}
	if queryType != "gremlin" && queryType != "cypher" {
		return "", "", fmt.Errorf("unsupported query type %q", req.Type)
	}

	direction := strings.ToLower(strings.TrimSpace(req.Direction))
	if direction == "" {
		direction = "both"
	}
	if direction != "both" && direction != "in" && direction != "out" {
		return "", "", fmt.Errorf("unsupported direction %q", req.Direction)
	}

	limit := req.Limit
	if limit <= 0 {
		limit = defaultExpansionLimit
	}
	if limit > maxExpansionLimit {
		limit = maxExpansionLimit
	}

	if queryType == "cypher" {
		return buildCypherExpansion(req, direction, limit), queryType, nil
	}
	return buildGremlinExpansion(req, direction, limit), queryType, nil
}

func buildGremlinExpansion(req ExpandVertexRequest, direction string, limit int) string {
	vertexStep, edgeStep := traversalSteps(direction)

	labels := ""
	edgeLabels := ""
	if label := strings.TrimSpace(req.Relationship); label != "" {
		labels = gremlinString(label)
		edgeLabels = labels
	}

	query := "g.V(" + gremlinString(req.ID) + ").as('start')." + vertexStep + "(" + labels + ")"
	if label := strings.TrimSpace(req.NeighborLabel); label != "" {
		query += ".hasLabel(" + gremlinString(label) + ")"
	}
	if len(req.ExcludedVertexIDs) > 0 {
		ids := make([]string, 0, len(req.ExcludedVertexIDs))
		for _, id := range req.ExcludedVertexIDs {
			if id = strings.TrimSpace(id); id != "" && id != req.ID {
				ids = append(ids, gremlinString(id))
			}
		}
		if len(ids) > 0 {
			query += ".filter(__.not(__.hasId(" + strings.Join(ids, ",") + ")))"
		}
	}

	query += fmt.Sprintf(".dedup().range(0,%d).as('neighbor').project('vertex','edges').by().by(__.select('start').%s(%s).where(otherV().where(eq('neighbor'))).dedup().fold())", limit, edgeStep, edgeLabels)
	return query
}

func buildNeighborSummaryQuery(req NeighborSummaryRequest) (string, string, error) {
	id := strings.TrimSpace(req.ID)
	if id == "" {
		return "", "", fmt.Errorf("vertex ID is required")
	}
	queryType := strings.ToLower(strings.TrimSpace(req.Type))
	if queryType == "" {
		queryType = "gremlin"
	}
	if queryType != "gremlin" && queryType != "cypher" {
		return "", "", fmt.Errorf("unsupported query type %q", req.Type)
	}
	direction := strings.ToLower(strings.TrimSpace(req.Direction))
	if direction == "" {
		direction = "both"
	}
	if direction != "both" && direction != "in" && direction != "out" {
		return "", "", fmt.Errorf("unsupported direction %q", req.Direction)
	}

	if queryType == "cypher" {
		return buildCypherNeighborSummary(id, direction), queryType, nil
	}
	vertexStep, edgeStep := traversalSteps(direction)
	query := fmt.Sprintf(
		"g.V(%s).project('nodes','relationships').by(%s().dedup().groupCount().by(label)).by(%s().group().by(label).by(otherV().dedup().count()))",
		gremlinString(id), vertexStep, edgeStep,
	)
	return query, queryType, nil
}

func buildCypherNeighborSummary(id, direction string) string {
	pattern := "(source)-[edge]-(neighbor)"
	switch direction {
	case "in":
		pattern = "(source)<-[edge]-(neighbor)"
	case "out":
		pattern = "(source)-[edge]->(neighbor)"
	}
	condition := "ID(source) = " + cypherString(id)
	return fmt.Sprintf(
		"MATCH %s WHERE %s UNWIND labels(neighbor) AS label RETURN 'node' AS kind, label, count(DISTINCT neighbor) AS count UNION ALL MATCH %s WHERE %s RETURN 'relationship' AS kind, type(edge) AS label, count(DISTINCT neighbor) AS count",
		pattern, condition, pattern, condition,
	)
}

func traversalSteps(direction string) (string, string) {
	switch direction {
	case "in":
		return "in", "inE"
	case "out":
		return "out", "outE"
	default:
		return "both", "bothE"
	}
}

func parseNeighborSummary(raw string) (NeighborSummaryResponse, error) {
	value, err := decodeData(raw)
	if err != nil {
		return NeighborSummaryResponse{}, err
	}

	nodes := map[string]int{}
	relationships := map[string]int{}
	collectNeighborCounts(value, nodes, relationships)
	return NeighborSummaryResponse{
		Nodes:         sortedNeighborOptions(nodes),
		Relationships: sortedNeighborOptions(relationships),
	}, nil
}

func collectNeighborCounts(value any, nodes, relationships map[string]int) {
	switch value := value.(type) {
	case []any:
		for _, item := range value {
			collectNeighborCounts(item, nodes, relationships)
		}
	case map[string]any:
		if counts, ok := value["nodes"].(map[string]any); ok {
			mergeNeighborCounts(nodes, counts)
		}
		if counts, ok := value["relationships"].(map[string]any); ok {
			mergeNeighborCounts(relationships, counts)
		}
		if kind, _ := value["kind"].(string); kind != "" {
			label, _ := value["label"].(string)
			if count, ok := neighborCount(value["count"]); ok && label != "" {
				if kind == "node" {
					nodes[label] += count
				} else if kind == "relationship" {
					relationships[label] += count
				}
			}
		}
		for key, item := range value {
			if key != "nodes" && key != "relationships" && key != "kind" && key != "label" && key != "count" {
				collectNeighborCounts(item, nodes, relationships)
			}
		}
	}
}

func mergeNeighborCounts(destination map[string]int, counts map[string]any) {
	for label, value := range counts {
		if count, ok := neighborCount(value); ok && label != "" {
			destination[label] += count
		}
	}
}

func neighborCount(value any) (int, bool) {
	switch value := value.(type) {
	case json.Number:
		count, err := value.Int64()
		return int(count), err == nil
	case float64:
		return int(value), true
	case int:
		return value, true
	case int64:
		return int(value), true
	default:
		return 0, false
	}
}

func sortedNeighborOptions(counts map[string]int) []NeighborOption {
	options := make([]NeighborOption, 0, len(counts))
	for label, count := range counts {
		options = append(options, NeighborOption{Label: label, Count: count})
	}
	sort.Slice(options, func(i, j int) bool { return options[i].Label < options[j].Label })
	return options
}

func buildCypherExpansion(req ExpandVertexRequest, direction string, limit int) string {
	edge := "[edge]"
	if label := strings.TrimSpace(req.Relationship); label != "" {
		edge = "[edge:" + cypherIdentifier(label) + "]"
	}

	pattern := "(source)-" + edge + "-(neighbor)"
	switch direction {
	case "in":
		pattern = "(source)<-" + edge + "-(neighbor)"
	case "out":
		pattern = "(source)-" + edge + "->(neighbor)"
	}
	if label := strings.TrimSpace(req.NeighborLabel); label != "" {
		pattern = strings.Replace(pattern, "(neighbor)", "(neighbor:"+cypherIdentifier(label)+")", 1)
	}

	conditions := []string{"ID(source) = " + cypherString(req.ID)}
	if len(req.ExcludedVertexIDs) > 0 {
		ids := make([]string, 0, len(req.ExcludedVertexIDs))
		for _, id := range req.ExcludedVertexIDs {
			if id = strings.TrimSpace(id); id != "" && id != req.ID {
				ids = append(ids, cypherString(id))
			}
		}
		if len(ids) > 0 {
			conditions = append(conditions, "NOT ID(neighbor) IN ["+strings.Join(ids, ",")+"]")
		}
	}

	return fmt.Sprintf("MATCH %s WHERE %s WITH DISTINCT source, neighbor LIMIT %d MATCH %s RETURN neighbor, edge", pattern, strings.Join(conditions, " AND "), limit, pattern)
}

func cypherString(value string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`).Replace(value) + `"`
}

func cypherIdentifier(value string) string {
	return "`" + strings.ReplaceAll(value, "`", "``") + "`"
}
