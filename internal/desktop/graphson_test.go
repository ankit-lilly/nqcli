package desktop

import (
	"testing"
)

func TestParseGraphSON_VertexList(t *testing.T) {
	t.Parallel()
	raw := `[
		{
			"@type": "g:Vertex",
			"@value": {
				"id": "v1",
				"label": "Study",
				"properties": {
					"name": [{"@type": "g:VertexProperty", "@value": {"id": "p1", "value": "ABC-123", "label": "name"}}]
				}
			}
		}
	]`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	if elems[0].Group != "nodes" {
		t.Fatalf("expected group 'nodes', got %q", elems[0].Group)
	}
	if elems[0].Data["id"] != "v1" {
		t.Fatalf("expected id 'v1', got %v", elems[0].Data["id"])
	}
	if elems[0].Data["label"] != "Study" {
		t.Fatalf("expected label 'Study', got %v", elems[0].Data["label"])
	}
	if elems[0].Data["name"] != "ABC-123" {
		t.Fatalf("expected name 'ABC-123', got %v", elems[0].Data["name"])
	}
}

func TestParseGraphSON_EdgeList(t *testing.T) {
	t.Parallel()
	raw := `[
		{
			"@type": "g:Edge",
			"@value": {
				"id": "e1",
				"label": "has_version",
				"outV": "v1",
				"inV": "v2",
				"outVLabel": "Study",
				"inVLabel": "StudyVersion"
			}
		}
	]`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	if elems[0].Group != "edges" {
		t.Fatalf("expected group 'edges', got %q", elems[0].Group)
	}
	if elems[0].Data["source"] != "v1" {
		t.Fatalf("expected source 'v1', got %v", elems[0].Data["source"])
	}
	if elems[0].Data["target"] != "v2" {
		t.Fatalf("expected target 'v2', got %v", elems[0].Data["target"])
	}
	if elems[0].Data["label"] != "has_version" {
		t.Fatalf("expected label 'has_version', got %v", elems[0].Data["label"])
	}
}

func TestParseGraphSON_MixedVerticesAndEdges(t *testing.T) {
	t.Parallel()
	raw := `[
		{"@type": "g:Vertex", "@value": {"id": "v1", "label": "Study"}},
		{"@type": "g:Edge", "@value": {"id": "e1", "label": "has_version", "outV": "v1", "inV": "v2"}},
		{"@type": "g:Vertex", "@value": {"id": "v2", "label": "StudyVersion"}}
	]`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(elems))
	}
	if elems[0].Group != "nodes" || elems[1].Group != "edges" || elems[2].Group != "nodes" {
		t.Fatalf("unexpected groups: %s, %s, %s", elems[0].Group, elems[1].Group, elems[2].Group)
	}
}

func TestParseGraphSON_NestedGList(t *testing.T) {
	t.Parallel()
	raw := `{
		"@type": "g:List",
		"@value": [
			{"@type": "g:Vertex", "@value": {"id": "v1", "label": "Study"}}
		]
	}`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
}

func TestParseGraphSON_Path(t *testing.T) {
	t.Parallel()
	raw := `{
		"@type": "g:Path",
		"@value": {
			"labels": {"@type": "g:List", "@value": []},
			"objects": {"@type": "g:List", "@value": [
				{"@type": "g:Vertex", "@value": {"id": "v1", "label": "Study"}},
				{"@type": "g:Edge", "@value": {"id": "e1", "label": "has_version", "outV": "v1", "inV": "v2"}},
				{"@type": "g:Vertex", "@value": {"id": "v2", "label": "StudyVersion"}}
			]}
		}
	}`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(elems))
	}
}

func TestParseGraphSON_DeduplicatesById(t *testing.T) {
	t.Parallel()
	raw := `[
		{"@type": "g:Vertex", "@value": {"id": "v1", "label": "Study"}},
		{"@type": "g:Vertex", "@value": {"id": "v1", "label": "Study"}}
	]`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 1 {
		t.Fatalf("expected 1 deduplicated element, got %d", len(elems))
	}
}

func TestParseGraphSON_EmptyArray(t *testing.T) {
	t.Parallel()
	elems, err := ParseGraphSON("[]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 0 {
		t.Fatalf("expected 0 elements, got %d", len(elems))
	}
}

