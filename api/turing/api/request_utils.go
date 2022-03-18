package api

import (
	"fmt"
	"strconv"

	"github.com/gojek/turing/api/turing/models"
)

// gets an int from the provided request variables. If not found, will throw
// an error.
func getIntFromVars(vars RequestVars, key string) (int, error) {
	var v string
	var ok bool
	if v, ok = vars.get(key); !ok {
		return 0, fmt.Errorf("key %s not found in vars", key)
	}
	return strconv.Atoi(v)
}

// gets an ID from the provided request variables. If not found, will throw
// an error.
func getIDFromVars(vars RequestVars, key string) (models.ID, error) {
	id, err := getIntFromVars(vars, key)
	return models.ID(id), err
}

// gets a deploy flag from the provided request variables. If not found, will throw
// an error.
func getBoolFromVars(vars RequestVars, key string) (bool, error) {
	var v string
	var ok bool
	if v, ok = vars.get(key); !ok {
		return false, fmt.Errorf("key %s not found in vars", key)
	}
	return strconv.ParseBool(v)
}
