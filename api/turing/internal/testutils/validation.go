package testutils

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/google/go-cmp/cmp"
)

// CompareObjects checks equality of 2 objects and returns a formatted error on failure.
// All nested unexported fields are ignored. See DeepAllowUnexported for more details.
func CompareObjects(actual interface{}, expected interface{}) error {
	allowUnexportedOn := actual
	if reflect.TypeOf(allowUnexportedOn).Kind() == reflect.Ptr {
		allowUnexportedOn = reflect.ValueOf(actual).Elem().Interface()
	}
	if !cmp.Equal(actual, expected, DeepAllowUnexported(allowUnexportedOn, expected)) {
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

// DeepAllowUnexported allows nested unexported fields to be allowed. This compare function is found here:
// https://github.com/google/go-cmp/issues/40#issuecomment-328615283
func DeepAllowUnexported(vs ...interface{}) cmp.Option {
	m := make(map[reflect.Type]struct{})
	for _, v := range vs {
		structTypes(reflect.ValueOf(v), m)
	}
	var typs []interface{}
	for t := range m {
		typs = append(typs, reflect.New(t).Elem().Interface())
	}
	return cmp.AllowUnexported(typs...)
}

func structTypes(v reflect.Value, m map[reflect.Type]struct{}) {
	if !v.IsValid() {
		return
	}
	switch v.Kind() {
	case reflect.Ptr:
		if !v.IsNil() {
			structTypes(v.Elem(), m)
		}
	case reflect.Interface:
		if !v.IsNil() {
			structTypes(v.Elem(), m)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			structTypes(v.Index(i), m)
		}
	case reflect.Map:
		for _, k := range v.MapKeys() {
			structTypes(v.MapIndex(k), m)
		}
	case reflect.Struct:
		m[v.Type()] = struct{}{}
		for i := 0; i < v.NumField(); i++ {
			structTypes(v.Field(i), m)
		}
	default:
	}
}
