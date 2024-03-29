package api

import (
	"testing"

	"gotest.tools/v3/assert"
)

func TestGetIntFromVars(t *testing.T) {
	tt := []struct {
		name     string
		vars     RequestVars
		key      string
		expected int
		hasErr   bool
	}{
		{
			"valid int",
			RequestVars{"project": {"1"}},
			"project",
			1,
			false,
		},
		{
			"invalid value",
			RequestVars{"project": {"a"}},
			"project",
			0,
			true,
		},
		{
			"key not found",
			RequestVars{"project": {"1"}},
			"pro",
			0,
			true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			got, err := getIntFromVars(tc.vars, tc.key)
			if tc.hasErr {
				assert.ErrorContains(t, err, "")
			} else {
				assert.NilError(t, err)
				assert.Equal(t, tc.expected, got)
			}
		})
	}
}
