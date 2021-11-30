package utils

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v3"
)

// MergeTwoYamls reads the original yaml file and overrides the original file with
// the override file. This overriding will follow the rules in MergeMaps.
func MergeTwoYamls(originalYAMLFile, overrideYAMLFile string) error {
	original, err := readYAML(originalYAMLFile)
	if err != nil {
		return fmt.Errorf("error reading original yaml: %s", err)
	}

	override, err := readYAML(overrideYAMLFile)
	if err != nil {
		return fmt.Errorf("error reading override yaml: %s", err)
	}

	merged, err := MergeMaps(original, override)
	if err != nil {
		return fmt.Errorf("error merging maps: %s", err)
	}

	var output bytes.Buffer
	yamlEncoder := yaml.NewEncoder(&output)
	yamlEncoder.SetIndent(2)
	err = yamlEncoder.Encode(merged)
	if err != nil {
		return fmt.Errorf("error encoding: %s", err)
	}

	f, err := os.OpenFile(originalYAMLFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("error writing to file: %s", err)
	}

	w := bufio.NewWriter(f)

	if _, err = w.Write(output.Bytes()); err != nil {
		return fmt.Errorf("error writing: %s", err)
	}

	if err = w.Flush(); err != nil {
		return fmt.Errorf("error flushing: %s", err)
	}

	return nil
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
	file, err := ioutil.ReadFile(filepath)
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
