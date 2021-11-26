package utils

// MergeMaps takes two maps with any value and merges it recursively.
// if the underlying value is also a map[string]interface{} it will replace only the values in that map
// but if the value is any other format, it will replace it with the override value
func MergeMaps(originalMap, override map[string]interface{}) (map[string]interface{}, error) {
	// Copy map over so we don't have any side effects to original map
	original := make(map[string]interface{})
	for k, v := range originalMap {
		original[k] = v
	}

	// Iterate over merged then add/replace
	for key, value := range override {
		if castedMap, ok := value.(map[string]interface{}); ok {
			originalVal, ok := original[key]
			if !ok {
				// key not found, add to map
				original[key] = castedMap
				continue
			}

			originalValMap, ok := originalVal.(map[string]interface{})
			if !ok {
				// value was something else, replace the value with map instead
				original[key] = castedMap
				continue
			}

			originalValMap, err := MergeMaps(originalValMap, castedMap)
			if err != nil {
				return nil, err
			}
			original[key] = originalValMap
		} else {
			original[key] = value
		}
	}
	return original, nil
}
