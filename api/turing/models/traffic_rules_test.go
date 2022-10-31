package models_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/engines/experiment/pkg/request"
	"github.com/caraml-dev/turing/engines/router"
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
					Name: "rule-name",
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
					"name": "rule-name",
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
					Name: "rule-name",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: request.HeaderFieldSource,
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
					"name": "rule-name",
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
		"success | unknown field_source": {
			serialized: []byte(`
				[{
					"name": "rule-name",
				 	"conditions": [{
						"field_source": "unknown",
						"field": "X-Region",
						"operator": "in",
						"values": ["region-a", "region-b"]
                    }],
					"routes": []
				}]
			`),
			trafficRules: models.TrafficRules{
				&models.TrafficRule{
					Name: "rule-name",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: "unknown",
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values: []string{
								"region-a", "region-b",
							},
						},
					},
					Routes: []string{},
				},
			},
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
