// +build unit

package validation_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/gojek/turing/engines/router"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojek/turing/api/turing/api/request"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service/mocks"
	"github.com/gojek/turing/api/turing/validation"
	"github.com/gojek/turing/engines/experiment/common"
	"github.com/gojek/turing/engines/experiment/manager"
)

func TestValidateLogConfig(t *testing.T) {
	tt := []struct {
		name   string
		input  request.LogConfig
		hasErr bool
	}{
		{
			name: "valid_nop",
			input: request.LogConfig{
				ResultLoggerType: "nop",
			},
			hasErr: false,
		},
		{
			name: "invalid_type",
			input: request.LogConfig{
				ResultLoggerType: "nope",
			},
			hasErr: true,
		},
		{
			name: "valid_bq",
			input: request.LogConfig{
				ResultLoggerType: "bigquery",
				BigQueryConfig: &request.BigQueryConfig{
					Table:                "project.dataset.table",
					ServiceAccountSecret: "acc",
				},
			},
			hasErr: false,
		},
		{
			name: "bq_missing_config",
			input: request.LogConfig{
				ResultLoggerType: "bigquery",
			},
			hasErr: true,
		},
		{
			name: "bq_invalid_table",
			input: request.LogConfig{
				ResultLoggerType: "bigquery",
				BigQueryConfig: &request.BigQueryConfig{
					Table:                "project:dataset.table",
					ServiceAccountSecret: "acc",
				},
			},
			hasErr: true,
		},
		{
			name: "bq_invalid_svc_account",
			input: request.LogConfig{
				ResultLoggerType: "bigquery",
				BigQueryConfig: &request.BigQueryConfig{
					Table:                "project.dataset.table",
					ServiceAccountSecret: "",
				},
			},
			hasErr: true,
		},
		{
			name: "kafka_valid_config",
			input: request.LogConfig{
				ResultLoggerType: "kafka",
				KafkaConfig: &request.KafkaConfig{
					Brokers:             "broker1,broker2",
					Topic:               "topic",
					SerializationFormat: "json",
				},
			},
			hasErr: false,
		},
		{
			name: "kafka_missing_config",
			input: request.LogConfig{
				ResultLoggerType: "kafka",
			},
			hasErr: true,
		},
		{
			name: "kafka_invalid_config_missing_brokers",
			input: request.LogConfig{
				ResultLoggerType: "kafka",
				KafkaConfig: &request.KafkaConfig{
					Topic:               "topic",
					SerializationFormat: "json",
				},
			},
			hasErr: true,
		},
		{
			name: "kafka_invalid_config_missing_topic",
			input: request.LogConfig{
				ResultLoggerType: "kafka",
				KafkaConfig: &request.KafkaConfig{
					Brokers:             "broker1,broker2",
					SerializationFormat: "json",
				},
			},
			hasErr: true,
		},
		{
			name: "kafka_invalid_config_invalid_serialization",
			input: request.LogConfig{
				ResultLoggerType: "kafka",
				KafkaConfig: &request.KafkaConfig{
					Brokers:             "broker1,broker2",
					Topic:               "topic",
					SerializationFormat: "test",
				},
			},
			hasErr: true,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			validate, err := validation.NewValidator(&mocks.ExperimentsService{})
			assert.NoError(t, err)
			err = validate.Struct(&tc.input)
			assert.Equal(t, tc.hasErr, err != nil)
		})
	}
}

func TestValidateExperimentEngineConfig(t *testing.T) {
	// Create mock experiment service
	client1 := manager.Client{ID: "1", Username: "1"}
	client2 := manager.Client{ID: "2", Username: "2"}
	expSvc := &mocks.ExperimentsService{}
	expSvc.On("ValidateExperimentConfig", "xp", manager.TuringExperimentConfig{Client: client1}).
		Return(nil)
	expSvc.On("ValidateExperimentConfig", "xp", manager.TuringExperimentConfig{Client: client2}).
		Return(errors.New("test-error"))

	// Define tests
	tests := []struct {
		name   string
		input  request.ExperimentEngineConfig
		hasErr bool
	}{
		{
			name: "valid_nop",
			input: request.ExperimentEngineConfig{
				Type: "nop",
			},
			hasErr: false,
		},
		{
			name: "valid_exp_config",
			input: request.ExperimentEngineConfig{
				Type: "xp",
				Config: manager.TuringExperimentConfig{
					Client: client1,
				},
			},
			hasErr: false,
		},
		{
			name: "invalid_exp_config",
			input: request.ExperimentEngineConfig{
				Type: "unknown",
				Config: manager.TuringExperimentConfig{
					Client: client2,
				},
			},
			hasErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			validate, err := validation.NewValidator(expSvc)
			assert.NoError(t, err)
			err = validate.Struct(&tc.input)
			assert.Equal(t, tc.hasErr, err != nil)
		})
	}
}

type routerConfigTestCase struct {
	routes         models.Routes
	defaultRouteID string
	trafficRules   models.TrafficRules
	expectedError  string
}

func (tt routerConfigTestCase) RouterConfig() *request.RouterConfig {
	return &request.RouterConfig{
		Routes:         tt.routes,
		DefaultRouteID: tt.defaultRouteID,
		TrafficRules:   tt.trafficRules,
		ExperimentEngine: &request.ExperimentEngineConfig{
			Type: "nop",
		},
		Timeout: "20s",
		LogConfig: &request.LogConfig{
			ResultLoggerType: "nop",
		},
	}
}

