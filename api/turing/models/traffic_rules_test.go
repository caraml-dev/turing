package models_test

import (
	"testing"

	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/engines/router"
	"github.com/stretchr/testify/require"
)

type trafficRulesTestCase struct {
	trafficRules  models.TrafficRules
	serialized    interface{}
	expectedError string
}

func TestTrafficRules_Value(t *testing.T) {
	suite := map[string]trafficRulesTestCase{
		"success": {
			trafficRules: models.TrafficRules{
				&models.TrafficRule{
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: "header",
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values: []string{
								"region-a", "region-b",
							},
						},
					},
					Routes: []string{
						"treatment-a",
						"treatment-b",
					},
				},
			},
			serialized: `
				[{
				 	"conditions": [{
						"field_source": "header",
						"field": "X-Region",
						"operator": "in",
						"values": ["region-a", "region-b"]
                    }],
					"routes": [
						"treatment-a",
						"treatment-b"
                    ]
				}]
			`,
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			value, err := tt.trafficRules.Value()
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

func TestTrafficRules_Scan(t *testing.T) {
	suite := map[string]trafficRulesTestCase{
		"success": {
			trafficRules: models.TrafficRules{
				&models.TrafficRule{
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: "header",
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values: []string{
								"region-a", "region-b",
							},
						},
					},
					Routes: []string{
						"treatment-a",
						"treatment-b",
					},
				},
			},
			serialized: []byte(`
				[{
				 	"conditions": [{
						"field_source": "header",
						"field": "X-Region",
						"operator": "in",
						"values": ["region-a", "region-b"]
                    }],
					"routes": [
						"treatment-a",
						"treatment-b"
                    ]
				}]
			`),
		},
		"failure | invalid type": {
			serialized:    100,
			expectedError: "type assertion to []byte failed",
		},
		"failure | unknown field_source": {
			serialized: []byte(`
				[{
				 	"conditions": [{
						"field_source": "unknown",
						"field": "X-Region",
						"operator": "in",
						"values": ["region-a", "region-b"]
                    }],
					"routes": []
				}]
			`),
			expectedError: "Unknown field source unknown",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			var rules models.TrafficRules
			err := rules.Scan(tt.serialized)
			if tt.expectedError == "" {
				require.NoError(t, err)
				require.Equal(t, tt.trafficRules, rules)
			} else {
				require.EqualError(t, err, tt.expectedError)
			}
		})
	}
}
