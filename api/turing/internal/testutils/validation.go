package testutils

import (
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
		return fmt.Errorf(cmp.Diff(actual, expected, cmp.AllowUnexported(allowUnexportedOn)))
	}
	return nil
}
