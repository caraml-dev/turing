package utils

import "reflect"

// GetStructFieldsAsString returns a list of the struct's field values,
// if they can be successfully cast as strings
func GetStructFieldsAsString(object interface{}) []string {
	vals := getStructFieldValues(object)

	stringVals := []string{}
	for _, val := range vals {
		if stringVal, ok := val.(string); ok {
			stringVals = append(stringVals, stringVal)
		}
	}

	return stringVals
}

func getStructFieldValues(object interface{}) []interface{} {
	v := reflect.ValueOf(object)

	values := make([]interface{}, v.NumField())

	for i := 0; i < v.NumField(); i++ {
		values[i] = v.Field(i).Interface()
	}

	return values
}
