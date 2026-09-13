package desktop

import (
	"fmt"
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
	elements, warning := boundGraphWithExternal(elements, map[string]bool{req.ID: true})
	return GraphResponse{Elements: elements, Warning: warning}
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
	vertexStep, edgeStep := "both", "bothE"
	switch direction {
	case "in":
		vertexStep, edgeStep = "in", "inE"
	case "out":
		vertexStep, edgeStep = "out", "outE"
	}

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

	return fmt.Sprintf("MATCH %s WHERE %s WITH DISTINCT neighbor LIMIT %d MATCH (source)-[edge]-(neighbor) RETURN neighbor, edge", pattern, strings.Join(conditions, " AND "), limit)
}

func cypherString(value string) string {
	return `"` + strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", `\n`, "\r", `\r`).Replace(value) + `"`
}

func cypherIdentifier(value string) string {
	return "`" + strings.ReplaceAll(value, "`", "``") + "`"
}
