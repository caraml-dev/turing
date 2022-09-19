package utils

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// MergeTwoYamls reads the original yaml file and overrides the original file with
// the override file. This overriding will follow the rules in MergeMaps.
func MergeTwoYamls(originalYAMLFile string, overrideYAMLFile *string) ([]byte, error) {
	original, err := readYAML(originalYAMLFile)
	if err != nil {
		return nil, fmt.Errorf("error reading original yaml: %s", err)
	}

	// If override file is empty, we just have to leave the override as an empty map
	var override map[string]interface{}
	if overrideYAMLFile != nil {
		var err error
		override, err = readYAML(*overrideYAMLFile)
		if err != nil {
			return nil, fmt.Errorf("error reading override yaml: %s", err)
		}
	}

	merged, err := MergeMaps(original, override)
	if err != nil {
		return nil, fmt.Errorf("error merging maps: %s", err)
	}

	var output bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&output)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(merged)
	if err != nil {
		return nil, fmt.Errorf("error encoding: %s", err)
	}

	return output.Bytes(), nil
}

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

func readYAML(filepath string) (map[string]interface{}, error) {
	file, err := os.ReadFile(filepath)
	if err != nil {
		return nil, err
	}

	y := make(map[string]interface{})
	err = yaml.Unmarshal(file, y)
	if err != nil {
		return nil, err
	}

	return y, nil
}

// WriteYAMLFile takes a byte slice and writes to the path.
func WriteYAMLFile(b []byte, path string) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("error opening file for writing: %s", err)
	}

	_, err = f.Write(b)
	if err != nil {
		return fmt.Errorf("error writing bytes into file: %s", err)
	}

	err = f.Sync()
	if err != nil {
		return fmt.Errorf("error syncing file to disk: %s", err)
	}

	return nil
}
