package desktop

import (
	"encoding/json"
	"fmt"
	"maps"
	"strings"
)

type GraphElement struct {
	Group string         `json:"group"`
	Data  map[string]any `json:"data"`
}

func ParseGraphSON(raw string) ([]GraphElement, error) {
	var parsed any
	if err := json.Unmarshal([]byte(raw), &parsed); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	collector := graphCollector{seen: make(map[string]struct{})}
	if err := collector.collect(parsed); err != nil {
		return nil, err
	}
	return collector.elements, nil
}

type graphCollector struct {
	elements []GraphElement
	seen     map[string]struct{}
}

func (c *graphCollector) collect(data any) error {
	switch v := data.(type) {
	case []any:
		for _, item := range v {
			if err := c.collect(item); err != nil {
				return err
			}
		}

	case map[string]any:
		if elem, ok := tryParseOpenCypherEntity(v); ok {
			c.appendUnique(elem)
			return nil
		}

		typeName, _ := v["@type"].(string)
		value := v["@value"]

		switch typeName {
		case "g:Vertex":
			elem, err := parseVertex(value)
			if err != nil {
				return err
			}
			c.appendUnique(elem)

		case "g:Edge":
			elem, err := parseEdge(value)
			if err != nil {
				return err
			}
			c.appendUnique(elem)

		case "g:List":
			if list, ok := value.([]any); ok {
				if err := c.collect(list); err != nil {
					return err
				}
			}

		case "g:Map":
			if mapArr, ok := value.([]any); ok {
				for i := 1; i < len(mapArr); i += 2 {
					if err := c.collect(mapArr[i]); err != nil {
						return err
					}
				}
			}

		case "g:Path":
			if pathMap, ok := value.(map[string]any); ok {
				if objects, ok := pathMap["objects"]; ok {
					if err := c.collect(objects); err != nil {
						return err
					}
				}
			}

		default:
			if elem, ok := tryParseUntypedEdge(v); ok {
				c.appendUnique(elem)
			} else if objects, ok := untypedPathObjects(v); ok {
				if err := c.collect(objects); err != nil {
					return err
				}
			} else if elem, ok := tryParseUntypedVertex(v); ok {
				c.appendUnique(elem)
			} else if dataVal, ok := v["data"]; ok {
				if err := c.collect(dataVal); err != nil {
					return err
				}
			} else {
				// Recurse into map values for project()/elementMap() results
				for _, val := range v {
					if err := c.collect(val); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func tryParseOpenCypherEntity(m map[string]any) (GraphElement, bool) {
	switch m["~entityType"] {
	case "node":
		id, ok := m["~id"]
		if !ok {
			return GraphElement{}, false
		}
		label := "Vertex"
		if labels, ok := m["~labels"].([]any); ok && len(labels) > 0 {
			label = fmt.Sprintf("%v", labels[0])
		} else if labels, ok := m["~labels"].([]string); ok && len(labels) > 0 {
			label = labels[0]
		}
		data := map[string]any{"id": fmt.Sprintf("%v", id), "label": label}
		if properties, ok := m["~properties"].(map[string]any); ok {
			maps.Copy(data, properties)
		}
		return GraphElement{Group: "nodes", Data: data}, true

	case "relationship":
		id, hasID := m["~id"]
		from, hasFrom := m["~start"]
		to, hasTo := m["~end"]
		if !hasID || !hasFrom || !hasTo {
			return GraphElement{}, false
		}
		data := map[string]any{
			"id":     fmt.Sprintf("%v", id),
			"source": fmt.Sprintf("%v", from),
			"target": fmt.Sprintf("%v", to),
			"label":  fmt.Sprintf("%v", m["~type"]),
		}
		if properties, ok := m["~properties"].(map[string]any); ok {
			maps.Copy(data, properties)
		}
		return GraphElement{Group: "edges", Data: data}, true
	}
	return GraphElement{}, false
}

func tryParseUntypedEdge(m map[string]any) (GraphElement, bool) {
	id, hasID := m["id"]
	inV, hasInV := m["inV"]
	outV, hasOutV := m["outV"]
	label, hasLabel := m["label"]
	if !hasID || !hasInV || !hasOutV || !hasLabel {
		return GraphElement{}, false
	}
	data := map[string]any{
		"id":     fmt.Sprintf("%v", id),
		"source": fmt.Sprintf("%v", outV),
		"target": fmt.Sprintf("%v", inV),
		"label":  label,
	}
	return GraphElement{Group: "edges", Data: data}, true
}

func untypedPathObjects(m map[string]any) ([]any, bool) {
	objects, ok := m["objects"]
	if !ok {
		return nil, false
	}
	arr, ok := objects.([]any)
	if !ok {
		return nil, false
	}
	return arr, true
}

func tryParseUntypedVertex(m map[string]any) (GraphElement, bool) {
	id, hasID := m["id"]
	label, hasLabel := m["label"]
	_, hasInV := m["inV"]
	if !hasID || !hasLabel || hasInV {
		return GraphElement{}, false
	}
	labelStr := fmt.Sprintf("%v", label)
	// elementMap() collision: a property named "label" shadows the vertex type
	// in plain JSON. Vertex type labels never contain spaces.
	if strings.Contains(labelStr, " ") {
		if it, ok := m["instanceType"].(string); ok && it != "" {
			labelStr = it
		}
	}
	data := map[string]any{
		"id":    fmt.Sprintf("%v", id),
		"label": labelStr,
	}
	if props, ok := m["properties"].(map[string]any); ok {
		for key, val := range props {
			if key == "id" || key == "label" {
				continue
			}
			data[key] = flattenUntypedProperty(val)
		}
	}
	// elementMap() puts properties flat alongside id/label
	for key, val := range m {
		if key == "id" || key == "label" || key == "properties" || key == "inV" || key == "outV" || key == "objects" {
			continue
		}
		if _, exists := data[key]; !exists {
			data[key] = val
		}
	}
	return GraphElement{Group: "nodes", Data: data}, true
}

func flattenUntypedProperty(val any) any {
	arr, ok := val.([]any)
	if !ok || len(arr) == 0 {
		return val
	}
	first, ok := arr[0].(map[string]any)
	if !ok {
		return arr[0]
	}
	if v, ok := first["value"]; ok {
		return v
	}
	return arr[0]
}

func parseVertex(value any) (GraphElement, error) {
	m, ok := value.(map[string]any)
	if !ok {
		return GraphElement{}, fmt.Errorf("vertex @value is not an object")
	}

	data := map[string]any{
		"id":    fmt.Sprintf("%v", m["id"]),
		"label": m["label"],
	}

	if props, ok := m["properties"].(map[string]any); ok {
		for key, val := range props {
			if key == "id" || key == "label" {
				continue
			}
			data[key] = flattenVertexProperty(val)
		}
	}

	return GraphElement{Group: "nodes", Data: data}, nil
}

func parseEdge(value any) (GraphElement, error) {
	m, ok := value.(map[string]any)
	if !ok {
		return GraphElement{}, fmt.Errorf("edge @value is not an object")
	}

	data := map[string]any{
		"id":     fmt.Sprintf("%v", m["id"]),
		"source": fmt.Sprintf("%v", m["outV"]),
		"target": fmt.Sprintf("%v", m["inV"]),
		"label":  m["label"],
	}

	if props, ok := m["properties"].(map[string]any); ok {
		for key, val := range props {
			data[key] = unwrapTypedValue(val)
		}
	}

	return GraphElement{Group: "edges", Data: data}, nil
}

func flattenVertexProperty(val any) any {
	arr, ok := val.([]any)
	if !ok || len(arr) == 0 {
		return unwrapTypedValue(val)
	}
	first, ok := arr[0].(map[string]any)
	if !ok {
		return unwrapTypedValue(arr[0])
	}
	if first["@type"] == "g:VertexProperty" {
		if inner, ok := first["@value"].(map[string]any); ok {
			return unwrapTypedValue(inner["value"])
		}
	}
	if v, ok := first["value"]; ok {
		return unwrapTypedValue(v)
	}
	return unwrapTypedValue(arr[0])
}

func unwrapTypedValue(val any) any {
	m, ok := val.(map[string]any)
	if !ok {
		return val
	}
	if _, hasType := m["@type"]; hasType {
		if inner, hasValue := m["@value"]; hasValue {
			return unwrapTypedValue(inner)
		}
	}
	return val
}

func (c *graphCollector) appendUnique(elem GraphElement) {
	id, _ := elem.Data["id"].(string)
	key := elem.Group + "\x00" + id
	if id == "" {
		return
	}
	if _, exists := c.seen[key]; exists {
		return
	}
	c.seen[key] = struct{}{}
	c.elements = append(c.elements, elem)
}

func parseValueMap(raw string) (map[string]any, error) {
	parsed, err := decodeData(raw)
	if err != nil {
		return nil, err
	}

	// Neptune returns [{...}] or {"data":[{...}], ...}
	switch v := parsed.(type) {
	case []any:
		if len(v) == 0 {
			return nil, nil
		}
		return flattenValueMapEntry(v[0])
	case map[string]any:
		if dataVal, ok := v["data"]; ok {
			if arr, ok := dataVal.([]any); ok && len(arr) > 0 {
				return flattenValueMapEntry(arr[0])
			}
		}
		return flattenValueMapEntry(v)
	}
	return nil, nil
}

func flattenValueMapEntry(entry any) (map[string]any, error) {
	m, ok := entry.(map[string]any)
	if !ok {
		return nil, nil
	}
	props := make(map[string]any)
	for key, val := range m {
		switch arr := val.(type) {
		case []any:
			if len(arr) == 1 {
				props[key] = arr[0]
			} else if len(arr) > 1 {
				props[key] = arr
			}
		default:
			props[key] = val
		}
	}
	return props, nil
}
