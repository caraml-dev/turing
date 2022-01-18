package utils

import "encoding/json"

// MergeJSON adds (or overrides if such keys exist) key/value data from `overrides` map
// into the `message` json.RawMessage and returns merged json data
func MergeJSON(message json.RawMessage, overrides map[string]interface{}) (json.RawMessage, error) {
	var original map[string]interface{}
	err := json.Unmarshal(message, &original)

	if err != nil {
		return nil, err
	}

	mergedMaps, err := MergeMaps(original, overrides)
	if err != nil {
		return nil, err
	}

	return json.Marshal(mergedMaps)
}
