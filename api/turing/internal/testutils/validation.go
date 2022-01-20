package testutils

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/google/go-cmp/cmp"
)

// CompareObjects checks equality of 2 objects and returns a formatted error on failure.
// If the object types have unexported fields, a custom marshaler is required to be defined.
func CompareObjects(actual interface{}, expected interface{}) error {
	allowUnexportedOn := actual
	if reflect.TypeOf(allowUnexportedOn).Kind() == reflect.Ptr {
		allowUnexportedOn = reflect.ValueOf(actual).Elem().Interface()
	}
	if !cmp.Equal(actual, expected, cmp.AllowUnexported(allowUnexportedOn)) {
		actualString := fmt.Sprintf("%+v", actual)
		expectedString := fmt.Sprintf("%+v", expected)

		// Attempt to encode values to JSON, for logging
		jsonActual, err := json.Marshal(actual)
		if err == nil {
			actualString = string(jsonActual)
		}
		jsonExpected, err := json.Marshal(expected)
		if err == nil {
			expectedString = string(jsonExpected)
		}

		return fmt.Errorf("Did not get expected configuration.\nEXPECTED:\n%v\nGOT:\n%v",
			expectedString, actualString)
	}
	return nil
}
