package models_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/turing/api/turing/models"
)

type defaultTrafficRuleTestCase struct {
	trafficRule   models.DefaultTrafficRule
	serialized    interface{}
	expectedError string
}

func TestDefaultTrafficRules_Value(t *testing.T) {
	suite := map[string]defaultTrafficRuleTestCase{
		"success": {
			trafficRule: models.DefaultTrafficRule{
				Routes: []string{
					"treatment-a",
					"treatment-b",
				},
			},
			serialized: `
				{
					"routes": [
						"treatment-a",
						"treatment-b"
					]
				}
			`,
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			value, err := tt.trafficRule.Value()
			if tt.expectedError == "" {
				byteValue, ok := value.([]byte)
				require.True(t, ok)
				require.NoError(t, err)
				require.JSONEq(t, tt.serialized.(string), string(byteValue))
			} else {
				require.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

func TestDefaultTrafficRules_Scan(t *testing.T) {
	suite := map[string]defaultTrafficRuleTestCase{
		"success": {
			trafficRule: models.DefaultTrafficRule{
				Routes: []string{
					"treatment-a",
					"treatment-b",
				},
			},
			serialized: []byte(`
					{
						"routes": [
							"treatment-a",
							"treatment-b"
						]
					}
				`),
		},
		"failure | invalid type": {
			serialized:    100,
			expectedError: "type assertion to []byte failed",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			var rule models.DefaultTrafficRule
			err := rule.Scan(tt.serialized)
			if tt.expectedError == "" {
				require.NoError(t, err)
				require.Equal(t, tt.trafficRule, rule)
			} else {
				require.EqualError(t, err, tt.expectedError)
			}
		})
	}
}