func TestValidateTrafficRules(t *testing.T) {
	suite := map[string]routerConfigTestCase{
		"success": {
			routes: models.Routes{
				{
					ID:       "route-a",
					Type:     "PROXY",
					Endpoint: "http://example.com/a",
					Timeout:  "10ms",
				},
				{
					ID:       "route-b",
					Type:     "PROXY",
					Endpoint: "http://example.com/b",
					Timeout:  "10ms",
				},
			},
			defaultRouteID: "route-a",
			trafficRules: models.TrafficRules{
				{
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: common.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a", "region-b"},
						},
					},
					Routes: []string{"route-b"},
				},
			},
		},
		"failure | empty conditions": {
			routes: models.Routes{
				{
					ID:       "route-a",
					Type:     "PROXY",
					Endpoint: "http://example.com/a",
					Timeout:  "10ms",
				},
				{
					ID:       "route-b",
					Type:     "PROXY",
					Endpoint: "http://example.com/b",
					Timeout:  "10ms",
				},
			},
			defaultRouteID: "route-a",
			trafficRules: models.TrafficRules{
				{
					Conditions: []*router.TrafficRuleCondition{},
					Routes:     []string{"route-b"},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules[0].Conditions' " +
				"Error:Field validation for 'Conditions' failed on the 'notBlank' tag",
		},
		"failure | empty routes": {
			routes: models.Routes{
				{
					ID:       "route-a",
					Type:     "PROXY",
					Endpoint: "http://example.com/a",
					Timeout:  "10ms",
				},
			},
			defaultRouteID: "route-a",
			trafficRules: models.TrafficRules{
				{
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: common.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-b"},
						},
					},
					Routes: []string{},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules[0].Routes' " +
				"Error:Field validation for 'Routes' failed on the 'notBlank' tag",
		},
		"failure | unsupported operator": {
			routes: models.Routes{
				{
					ID:       "route-a",
					Type:     "PROXY",
					Endpoint: "http://example.com/a",
					Timeout:  "10ms",
				},
				{
					ID:       "route-b",
					Type:     "PROXY",
					Endpoint: "http://example.com/b",
					Timeout:  "10ms",
				},
			},
			defaultRouteID: "route-a",
			trafficRules: models.TrafficRules{
				{
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: common.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.RuleConditionOperator{},
							Values:      []string{"region-b"},
						},
					},
					Routes: []string{"route-b"},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules[0].Conditions[0].Operator' " +
				"Error:Field validation for 'Operator' failed on the 'required' tag",
		},
		"failure | unsupported field source": {
			routes: models.Routes{
				{
					ID:       "route-a",
					Type:     "PROXY",
					Endpoint: "http://example.com/a",
					Timeout:  "10ms",
				},
				{
					ID:       "route-b",
					Type:     "PROXY",
					Endpoint: "http://example.com/b",
					Timeout:  "10ms",
				},
			},
			defaultRouteID: "route-a",
			trafficRules: models.TrafficRules{
				{
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: "unknown",
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-b"},
						},
					},
					Routes: []string{"route-b"},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules[0].Conditions[0].FieldSource' " +
				"Error:Field validation for 'FieldSource' failed on the 'oneof' tag",
		},
		"failure | incomplete condition": {
			routes: models.Routes{
				{
					ID:       "route-a",
					Type:     "PROXY",
					Endpoint: "http://example.com/a",
					Timeout:  "10ms",
				},
				{
					ID:       "route-b",
					Type:     "PROXY",
					Endpoint: "http://example.com/b",
					Timeout:  "10ms",
				},
			},
			defaultRouteID: "route-a",
			trafficRules: models.TrafficRules{
				{
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: common.HeaderFieldSource,
							Field:       "",
							Operator:    router.InConditionOperator,
							Values:      []string{},
						},
					},
					Routes: []string{"route-b"},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules[0].Conditions[0].Field' " +
				"Error:Field validation for 'Field' failed on the 'required' tag\n" +
				"Key: 'RouterConfig.TrafficRules[0].Conditions[0].Values' " +
				"Error:Field validation for 'Values' failed on the 'notBlank' tag",
		},
		"failure | unknown route": {
			routes: models.Routes{
				{
					ID:       "route-a",
					Type:     "PROXY",
					Endpoint: "http://example.com/a",
					Timeout:  "10ms",
				},
			},
			defaultRouteID: "route-a",
			trafficRules: models.TrafficRules{
				{
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: common.PayloadFieldSource,
							Field:       "some_property",
							Operator:    router.InConditionOperator,
							Values:      []string{"some_value"},
						},
					},
					Routes: []string{"route-c"},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules[0].Routes[0]' " +
				"Error:Field validation for '' failed on the 'oneof' tag",
		},
		"failure | rule contains default route": {
			routes: models.Routes{
				{
					ID:       "route-a",
					Type:     "PROXY",
					Endpoint: "http://example.com/a",
					Timeout:  "10ms",
				},
			},
			defaultRouteID: "route-a",
			trafficRules: models.TrafficRules{
				{
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: common.PayloadFieldSource,
							Field:       "some_property",
							Operator:    router.InConditionOperator,
							Values:      []string{"some_value"},
						},
					},
					Routes: []string{"route-a"},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules[0].Routes[0]' " +
				"Error:Field validation for '' failed on the 'necsfield' tag",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			validate, err := validation.NewValidator(&mocks.ExperimentsService{})
			require.NoError(t, err)

			err = validate.Struct(tt.RouterConfig())
			if tt.expectedError == "" {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, tt.expectedError)
			}
		})
	}
}

type DNSTestStruct struct {
	Name string `validate:"dns,lte=50,gte=3"`
}

func genString(length int) string {
	var sb strings.Builder
	for i := 0; i < length; i++ {
		sb.WriteString("a")
	}

	return sb.String()
}
