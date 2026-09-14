package neptune

import (
	"bytes"
	"encoding/json"
)

func unwrapGraphSON(raw []byte) (string, bool) {
	if !bytes.Contains(raw, []byte(`"@type"`)) {
		return "", false
	}
	var parsed any
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return "", false
	}

	unwrapped, changed := unwrapGraphSONValue(parsed)
	if !changed {
		return "", false
	}

	content, err := json.Marshal(unwrapped)
	if err != nil {
		return "", false
	}
	return string(content), true
}

func unwrapGraphSONValue(value any) (any, bool) {
	switch v := value.(type) {
	case []any:
		changed := false
		for i, item := range v {
			v[i], changed = unwrapGraphSONValueChanged(item, changed)
		}
		return v, changed
	case map[string]any:
		typeName, hasType := v["@type"].(string)
		if hasType {
			if inner, ok := v["@value"]; ok {
				return unwrapTypedGraphSONValue(typeName, inner)
			}
		}

		changed := false
		for key, item := range v {
			v[key], changed = unwrapGraphSONValueChanged(item, changed)
		}
		return v, changed
	default:
		return value, false
	}
}

func unwrapGraphSONValueChanged(value any, changed bool) (any, bool) {
	unwrapped, childChanged := unwrapGraphSONValue(value)
	return unwrapped, changed || childChanged
}

func unwrapTypedGraphSONValue(typeName string, value any) (any, bool) {
	switch typeName {
	case "g:List", "g:Set", "g:BulkSet":
		unwrapped, _ := unwrapGraphSONValue(value)
		return unwrapped, true
	case "g:Map":
		return unwrapGraphSONMap(value), true
	case "g:T":
		return value, true
	case "g:VertexProperty":
		if property, ok := value.(map[string]any); ok {
			if propertyValue, ok := property["value"]; ok {
				unwrapped, _ := unwrapGraphSONValue(propertyValue)
				return unwrapped, true
			}
		}
		unwrapped, _ := unwrapGraphSONValue(value)
		return unwrapped, true
	case "g:Property":
		if property, ok := value.(map[string]any); ok {
			if propertyValue, ok := property["value"]; ok {
				unwrapped, _ := unwrapGraphSONValue(propertyValue)
				return unwrapped, true
			}
		}
		unwrapped, _ := unwrapGraphSONValue(value)
		return unwrapped, true
	case "g:Path":
		unwrapped, _ := unwrapGraphSONValue(value)
		return unwrapped, true
	case "g:Vertex", "g:Edge":
		unwrapped, _ := unwrapGraphSONValue(value)
		return unwrapped, true
	default:
		unwrapped, _ := unwrapGraphSONValue(value)
		return unwrapped, true
	}
}

func unwrapGraphSONMap(value any) any {
	items, ok := value.([]any)
	if !ok {
		unwrapped, _ := unwrapGraphSONValue(value)
		return unwrapped
	}

	result := make(map[string]any, len(items)/2)
	for i := 0; i+1 < len(items); i += 2 {
		key, _ := unwrapGraphSONValue(items[i])
		keyString, ok := key.(string)
		if !ok || keyString == "" {
			continue
		}
		result[keyString], _ = unwrapGraphSONValue(items[i+1])
	}
	return result
}
