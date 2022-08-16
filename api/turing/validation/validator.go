package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/caraml-dev/turing/engines/router"

	"github.com/caraml-dev/turing/api/turing/api/request"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
)

var tableRegexString string = `.+\.[a-zA-Z0-9_]+\.[a-zA-Z0-9_]+`

// NewValidator creates a new validator using the given defaults
func NewValidator(expSvc service.ExperimentsService) (*validator.Validate, error) {
	instance := validator.New()
	if err := instance.RegisterValidation("notBlank", validators.NotBlank); err != nil {
		return nil, err
	}
	// Register validators
	instance.RegisterStructValidation(validateLogConfig, request.LogConfig{})
	if expSvc != nil {
		instance.RegisterStructValidation(
			newExperimentConfigValidator(expSvc),
			request.ExperimentEngineConfig{},
		)
	}
	instance.RegisterStructValidation(validateRouterConfig, request.RouterConfig{})

	// register common.RuleConditionOperator type to use its String representation for validation
	instance.RegisterCustomTypeFunc(func(field reflect.Value) interface{} {
		if val, ok := field.Interface().(router.RuleConditionOperator); ok {
			return val.String()
		}
		return nil
	}, router.RuleConditionOperator{})

	return instance, nil
}

func validateLogConfig(sl validator.StructLevel) {
	field := sl.Current().Interface().(request.LogConfig)
	switch field.ResultLoggerType {
	case models.NopLogger:
		return
	case models.BigQueryLogger:
		bqConf := field.BigQueryConfig
		if bqConf == nil {
			sl.ReportError(field.BigQueryConfig,
				"bigquery_config", "BigQueryConfig", "bigquery-config-missing", "")
		} else {
			tableRegex := regexp.MustCompile(tableRegexString)
			if !tableRegex.MatchString(bqConf.Table) {
				sl.ReportError(field.BigQueryConfig,
					"bigquery_config", "BigQueryConfig", "bigquery-config-invalid-tablename", "")
			}
			if len(bqConf.ServiceAccountSecret) == 0 {
				sl.ReportError(field.BigQueryConfig,
					"bigquery_config", "BigQueryConfig", "bigquery-config-missing-svc-account", "")
			}
		}
		return
	case models.KafkaLogger:
		kafkaConf := field.KafkaConfig
		if kafkaConf == nil {
			sl.ReportError(field.KafkaConfig,
				"kafka_config", "KafkaConfig", "kafka-config-missing", "")
		} else {
			if len(kafkaConf.Brokers) == 0 {
				sl.ReportError(field.KafkaConfig,
					"kafka_config", "KafkaConfig", "kafka-config-brokers-missing", "")
			}
			if len(kafkaConf.Topic) == 0 {
				sl.ReportError(field.KafkaConfig,
					"kafka_config", "KafkaConfig", "kafka-config-topic-missing", "")
			}
			if kafkaConf.SerializationFormat != models.JSONSerializationFormat &&
				kafkaConf.SerializationFormat != models.ProtobufSerializationFormat {
				sl.ReportError(field.KafkaConfig,
					"kafka_config", "KafkaConfig", "kafka-serialization-format-oneOf",
					string(kafkaConf.SerializationFormat))
			}
		}
		return
	default:
		sl.ReportError(field.ResultLoggerType, "type", "ResultLoggerType", "oneof", "bigquery,nop")
	}
}

func newExperimentConfigValidator(expSvc service.ExperimentsService) func(validator.StructLevel) {
	supportedEngines := make(map[string]bool)
	supportedEnginesStr := models.ExperimentEngineTypeNop
	for _, engine := range expSvc.ListEngines() {
		supportedEnginesStr = fmt.Sprintf("%s,%s", supportedEnginesStr, engine.Name)
		supportedEngines[engine.Name] = true
	}

	validationFunc := func(sl validator.StructLevel) {
		field := sl.Current().Interface().(request.ExperimentEngineConfig)
		switch {
		case field.Type == models.ExperimentEngineTypeNop:
			return
		case supportedEngines[field.Type]:
			err := expSvc.ValidateExperimentConfig(field.Type, field.Config)
			if err != nil {
				sl.ReportError(field.Config, "config", "ExperimentEngineConfig.Config", err.Error(), "")
			}
		default:
			sl.ReportError(field.Type, "type", "Type", "oneof", supportedEnginesStr)
		}
	}
	return validationFunc
}

func validateRouterConfig(sl validator.StructLevel) {
	routerConfig := sl.Current().Interface().(request.RouterConfig)
	instance := sl.Validator()

	routeIds := make([]string, len(routerConfig.Routes))
	for idx, route := range routerConfig.Routes {
		routeIds[idx] = route.ID
	}
	routeIdsStr := strings.Join(routeIds, " ")

	// Validate default route
	if routerConfig.Ensembler == nil || routerConfig.Ensembler.Type == models.EnsemblerStandardType {
		if routerConfig.DefaultRouteID == nil {
			sl.ReportError(routerConfig.DefaultRouteID, "default_route_id", "DefaultRouteID",
				"should be set for chosen ensembler type", "")
		} else {
			if err := instance.Var(*routerConfig.DefaultRouteID, fmt.Sprintf("oneof=%s", routeIdsStr)); err != nil {
				ns := "DefaultRouteID"
				sl.ReportValidationErrors(ns, ns, err.(validator.ValidationErrors))
			}
		}
	} else if routerConfig.DefaultRouteID != nil && *routerConfig.DefaultRouteID != "" {
		sl.ReportError(routerConfig.DefaultRouteID, "default_route_id", "DefaultRouteID",
			"should not be set for chosen ensembler type", *routerConfig.DefaultRouteID)
	}

	// Validate traffic rules
	if routerConfig.TrafficRules != nil {
		for ruleIdx, rule := range routerConfig.TrafficRules {
			if rule.Routes != nil {
				for idx, routeID := range rule.Routes {
					ns := fmt.Sprintf("TrafficRules[%d].Routes[%d]", ruleIdx, idx)
					if err := instance.Var(routeID, fmt.Sprintf("oneof=%s", routeIdsStr)); err != nil {
						sl.ReportValidationErrors(ns, ns, err.(validator.ValidationErrors))
					}

					if err := instance.VarWithValue(routeID, routerConfig.DefaultRouteID, "necsfield"); err != nil {
						sl.ReportValidationErrors(ns, ns, err.(validator.ValidationErrors))
					}
				}
			}
		}
	}
}
