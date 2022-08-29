package utils

import (
	"github.com/golang-collections/collections/set"
)

func IsUniqueStrings(strings []string) bool {
	values := make([]interface{}, len(strings))
	for i := range strings {
		values[i] = strings[i]
	}
	valueSet := set.New(values...)
	return valueSet.Len() == len(strings)
}
