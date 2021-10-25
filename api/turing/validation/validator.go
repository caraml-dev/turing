package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/gojek/turing/engines/router"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
	"github.com/gojek/turing/api/turing/api/request"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service"
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
	instance.RegisterStructValidation(newExperimentConfigValidator, request.ExperimentEngineConfig{})
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

func newExperimentConfigValidator(sl validator.StructLevel) {
	field := sl.Current().Interface().(request.ExperimentEngineConfig)
	switch field.Type {
	case string(models.ExperimentEngineTypeNop),
		string(models.ExperimentEngineTypeLitmus),
		string(models.ExperimentEngineTypeXp):
		return
	default:
		sl.ReportError(field.Type, "type", "Type", "oneof", "litmus,xp,nop")
	}
}

func validateRouterConfig(sl validator.StructLevel) {
	routerConfig := sl.Current().Interface().(request.RouterConfig)

	if routerConfig.TrafficRules != nil {
		routeIds := make([]string, len(routerConfig.Routes))
		for idx, route := range routerConfig.Routes {
			routeIds[idx] = route.ID
		}

		routeIdsStr := strings.Join(routeIds, " ")

		for ruleIdx, rule := range routerConfig.TrafficRules {
			if rule.Routes != nil {
				for idx, routeID := range rule.Routes {
					ns := fmt.Sprintf("TrafficRules[%d].Routes[%d]", ruleIdx, idx)
					instance := sl.Validator()
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