func TestParseGraphSON_PlainJSON(t *testing.T) {
	t.Parallel()
	raw := `[{"id": "1", "label": "Study", "properties": {"name": [{"value": "ABC"}]}}]`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 1 {
		t.Fatalf("expected 1 element from untyped vertex, got %d", len(elems))
	}
	if elems[0].Group != "nodes" {
		t.Fatalf("expected group 'nodes', got %q", elems[0].Group)
	}
	if elems[0].Data["name"] != "ABC" {
		t.Fatalf("expected name 'ABC', got %v", elems[0].Data["name"])
	}
}

func TestParseGraphSON_TypedScalars(t *testing.T) {
	t.Parallel()
	raw := `[
		{
			"@type": "g:Vertex",
			"@value": {
				"id": "v1",
				"label": "Study",
				"properties": {
					"count": [{"@type": "g:VertexProperty", "@value": {"id": "p1", "value": {"@type": "g:Int32", "@value": 42}, "label": "count"}}]
				}
			}
		}
	]`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	count, ok := elems[0].Data["count"].(float64)
	if !ok {
		t.Fatalf("expected count to be float64, got %T: %v", elems[0].Data["count"], elems[0].Data["count"])
	}
	if count != 42 {
		t.Fatalf("expected count 42, got %v", count)
	}
}

func TestParseGraphSON_GMap(t *testing.T) {
	t.Parallel()
	raw := `{
		"@type": "g:Map",
		"@value": [
			"vertex",
			{"@type": "g:Vertex", "@value": {"id": "v1", "label": "Study"}}
		]
	}`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	if elems[0].Data["id"] != "v1" {
		t.Fatalf("expected id 'v1', got %v", elems[0].Data["id"])
	}
}

func TestParseGraphSON_UntypedVertexFromNeptune(t *testing.T) {
	t.Parallel()
	raw := `[{
		"id": "56cd5d13-d110-bed0-27f9-31e6f7010ed5",
		"label": "Code",
		"properties": {
			"code": [{"id": -258375936, "label": "code", "value": "C98388"}],
			"decode": [{"id": 692826160, "label": "decode", "value": "INTERVENTIONAL"}]
		}
	}]`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	if elems[0].Group != "nodes" {
		t.Fatalf("expected group 'nodes', got %q", elems[0].Group)
	}
	if elems[0].Data["code"] != "C98388" {
		t.Fatalf("expected code 'C98388', got %v", elems[0].Data["code"])
	}
	if elems[0].Data["decode"] != "INTERVENTIONAL" {
		t.Fatalf("expected decode 'INTERVENTIONAL', got %v", elems[0].Data["decode"])
	}
}

func TestParseGraphSON_UntypedEdge(t *testing.T) {
	t.Parallel()
	raw := `[{
		"id": "e1",
		"label": "has_version",
		"type": "edge",
		"inV": "v2",
		"outV": "v1",
		"inVLabel": "StudyVersion",
		"outVLabel": "Study"
	}]`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	if elems[0].Group != "edges" {
		t.Fatalf("expected group 'edges', got %q", elems[0].Group)
	}
	if elems[0].Data["source"] != "v1" {
		t.Fatalf("expected source 'v1', got %v", elems[0].Data["source"])
	}
	if elems[0].Data["target"] != "v2" {
		t.Fatalf("expected target 'v2', got %v", elems[0].Data["target"])
	}
}

func TestParseGraphSON_OpenCypherEntities(t *testing.T) {
	t.Parallel()
	raw := `{"results":[{"neighbor":{"~id":"v2","~entityType":"node","~labels":["StudyVersion"],"~properties":{"name":"Version 2"}},"edge":{"~id":"e1","~entityType":"relationship","~start":"v1","~end":"v2","~type":"has_version","~properties":{"active":true}}}]}`

	elements, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elements) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(elements))
	}

	byID := make(map[string]GraphElement, len(elements))
	for _, element := range elements {
		byID[element.Data["id"].(string)] = element
	}
	if got := byID["v2"].Data["label"]; got != "StudyVersion" {
		t.Fatalf("expected node label StudyVersion, got %v", got)
	}
	if got := byID["v2"].Data["name"]; got != "Version 2" {
		t.Fatalf("expected node property, got %v", got)
	}
	if got := byID["e1"].Data["source"]; got != "v1" {
		t.Fatalf("expected edge source v1, got %v", got)
	}
	if got := byID["e1"].Data["active"]; got != true {
		t.Fatalf("expected edge property, got %v", got)
	}
}

