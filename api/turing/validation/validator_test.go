package validation_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/go-playground/validator/v10"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/caraml-dev/turing/api/turing/api/request"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service/mocks"
	"github.com/caraml-dev/turing/api/turing/validation"
	"github.com/caraml-dev/turing/engines/experiment/manager"
	expRequest "github.com/caraml-dev/turing/engines/experiment/pkg/request"
	"github.com/caraml-dev/turing/engines/router"
	routerConfig "github.com/caraml-dev/turing/engines/router/missionctl/config"
)

var dummyConfig json.RawMessage

func getDefaultValidator() (*validator.Validate, error) {
	mockExperimentsService := &mocks.ExperimentsService{}
	mockExperimentsService.On("ListEngines").Return([]manager.Engine{{Name: "custom"}})
	mockExperimentsService.On("ValidateExperimentConfig", "custom", dummyConfig).
		Return(nil)
	return validation.NewValidator(mockExperimentsService)
}
func TestValidateEnsemblerStandardConfig(t *testing.T) {
	tt := map[string]struct {
		input models.EnsemblerStandardConfig
		err   string
	}{
		"failure | experiment mappings and route name path undefined": {
			input: models.EnsemblerStandardConfig{},
			err: "Key: 'EnsemblerStandardConfig.ExperimentMappings' Error:Field validation for 'ExperimentMappings' " +
				"failed on the 'required when RouteNamePath is not set' tag",
		},
		"failure | experiment mappings and route name path empty": {
			input: models.EnsemblerStandardConfig{
				ExperimentMappings: []models.ExperimentMapping{},
				RouteNamePath:      "",
			},
			err: "Key: 'EnsemblerStandardConfig.ExperimentMappings' Error:Field validation for 'ExperimentMappings' " +
				"failed on the 'required when RouteNamePath is not set' tag",
		},
		"failure | experiment mappings empty and route name path undefined": {
			input: models.EnsemblerStandardConfig{
				ExperimentMappings: []models.ExperimentMapping{},
			},
			err: "Key: 'EnsemblerStandardConfig.ExperimentMappings' Error:Field validation for 'ExperimentMappings' " +
				"failed on the 'required when RouteNamePath is not set' tag",
		},
		"failure | experiment mappings undefined and route name path empty": {
			input: models.EnsemblerStandardConfig{
				RouteNamePath: "",
			},
			err: "Key: 'EnsemblerStandardConfig.ExperimentMappings' Error:Field validation for 'ExperimentMappings' " +
				"failed on the 'required when RouteNamePath is not set' tag",
		},
		"failure | experiment mappings and route name path defined": {
			input: models.EnsemblerStandardConfig{
				ExperimentMappings: []models.ExperimentMapping{
					{
						Experiment: "experiment-1",
						Treatment:  "treatment-1",
						Route:      "route-1",
					},
				},
				RouteNamePath: "route-1",
			},
			err: "Key: 'EnsemblerStandardConfig.ExperimentMappings' Error:Field validation for 'ExperimentMappings' " +
				"failed on the 'excluded when RouteNamePath is set' tag",
		},
		"success | only experiment mappings defined": {
			input: models.EnsemblerStandardConfig{
				ExperimentMappings: []models.ExperimentMapping{
					{
						Experiment: "experiment-1",
						Treatment:  "treatment-1",
						Route:      "route-1",
					},
				},
			},
		},
		"success | only route name path defined": {
			input: models.EnsemblerStandardConfig{
				RouteNamePath: "route-1",
			},
		},
	}

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			validate, err := validation.NewValidator(nil)
			assert.NoError(t, err)
			err = validate.Struct(tc.input)
			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

func TestValidateLogConfig(t *testing.T) {
	tt := map[string]struct {
		input  request.LogConfig
		hasErr bool
	}{
		"valid_nop": {
			input: request.LogConfig{
				ResultLoggerType: "nop",
			},
			hasErr: false,
		},
		"invalid_type": {
			input: request.LogConfig{
				ResultLoggerType: "nope",
			},
			hasErr: true,
		},
		"valid_upi": {
			input: request.LogConfig{
				ResultLoggerType: "upi",
			},
			hasErr: false,
		},
		"valid_bq": {
			input: request.LogConfig{
				ResultLoggerType: "bigquery",
				BigQueryConfig: &request.BigQueryConfig{
					Table:                "project.dataset.table",
					ServiceAccountSecret: "acc",
				},
			},
			hasErr: false,
		},
		"bq_missing_config": {
			input: request.LogConfig{
				ResultLoggerType: "bigquery",
			},
			hasErr: true,
		},
		"bq_invalid_table": {
			input: request.LogConfig{
				ResultLoggerType: "bigquery",
				BigQueryConfig: &request.BigQueryConfig{
					Table:                "project:dataset.table",
					ServiceAccountSecret: "acc",
				},
			},
			hasErr: true,
		},
		"bq_invalid_svc_account": {
			input: request.LogConfig{
				ResultLoggerType: "bigquery",
				BigQueryConfig: &request.BigQueryConfig{
					Table:                "project.dataset.table",
					ServiceAccountSecret: "",
				},
			},
			hasErr: true,
		},
		"kafka_valid_config": {
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
		"kafka_missing_config": {
			input: request.LogConfig{
				ResultLoggerType: "kafka",
			},
			hasErr: true,
		},
		"kafka_invalid_config_missing_brokers": {
			input: request.LogConfig{
				ResultLoggerType: "kafka",
				KafkaConfig: &request.KafkaConfig{
					Topic:               "topic",
					SerializationFormat: "json",
				},
			},
			hasErr: true,
		},
		"kafka_invalid_config_missing_topic": {
			input: request.LogConfig{
				ResultLoggerType: "kafka",
				KafkaConfig: &request.KafkaConfig{
					Brokers:             "broker1,broker2",
					SerializationFormat: "json",
				},
			},
			hasErr: true,
		},
		"kafka_invalid_config_invalid_serialization": {
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

	for name, tc := range tt {
		t.Run(name, func(t *testing.T) {
			validate, err := getDefaultValidator()
			assert.NoError(t, err)
			err = validate.Struct(&tc.input)
			assert.Equal(t, tc.hasErr, err != nil)
		})
	}
}

func TestValidateExperimentEngineConfig(t *testing.T) {
	validationErr := "test-error"
	// Create mock experiment service
	config1 := json.RawMessage(`{"client": {"id": "1", "username": "1"}}`)
	config2 := json.RawMessage(`{"client": {"id": "2", "username": "2"}}`)
	expSvc := &mocks.ExperimentsService{}
	expSvc.On("ListEngines").Return([]manager.Engine{{Name: "custom"}})
	expSvc.On("ValidateExperimentConfig", "custom", config1).
		Return(nil)
	expSvc.On("ValidateExperimentConfig", "custom", config2).
		Return(errors.New(validationErr))

	// Define tests
	tests := map[string]struct {
		input request.ExperimentEngineConfig
		err   string
	}{
		"success | valid nop": {
			input: request.ExperimentEngineConfig{
				Type: "nop",
			},
		},
		"success | valid experiment config": {
			input: request.ExperimentEngineConfig{
				Type:   "custom",
				Config: config1,
			},
		},
		"failure | unknown engine type": {
			input: request.ExperimentEngineConfig{
				Type:   "unknown",
				Config: config2,
			},
			err: "Key: 'ExperimentEngineConfig.type' " +
				"Error:Field validation for 'type' failed on the 'oneof' tag",
		},
		"failure | validation error": {
			input: request.ExperimentEngineConfig{
				Type:   "custom",
				Config: config2,
			},
			err: fmt.Sprintf(
				"Key: 'ExperimentEngineConfig.config' "+
					"Error:Field validation for 'config' failed on the '%s' tag", validationErr),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			validate, err := validation.NewValidator(expSvc)
			assert.NoError(t, err)
			err = validate.Struct(&tc.input)
			if tc.err == "" {
				assert.NoError(t, err)
			} else {
				assert.EqualError(t, err, tc.err)
			}
		})
	}
}

type routerConfigTestCase struct {
	protocol           routerConfig.Protocol
	routes             models.Routes
	enricher           *request.EnricherEnsemblerConfig
	ensembler          *models.Ensembler
	experimentEngine   *request.ExperimentEngineConfig
	defaultRouteID     *string
	defaultTrafficRule *models.DefaultTrafficRule
	trafficRules       models.TrafficRules
	autoscalingPolicy  *models.AutoscalingPolicy
	expectedError      string
	logConfig          *request.LogConfig
}

func (tt routerConfigTestCase) RouterConfig() *request.RouterConfig {
	experimentEngine := &request.ExperimentEngineConfig{
		Type: "nop",
	}
	if tt.experimentEngine != nil {
		experimentEngine = tt.experimentEngine
	}
	if tt.logConfig == nil {
		tt.logConfig = &request.LogConfig{ResultLoggerType: "nop"}
	}
	return &request.RouterConfig{
		Routes:             tt.routes,
		DefaultRouteID:     tt.defaultRouteID,
		DefaultTrafficRule: tt.defaultTrafficRule,
		TrafficRules:       tt.trafficRules,
		AutoscalingPolicy:  tt.autoscalingPolicy,
		ExperimentEngine:   experimentEngine,
		Timeout:            "20s",
		Protocol:           &tt.protocol,
		LogConfig:          tt.logConfig,
		Enricher:           tt.enricher,
		Ensembler:          tt.ensembler,
	}
}

func TestValidateTrafficRules(t *testing.T) {
	// Common variables used by suite tests
	ruleName := "rule-name"
	routeAID, routeBID, routeCID := "route-a", "route-b", "route-c"
	routeA, routeB, routeC := &models.Route{
		ID:       routeAID,
		Type:     "PROXY",
		Endpoint: "http://example.com/a",
		Timeout:  "10ms",
	}, &models.Route{
		ID:       routeBID,
		Type:     "PROXY",
		Endpoint: "http://example.com/b",
		Timeout:  "10ms",
	}, &models.Route{
		ID:       routeCID,
		Type:     "PROXY",
		Endpoint: "http://example.com/c",
		Timeout:  "10ms",
	}
	defaultTrafficRule := &models.DefaultTrafficRule{
		Routes: []string{routeAID, routeBID},
	}

	suite := map[string]routerConfigTestCase{
		"success": {
			routes:             models.Routes{routeA, routeB},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: ruleName,
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a", "region-b"},
						},
					},
					Routes: []string{"route-a", "route-b"},
				},
			},
		},
		"success | valid trailing symbol": {
			routes:             models.Routes{routeA, routeB},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: "aBc -_()#$%&:.",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a", "region-b"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
			},
		},
		"success | valid trailing alphabet ": {
			routes:             models.Routes{routeA, routeB},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: "aBc -_()#$%&:.d",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a", "region-b"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
			},
		},
		"success | valid complex Traffic Rules": {
			routes:             models.Routes{routeA, routeB, routeC},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: "rule-a",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Country",
							Operator:    router.InConditionOperator,
							Values:      []string{"ID"},
						},
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
				{
					Name: "rule-b",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Country",
							Operator:    router.InConditionOperator,
							Values:      []string{"SG"},
						},
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-b"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
				{
					Name: "rule-c",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Country",
							Operator:    router.InConditionOperator,
							Values:      []string{"VN"},
						},
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a"},
						},
					},
					Routes: []string{routeAID, routeCID},
				},
				{
					Name: "rule-d",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Country",
							Operator:    router.InConditionOperator,
							Values:      []string{"SG"},
						},
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-c"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
			},
		},
		"failure | empty conditions": {
			routes:             models.Routes{routeA, routeB},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name:       ruleName,
					Conditions: []*router.TrafficRuleCondition{},
					Routes:     []string{routeAID},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules[0].Conditions' " +
				"Error:Field validation for 'Conditions' failed on the 'notBlank' tag",
		},
		"failure | empty routes": {
			routes:             models.Routes{routeA, routeB},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: ruleName,
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-b"},
						},
					},
					Routes: []string{},
				},
			},
			expectedError: strings.Join([]string{
				"Key: 'RouterConfig.TrafficRules[0].Routes' Error:Field validation for 'Routes' failed on the 'notBlank' tag",
				"Key: 'RouterConfig.TrafficRules' Error:Field validation for 'TrafficRules' failed on the " +
					"'Fallback Route (DefaultRouteId): 'route-a' should be associated to all Traffic Rules' tag",
			}, "\n"),
		},
		"failure | unsupported operator": {
			routes:             models.Routes{routeA, routeB},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: ruleName,
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.RuleConditionOperator{},
							Values:      []string{"region-b"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules[0].Conditions[0].Operator' " +
				"Error:Field validation for 'Operator' failed on the 'required' tag",
		},
		"failure : unsupported field source for UPI": {
			protocol:           routerConfig.UPI,
			routes:             models.Routes{routeA, routeB},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: ruleName,
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.PayloadFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-b"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules[0].Conditions[0].FieldSource' Error:Field validation for " +
				"'TrafficRules[0].Conditions[0].FieldSource' failed on the 'oneof' tag",
		},
		"failure : unsupported field source for HTTP": {
			protocol:           routerConfig.HTTP,
			routes:             models.Routes{routeA, routeB},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: ruleName,
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.PredictionContextSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-b"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules[0].Conditions[0].FieldSource' Error:Field validation for " +
				"'TrafficRules[0].Conditions[0].FieldSource' failed on the 'oneof' tag",
		},
		"failure | incomplete condition": {
			routes:             models.Routes{routeA, routeB},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: ruleName,
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "",
							Operator:    router.InConditionOperator,
							Values:      []string{},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules[0].Conditions[0].Field' " +
				"Error:Field validation for 'Field' failed on the 'required' tag\n" +
				"Key: 'RouterConfig.TrafficRules[0].Conditions[0].Values' " +
				"Error:Field validation for 'Values' failed on the 'notBlank' tag",
		},
		"failure | nop ensembler missing default route": {
			routes: models.Routes{routeA},
			expectedError: "Key: 'RouterConfig.default_route_id' " +
				"Error:Field validation for 'default_route_id' failed on the 'should be set for chosen ensembler type' tag",
		},
		"failure | standard ensembler missing default route": {
			routes: models.Routes{routeA},
			experimentEngine: &request.ExperimentEngineConfig{
				Type: "custom",
			},
			ensembler: &models.Ensembler{Type: models.EnsemblerStandardType},
			expectedError: "Key: 'RouterConfig.default_route_id' " +
				"Error:Field validation for 'default_route_id' failed on the 'should be set for chosen ensembler type' tag",
		},
		"failure | docker ensembler has default route": {
			routes:         models.Routes{routeA},
			ensembler:      &models.Ensembler{Type: models.EnsemblerDockerType},
			defaultRouteID: &routeAID,
			expectedError: "Key: 'RouterConfig.default_route_id' " +
				"Error:Field validation for 'default_route_id' failed on the 'should not be set for chosen ensembler type' tag",
		},
		"failure | pyfunc ensembler has default route": {
			routes:         models.Routes{routeA},
			ensembler:      &models.Ensembler{Type: models.EnsemblerPyFuncType},
			defaultRouteID: &routeAID,
			expectedError: "Key: 'RouterConfig.default_route_id' " +
				"Error:Field validation for 'default_route_id' failed on the 'should not be set for chosen ensembler type' tag",
		},
		"failure | unknown default route": {
			routes:         models.Routes{routeA},
			defaultRouteID: &routeBID,
			expectedError:  "Key: 'RouterConfig.DefaultRouteID' Error:Field validation for '' failed on the 'oneof' tag",
		},
		"failure | unknown traffic rule route": {
			routes:             models.Routes{routeA, routeB},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: ruleName,
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.PayloadFieldSource,
							Field:       "some_property",
							Operator:    router.InConditionOperator,
							Values:      []string{"some_value"},
						},
					},
					Routes: []string{routeAID, routeCID},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules[0].Routes[1]' " +
				"Error:Field validation for '' failed on the 'oneof' tag",
		},
		"failure | rule does not contains default route": {
			routes:         models.Routes{routeA},
			defaultRouteID: &routeAID,
			trafficRules: models.TrafficRules{
				{
					Name: ruleName,
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.PayloadFieldSource,
							Field:       "some_property",
							Operator:    router.InConditionOperator,
							Values:      []string{"some_value"},
						},
					},
					Routes: []string{"route-a"},
				},
			},
			expectedError: strings.Join([]string{
				"Key: 'RouterConfig.DefaultTrafficRule' Error:Field validation for " +
					"'DefaultTrafficRule' failed on the 'Since 1 or more Custom Traffic rules have been specified, " +
					"a default Traffic rule is required.' tag",
			}, "\n"),
		},
		"failure | empty name": {
			routes:             models.Routes{routeA, routeB},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a", "region-b"},
						},
					},
					Routes: []string{routeBID},
				},
			},
			expectedError: strings.Join([]string{
				"Key: 'RouterConfig.TrafficRules[0].Name' Error:Field validation for 'Name' failed on the 'required' tag",
				"Key: 'RouterConfig.TrafficRules' Error:Field validation for 'TrafficRules' failed on the " +
					"'Fallback Route (DefaultRouteId): 'route-a' should be associated to all Traffic Rules' tag",
				strings.Join([]string{
					"Key: 'RouterConfig.TrafficRule' Error:Field validation for 'TrafficRule' failed on the",
					"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
					"and have no trailing spaces and can contain letters, numbers, blank spaces and the",
					"following symbols: -_()#$&:.' tag",
				}, " "),
			}, "\n"),
		},
		"failure | More than 64 characters": {
			routes:             models.Routes{routeA, routeB},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: "abcdefghijklmnopqrstuvwxyz abcdefghijklmnopqrstuvwxyz abcdefghijklmnopqrstuvwxyz",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a", "region-b"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
			},
			expectedError: strings.Join([]string{
				"Key: 'RouterConfig.TrafficRule' Error:Field validation for 'TrafficRule' failed on the",
				"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
				"and have no trailing spaces and can contain letters, numbers, blank spaces and the",
				"following symbols: -_()#$&:.' tag",
			}, " "),
		},
		"failure | Invalid trailing character": {
			routes:             models.Routes{routeA, routeB},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: "abc@",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a", "region-b"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
			},
			expectedError: strings.Join([]string{
				"Key: 'RouterConfig.TrafficRule' Error:Field validation for 'TrafficRule' failed on the",
				"'Name must be between 4-64 characters long, and begin with an alphanumeric character",
				"and have no trailing spaces and can contain letters, numbers, blank spaces and the",
				"following symbols: -_()#$&:.' tag",
			}, " "),
		},
		"failure | Non-unique Traffic Rule names": {
			routes:             models.Routes{routeA, routeB, routeC},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: "abcd",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
				{
					Name: "abcd",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-c"},
						},
					},
					Routes: []string{routeAID, routeCID},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules' " +
				"Error:Field validation for 'TrafficRules' failed on the 'unique' tag",
		},
		"failure | Invalid name default-traffic-rule": {
			routes:             models.Routes{routeA, routeB},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: "default-traffic-rule",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a", "region-b"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
			},
			expectedError: strings.Join([]string{
				"Key: 'RouterConfig.TrafficRule' Error:Field validation for 'TrafficRule' failed on the",
				"'default-traffic-rule is a reserved name, and cannot be used as the name for a Custom Traffic Rule.' tag",
			}, " "),
		},
		"failure | missing default route id in default traffic rule": {
			routes:         models.Routes{routeA, routeB},
			defaultRouteID: &routeAID,
			defaultTrafficRule: &models.DefaultTrafficRule{
				Routes: []string{routeBID},
			},
			trafficRules: models.TrafficRules{
				{
					Name: ruleName,
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a", "region-b"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
			},
			expectedError: strings.Join([]string{
				"Key: 'RouterConfig.TrafficRules' Error:Field validation for 'TrafficRules' failed on the",
				"'Fallback Route (DefaultRouteId): 'route-a' should be associated to all Traffic Rules' tag",
			}, " "),
		},
		"failure | Overlapping Traffic Rules": {
			routes:             models.Routes{routeA, routeB, routeC},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: "rule-a",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
				{
					Name: "rule-c",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a"},
						},
					},
					Routes: []string{routeAID, routeCID},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules' Error:Field validation for 'TrafficRules' " +
				"failed on the 'Rules Orthogonality check failed, following pairs of rules are overlapping - " +
				"(rule-a,rule-c).' tag",
		},
		"failure | Complex overlapping Traffic Rules": {
			routes:             models.Routes{routeA, routeB, routeC},
			defaultRouteID:     &routeAID,
			defaultTrafficRule: defaultTrafficRule,
			trafficRules: models.TrafficRules{
				{
					Name: "rule-a",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Country",
							Operator:    router.InConditionOperator,
							Values:      []string{"ID"},
						},
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
				{
					Name: "rule-b",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Country",
							Operator:    router.InConditionOperator,
							Values:      []string{"ID", "SG"},
						},
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a", "region-b", "region-c"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
				{
					Name: "rule-c",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Country",
							Operator:    router.InConditionOperator,
							Values:      []string{"SG"},
						},
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-a"},
						},
					},
					Routes: []string{routeAID, routeCID},
				},
				{
					Name: "rule-d",
					Conditions: []*router.TrafficRuleCondition{
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Country",
							Operator:    router.InConditionOperator,
							Values:      []string{"ID"},
						},
						{
							FieldSource: expRequest.HeaderFieldSource,
							Field:       "X-Region",
							Operator:    router.InConditionOperator,
							Values:      []string{"region-b"},
						},
					},
					Routes: []string{routeAID, routeBID},
				},
			},
			expectedError: "Key: 'RouterConfig.TrafficRules' Error:Field validation for 'TrafficRules' " +
				"failed on the 'Rules Orthogonality check failed, following pairs of rules are overlapping - " +
				"(rule-a,rule-b), (rule-b,rule-c), (rule-b,rule-d).' tag",
		},
	}

	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			validate, err := getDefaultValidator()
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

