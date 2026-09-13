package desktop

import (
	"encoding/json"
	"fmt"
	"strings"
)

func gremlinString(value string) string {
	return "'" + strings.NewReplacer(
		"\\", "\\\\",
		"'", "\\'",
		"\n", "\\n",
		"\r", "\\r",
		"\t", "\\t",
	).Replace(value) + "'"
}

// decodeData normalizes typed GraphSON and the REST API's plain JSON envelope.
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
	switch value := value.(type) {
	case []any:
		for index := range value {
			value[index] = normalizeData(value[index])
		}
		return value
	case map[string]any:
		if graphSONType, ok := value["@type"].(string); ok {
			inner := value["@value"]
			if graphSONType == "g:Map" {
				result := map[string]any{}
				if pairs, ok := inner.([]any); ok {
					for index := 0; index+1 < len(pairs); index += 2 {
						result[fmt.Sprint(normalizeData(pairs[index]))] = normalizeData(pairs[index+1])
					}
				}
				return result
			}
			return normalizeData(inner)
		}
		if data, ok := value["data"]; ok {
			return normalizeData(data)
		}
		for key, item := range value {
			value[key] = normalizeData(item)
		}
		return value
	default:
		return value
	}
}
