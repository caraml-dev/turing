package utils

import "encoding/json"

// MergeJSON adds (or overrides if such keys exist) key/value data from `overrides` map
// into the `message` json.RawMessage and returns merged json data
func MergeJSON(message json.RawMessage, overrides map[string]interface{}) (json.RawMessage, error) {
	var result map[string]interface{}
	if len(message) > 0 {
		err := json.Unmarshal(message, &result)

		if err != nil {
			return nil, err
		}

		result, err = MergeMaps(result, overrides)
		if err != nil {
			return nil, err
		}
	} else {
		result = overrides
	}

	return json.Marshal(result)
}