func TestValidateAutoscaling(t *testing.T) {
	enricherBasic := request.EnricherEnsemblerConfig{
		Image: "lala",
		ResourceRequest: &models.ResourceRequest{
			MinReplica: 0,
			MaxReplica: 5,
			CPURequest: resource.Quantity{
				Format: "500M",
			},
			MemoryRequest: resource.Quantity{
				Format: "1G",
			},
		},
		Endpoint: "endpoint",
		Timeout:  "6s",
		Port:     8080,
		Env: []*models.EnvVar{
			{
				Name:  "key",
				Value: "value",
			},
		},
	}
	makeEnricher := func(
		enricher request.EnricherEnsemblerConfig,
		policy *models.AutoscalingPolicy,
	) *request.EnricherEnsemblerConfig {
		newEnr := enricher
		newEnr.AutoscalingPolicy = policy
		return &newEnr
	}
	routeID := "abc"
	route := &models.Route{
		ID:       routeID,
		Type:     "PROXY",
		Endpoint: "http://example.com/a",
		Timeout:  "10ms",
	}
	id := models.ID(1)
	suite := map[string]routerConfigTestCase{
		"success | no autoscaling policy | all components": {
			routes:   models.Routes{route},
			enricher: &enricherBasic,
			ensembler: &models.Ensembler{
				Type: models.EnsemblerDockerType,
			},
		},
		"success | no autoscaling policy | pyfunc ensembler": {
			routes: models.Routes{route},
			ensembler: &models.Ensembler{
				Type: models.EnsemblerPyFuncType,
				PyfuncConfig: &models.EnsemblerPyfuncConfig{
					ProjectID:       &id,
					EnsemblerID:     &id,
					ResourceRequest: &models.ResourceRequest{},
					Timeout:         "1s",
					Env:             models.EnvVars{},
				},
			},
		},
		"success | valid autoscaling policy | all components": {
			routes: models.Routes{route},
			autoscalingPolicy: &models.AutoscalingPolicy{
				Metric: models.AutoscalingMetricCPU,
				Target: "90",
			},
			enricher: makeEnricher(enricherBasic, &models.AutoscalingPolicy{
				Metric: models.AutoscalingMetricConcurrency,
				Target: "2",
			}),
			ensembler: &models.Ensembler{
				Type: models.EnsemblerDockerType,
				DockerConfig: &models.EnsemblerDockerConfig{
					Endpoint: "http://abc.com",
					Port:     8080,
					Image:    "nginx",
					ResourceRequest: &models.ResourceRequest{
						CPURequest:    resource.Quantity{Format: "500m"},
						MemoryRequest: resource.Quantity{Format: "1Gi"},
					},
					AutoscalingPolicy: &models.AutoscalingPolicy{
						Metric: models.AutoscalingMetricRPS,
						Target: "400",
					},
					Timeout: "5s",
					Env:     models.EnvVars{},
				},
			},
		},
		"success | valid autoscaling policy | pyfunc ensembler": {
			routes: models.Routes{route},
			ensembler: &models.Ensembler{
				Type: models.EnsemblerPyFuncType,
				PyfuncConfig: &models.EnsemblerPyfuncConfig{
					ProjectID:       &id,
					EnsemblerID:     &id,
					ResourceRequest: &models.ResourceRequest{},
					AutoscalingPolicy: &models.AutoscalingPolicy{
						Metric: models.AutoscalingMetricMemory,
						Target: "80",
					},
					Timeout: "1s",
					Env:     models.EnvVars{},
				},
			},
		},
		"failure | invalid autoscaling metric": {
			routes:         models.Routes{route},
			defaultRouteID: &routeID,
			autoscalingPolicy: &models.AutoscalingPolicy{
				Metric: "abc",
				Target: "100",
			},
			expectedError: strings.Join([]string{"Key: 'RouterConfig.AutoscalingPolicy.Metric' ",
				"Error:Field validation for 'Metric' failed on the 'oneof' tag"}, ""),
		},
		"failure | invalid autoscaling target": {
			routes:         models.Routes{route},
			defaultRouteID: &routeID,
			autoscalingPolicy: &models.AutoscalingPolicy{
				Metric: models.AutoscalingMetricRPS,
				Target: "hundred",
			},
			expectedError: strings.Join([]string{"Key: 'RouterConfig.AutoscalingPolicy.Target' ",
				"Error:Field validation for 'Target' failed on the 'number' tag"}, ""),
		},
	}
	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			validate, err := getDefaultValidator()
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

func TestValidateStdEnsemblerNotConfiguredForNopExpEngine(t *testing.T) {
	routeID := "abc"
	route := &models.Route{
		ID:       routeID,
		Type:     "PROXY",
		Endpoint: "http://example.com/a",
		Timeout:  "10ms",
	}

	suite := map[string]routerConfigTestCase{
		"success": {
			routes:         models.Routes{route},
			defaultRouteID: &routeID,
			experimentEngine: &request.ExperimentEngineConfig{
				Type: "custom",
			},
			ensembler: &models.Ensembler{
				Type: models.EnsemblerStandardType,
				StandardConfig: &models.EnsemblerStandardConfig{
					RouteNamePath: "route-1",
				},
			},
		},
		"failure | standard ensembler is configured with nop exp engine": {
			routes:         models.Routes{route},
			defaultRouteID: &routeID,
			ensembler: &models.Ensembler{
				Type: models.EnsemblerStandardType,
				StandardConfig: &models.EnsemblerStandardConfig{
					RouteNamePath: "route-1",
				},
			},
			expectedError: strings.Join([]string{"Key: 'RouterConfig.Ensembler.Type' ",
				"Error:Field validation for 'Ensembler.Type' failed on the ",
				"'should not be standard if an nop experiment engine is configured' tag"}, ""),
		},
	}
	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			validate, err := getDefaultValidator()
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

func TestValidateUPIRouter(t *testing.T) {
	routeID := "abc"
	route := &models.Route{ID: routeID}

	suite := map[string]routerConfigTestCase{
		"success": {
			routes:         models.Routes{route},
			defaultRouteID: &routeID,
			protocol:       routerConfig.UPI,
			experimentEngine: &request.ExperimentEngineConfig{
				Type: "custom",
			},
			ensembler: &models.Ensembler{
				Type: models.EnsemblerStandardType,
			},
			logConfig: &request.LogConfig{
				ResultLoggerType: models.UPILogger,
			},
		},
		"failure | unsupported ensembler type": {
			routes:   models.Routes{route},
			protocol: routerConfig.UPI,
			ensembler: &models.Ensembler{
				Type: models.EnsemblerDockerType,
			},
			expectedError: "Key: 'RouterConfig.Ensembler.Type' Error:Field validation for 'Ensembler.Type' " +
				"failed on the 'only standard ensembler is supported for UPI' tag",
		},
	}
	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			validate, err := getDefaultValidator()
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

func TestValidateHTTPRouter(t *testing.T) {
	routeID := "abc"
	route := &models.Route{ID: routeID}

	suite := map[string]routerConfigTestCase{
		"success": {
			routes:         models.Routes{route},
			defaultRouteID: &routeID,
			logConfig: &request.LogConfig{
				ResultLoggerType: models.NopLogger,
			},
		},
		"failure | unsupported logger type": {
			routes:         models.Routes{route},
			defaultRouteID: &routeID,
			logConfig: &request.LogConfig{
				ResultLoggerType: models.UPILogger},
			expectedError: "Key: 'RouterConfig.LogConfig.ResultLoggerType' Error:Field validation for " +
				"'LogConfig.ResultLoggerType' failed on the 'logger should not be upi' tag",
		},
	}
	for name, tt := range suite {
		t.Run(name, func(t *testing.T) {
			validate, err := getDefaultValidator()
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
