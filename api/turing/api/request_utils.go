package api

import (
	"fmt"
	"strconv"
)

// gets an int from the provided request variables. If not found, will throw
// an error.
func getIntFromVars(vars map[string]string, key string) (int, error) {
	var v string
	var ok bool
	if v, ok = vars[key]; !ok {
		return 0, fmt.Errorf("key %s not found in vars", key)
	}
	return strconv.Atoi(v)
}