func TestParseGraphSON_NonGraphObject(t *testing.T) {
	t.Parallel()
	raw := `{"message": "hello"}`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 0 {
		t.Fatalf("expected 0 elements from non-graph object, got %d", len(elems))
	}
}

func TestParseGraphSON_DataWithMetaEnvelope(t *testing.T) {
	t.Parallel()
	raw := `{
		"data": [{
			"id": "v1",
			"label": "StudyVersion",
			"properties": {
				"versionIdentifier": [{"id": 1, "label": "versionIdentifier", "value": "1.0.0"}]
			}
		}],
		"meta": {}
	}`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 1 {
		t.Fatalf("expected 1 element, got %d", len(elems))
	}
	if elems[0].Group != "nodes" {
		t.Fatalf("expected group 'nodes', got %q", elems[0].Group)
	}
	if elems[0].Data["versionIdentifier"] != "1.0.0" {
		t.Fatalf("expected versionIdentifier '1.0.0', got %v", elems[0].Data["versionIdentifier"])
	}
}

func TestParseGraphSON_UntypedPath(t *testing.T) {
	t.Parallel()
	raw := `[{
		"labels": [[], [], [], []],
		"objects": [
			{"id": "v1", "label": "Study"},
			{"id": "e1", "label": "has_version", "inV": "v2", "outV": "v1", "inVLabel": "StudyVersion", "outVLabel": "Study"},
			{"id": "v2", "label": "StudyVersion"},
			{"id": "e2", "label": "has_design", "inV": "v3", "outV": "v2", "inVLabel": "StudyDesign", "outVLabel": "StudyVersion"},
			{"id": "v3", "label": "StudyDesign"}
		]
	}]`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	nodes := 0
	edges := 0
	for _, e := range elems {
		if e.Group == "nodes" {
			nodes++
		} else {
			edges++
		}
	}
	if nodes != 3 {
		t.Fatalf("expected 3 nodes, got %d", nodes)
	}
	if edges != 2 {
		t.Fatalf("expected 2 edges, got %d", edges)
	}
}

func TestParseGraphSON_ProjectElementMap(t *testing.T) {
	t.Parallel()
	raw := `[{
		"study": {"id": "v1", "label": "Study", "name": "ABC-123", "createdAt": "2025-01-01"},
		"versions": [
			{"id": "v2", "label": "StudyVersion", "versionIdentifier": "1.0.0"},
			{"id": "v3", "label": "StudyVersion", "versionIdentifier": "2.0.0"}
		]
	}]`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 3 {
		t.Fatalf("expected 3 elements, got %d", len(elems))
	}
	for _, e := range elems {
		if e.Group != "nodes" {
			t.Fatalf("expected all nodes, got %q", e.Group)
		}
	}
	// Verify flat properties from elementMap are captured
	for _, e := range elems {
		if e.Data["id"] == "v1" {
			if e.Data["name"] != "ABC-123" {
				t.Fatalf("expected name 'ABC-123', got %v", e.Data["name"])
			}
		}
	}
}

func TestParseGraphSON_ElementMapLabelCollision(t *testing.T) {
	t.Parallel()
	raw := `[{
		"study": {
			"id": "XKA-IW-NFHZ",
			"label": "A Phase IV, Observational Study of VIM-001",
			"instanceType": "Study",
			"name": "XKA-IW-NFHZ"
		},
		"versions": [{
			"id": "v2",
			"label": "StudyVersion",
			"instanceType": "StudyVersion",
			"versionIdentifier": "1.0.0"
		}]
	}]`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(elems))
	}
	for _, e := range elems {
		if e.Data["id"] == "XKA-IW-NFHZ" {
			if e.Data["label"] != "Study" {
				t.Fatalf("expected label 'Study' (from instanceType), got %v", e.Data["label"])
			}
		}
		if e.Data["id"] == "v2" {
			if e.Data["label"] != "StudyVersion" {
				t.Fatalf("expected label 'StudyVersion', got %v", e.Data["label"])
			}
		}
	}
}

func TestParseGraphSON_GroupCountNoFalseVertices(t *testing.T) {
	t.Parallel()
	raw := `[{"Study": 5, "StudyVersion": 10, "Timing": 3}]`
	elems, err := ParseGraphSON(raw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(elems) != 0 {
		t.Fatalf("expected 0 elements from groupCount, got %d", len(elems))
	}
}
