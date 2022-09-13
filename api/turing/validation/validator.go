package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/golang-collections/collections/set"

	"github.com/caraml-dev/turing/api/turing/api/request"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/engines/router"
	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"
)

var tableRegexString string = `.+\.[a-zA-Z0-9_]+\.[a-zA-Z0-9_]+`
var trafficRuleNameRegex = regexp.MustCompile(`^[A-Za-z\d][\w\d \-()#$%&:.]{2,62}[\w\d\-()#$%&:.]$`)

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

func checkTrafficRuleName(sl validator.StructLevel, fieldName string, value string) {
	nameRegexDescription := strings.Join([]string{
		"Name must be between 4-64 characters long, and begin with an alphanumeric character",
		"and have no trailing spaces and can contain letters, numbers, blank spaces and the following symbols: -_()#$&:.",
	}, " ")
	if !trafficRuleNameRegex.MatchString(value) {
		sl.ReportError(value, fieldName, "trafficRuleName", nameRegexDescription, fmt.Sprintf("%v", value))
	}
	invalidNameDescription :=
		"default-traffic-rule is a reserved name, and cannot be used as the name for a Custom Traffic Rule."
	if value == "default-traffic-rule" {
		sl.ReportError(value, fieldName, "trafficRuleName", invalidNameDescription, fmt.Sprintf("%v", value))
	}
}

func checkDefaultTrafficRule(
	sl validator.StructLevel,
	fieldName string,
	defaultTrafficRule *models.DefaultTrafficRule,
	defaultRouteID *string,
) {
	defaultTrafficRuleDescription := strings.Join([]string{
		"Since 1 or more Custom Traffic rules have been specified,",
		"a default Traffic rule is required.",
	}, " ")
	if defaultTrafficRule == nil {
		sl.ReportError(defaultTrafficRule, fieldName, "defaultTrafficRule", defaultTrafficRuleDescription, "")
	}

	// DefaultRouteId should be present in Default Traffic Rule
	if defaultTrafficRule != nil && defaultRouteID != nil {
		missingDefaultRouteDescription := fmt.Sprintf(
			"Fallback Route (DefaultRouteId): '%s' should be associated for the Default Traffic Rule", *defaultRouteID)
		containsDefaultRouteID := false
		for _, route := range defaultTrafficRule.Routes {
			if route == *defaultRouteID {
				containsDefaultRouteID = true
			}
		}
		if !containsDefaultRouteID {
			sl.ReportError(defaultTrafficRule, fieldName, "defaultTrafficRule", missingDefaultRouteDescription, "")
		}
	}
}

func checkDanglingRoutes(
	sl validator.StructLevel,
	fieldName string,
	allRoutes models.Routes,
	allRulesRoutes *set.Set,
) {
	danglingRoutesDescription :=
		"These route(s) should be removed since they have no rule associated and will never be called: %s"
	danglingRoutes := make([]string, 0)
	for _, route := range allRoutes {
		if !allRulesRoutes.Has(route.ID) {
			danglingRoutes = append(danglingRoutes, route.ID)
		}
	}

	if len(danglingRoutes) > 0 {
		sl.ReportError(
			"", fieldName, "danglingRoutes", fmt.Sprintf(danglingRoutesDescription, strings.Join(danglingRoutes, ",")), "",
		)
	}
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
	allRuleRoutesSet := set.New()
	if routerConfig.TrafficRules != nil {
		if len(routerConfig.TrafficRules) > 0 {
			checkDefaultTrafficRule(sl, "DefaultTrafficRule", routerConfig.DefaultTrafficRule, routerConfig.DefaultRouteID)
			if routerConfig.DefaultTrafficRule != nil {
				for _, route := range routerConfig.DefaultTrafficRule.Routes {
					allRuleRoutesSet.Insert(route)
				}
			}
		}
		rulesContainDefaultRouteID := true
		for ruleIdx, rule := range routerConfig.TrafficRules {
			checkTrafficRuleName(sl, "TrafficRule", rule.Name)
			containsDefaultRouteID := false
			// Consider success if DefaultRouteID is not provided
			if routerConfig.DefaultRouteID == nil {
				containsDefaultRouteID = true
			}
			if rule.Routes != nil {
				for idx, routeID := range rule.Routes {
					allRuleRoutesSet.Insert(routeID)
					ns := fmt.Sprintf("TrafficRules[%d].Routes[%d]", ruleIdx, idx)
					if err := instance.Var(routeID, fmt.Sprintf("oneof=%s", routeIdsStr)); err != nil {
						sl.ReportValidationErrors(ns, ns, err.(validator.ValidationErrors))
					}

					// Check if DefaultRouteID is provided
					if routerConfig.DefaultRouteID != nil && routeID == *routerConfig.DefaultRouteID {
						containsDefaultRouteID = true
					}
				}
			}
			if !containsDefaultRouteID {
				rulesContainDefaultRouteID = false
			}
		}
		// If Traffic Rules are configured, DefaultRouteId should be present in all rules
		if !rulesContainDefaultRouteID {
			missingDefaultRouteDescription := fmt.Sprintf(
				"Fallback Route (DefaultRouteId): '%s' should be associated for all Traffic Rules", *routerConfig.DefaultRouteID,
			)
			sl.ReportError(
				"TrafficRules", "TrafficRules", "DefaultRouteID", missingDefaultRouteDescription, "",
			)
		}
	}

	// Validate dangling routes
	if routerConfig.TrafficRules != nil && len(routerConfig.TrafficRules) > 0 {
		checkDanglingRoutes(sl, "Routes", routerConfig.Routes, allRuleRoutesSet)
	}
}
