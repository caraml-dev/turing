package validation

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"github.com/golang-collections/collections/set"

	"github.com/go-playground/validator/v10"
	"github.com/go-playground/validator/v10/non-standard/validators"

	"github.com/caraml-dev/turing/api/turing/api/request"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
	expRequest "github.com/caraml-dev/turing/engines/experiment/pkg/request"
	"github.com/caraml-dev/turing/engines/router"
	routerConfig "github.com/caraml-dev/turing/engines/router/missionctl/config"
)

var tableRegexString = `.+\.[a-zA-Z0-9_]+\.[a-zA-Z0-9_]+`
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

	instance.RegisterStructValidation(validateEnsemblerStandardConfig, models.EnsemblerStandardConfig{})

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

func validateEnsemblerStandardConfig(sl validator.StructLevel) {
	ensemblerStandardConfig := sl.Current().Interface().(models.EnsemblerStandardConfig)
	// Verify that the ExperimentMappings and RouteNamePath are not both empty at the same time
	if (len(ensemblerStandardConfig.ExperimentMappings) == 0) && ensemblerStandardConfig.RouteNamePath == "" {
		sl.ReportError(ensemblerStandardConfig.ExperimentMappings,
			"ExperimentMappings", "ExperimentMappings", "required when RouteNamePath is not set", "")
	}
	// Verify that the ExperimentMappings and RouteNamePath are not both set at the same time
	if len(ensemblerStandardConfig.ExperimentMappings) > 0 && ensemblerStandardConfig.RouteNamePath != "" {
		sl.ReportError(ensemblerStandardConfig.ExperimentMappings,
			"ExperimentMappings", "ExperimentMappings", "excluded when RouteNamePath is set", "")
	}
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
) {
	defaultTrafficRuleDescription := strings.Join([]string{
		"Since 1 or more Custom Traffic rules have been specified,",
		"a default Traffic rule is required.",
	}, " ")
	if defaultTrafficRule == nil {
		sl.ReportError(defaultTrafficRule, fieldName, "defaultTrafficRule", defaultTrafficRuleDescription, "")
	}
}

func validateDefaultRouteTrafficRules(
	sl validator.StructLevel,
	fieldName string,
	trafficRules models.TrafficRules,
	defaultRouteID *string,
) {
	// If Traffic Rules are configured, DefaultRouteId should be present in all rules
	if defaultRouteID != nil {
		missingDefaultRouteDescription := fmt.Sprintf(
			"Fallback Route (DefaultRouteId): '%s' should be associated to all Traffic Rules", *defaultRouteID)
		for _, rule := range trafficRules {
			containsDefaultRouteID := false
			for _, route := range rule.Routes {
				if route == *defaultRouteID {
					containsDefaultRouteID = true
					break
				}
			}
			if !containsDefaultRouteID {
				sl.ReportError(defaultRouteID, fieldName, fieldName, missingDefaultRouteDescription, "")
			}
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

func validateConditionOrthogonality(
	sl validator.StructLevel,
	fieldName string,
	allRules models.TrafficRules,
) {
	trafficConditionsValueMap := map[string]map[string]*set.Set{}
	for _, rule := range allRules {
		trafficConditionsValueMap[rule.Name] = map[string]*set.Set{}
		for _, condition := range rule.Conditions {
			conditionValues := []interface{}{}
			for _, val := range condition.Values {
				conditionValues = append(conditionValues, val)
			}
			trafficConditionsValueMap[rule.Name][condition.Field] = set.New(conditionValues...)
		}
	}

	allErrors := []string{}
	for idx, rule1 := range allRules {
		rule1Fields, ok := trafficConditionsValueMap[rule1.Name]
		// Check that there are rule conditions
		if ok {
			// Check that the current rule and the other are orthogonal
			rulesOverlap := true

			for _, rule2 := range allRules[idx+1:] {
				// Rules with same name should be skipped for orthogonality checks since they will fail unique name validation
				if rule1.Name == rule2.Name {
					continue
				}
				for field, currSet := range rule1Fields {
					isCurrValEmpty, isOtherValEmpty := false, false
					if !ok || currSet == nil || currSet.Len() == 0 {
						isCurrValEmpty = true
					}
					otherSet, ok := trafficConditionsValueMap[rule2.Name][field]
					if !ok || otherSet == nil || otherSet.Len() == 0 {
						isOtherValEmpty = true
					}

					// If both values non-empty, check overlap.
					// If only one of the values is empty, we can skip further checks.
					// If both empty, nothing to do.
					if !isCurrValEmpty && !isOtherValEmpty {
						if currSet.Intersection(otherSet).Len() == 0 {
							// At least one field value does not overlap, we can terminate the check for
							// this other rule.
							rulesOverlap = false
							break
						}
					} else if !isCurrValEmpty || !isOtherValEmpty {
						rulesOverlap = false
						break
					}
				}

				if rulesOverlap {
					allErrors = append(allErrors, fmt.Sprintf("(%s,%s)", rule1.Name, rule2.Name))
				}
			}
		}
	}

	if len(allErrors) != 0 {
		errorMessage := fmt.Sprintf(
			"Rules Orthogonality check failed, following pairs of rules are overlapping - %s.", strings.Join(allErrors, ", "),
		)
		sl.ReportError(allRules, fieldName, "TrafficRules", errorMessage, "")
	}
}

func validateRouterConfig(sl validator.StructLevel) {
	router := sl.Current().Interface().(request.RouterConfig)
	instance := sl.Validator()

	routeIds := make([]string, len(router.Routes))
	for idx, route := range router.Routes {
		routeIds[idx] = route.ID
	}
	routeIdsStr := strings.Join(routeIds, " ")

	// Validate default route
	if router.Ensembler == nil || router.Ensembler.Type == models.EnsemblerStandardType {
		if router.DefaultRouteID == nil {
			sl.ReportError(router.DefaultRouteID, "default_route_id", "DefaultRouteID",
				"should be set for chosen ensembler type", "")
		} else {
			if err := instance.Var(*router.DefaultRouteID, fmt.Sprintf("oneof=%s", routeIdsStr)); err != nil {
				ns := "DefaultRouteID"
				sl.ReportValidationErrors(ns, ns, err.(validator.ValidationErrors))
			}
		}
	} else if router.DefaultRouteID != nil && *router.DefaultRouteID != "" {
		sl.ReportError(router.DefaultRouteID, "default_route_id", "DefaultRouteID",
			"should not be set for chosen ensembler type", *router.DefaultRouteID)
	}

	// Validate traffic rules
	allRuleRoutesSet := set.New()
	if router.TrafficRules != nil {
		if len(router.TrafficRules) > 0 {
			checkDefaultTrafficRule(sl, "DefaultTrafficRule", router.DefaultTrafficRule)
			if router.DefaultTrafficRule != nil {
				allRules := append(router.TrafficRules, &models.TrafficRule{
					Name:   "default-traffic-rule",
					Routes: router.DefaultTrafficRule.Routes,
				})
				validateDefaultRouteTrafficRules(sl, "TrafficRules", allRules, router.DefaultRouteID)
				for _, route := range router.DefaultTrafficRule.Routes {
					allRuleRoutesSet.Insert(route)
				}
			}
		}
		for ruleIdx, rule := range router.TrafficRules {
			checkTrafficRuleName(sl, "TrafficRule", rule.Name)
			if rule.Routes != nil {
				for idx, routeID := range rule.Routes {
					allRuleRoutesSet.Insert(routeID)
					ns := fmt.Sprintf("TrafficRules[%d].Routes[%d]", ruleIdx, idx)
					if err := instance.Var(routeID, fmt.Sprintf("oneof=%s", routeIdsStr)); err != nil {
						sl.ReportValidationErrors(ns, ns, err.(validator.ValidationErrors))
					}
				}
			}

			if rule.Conditions != nil {
				// validate the field source of traffic rules are valid for given protocol
				allowedFieldSource := []string{string(expRequest.HeaderFieldSource),
					string(expRequest.PayloadFieldSource)}
				if router.Protocol != nil && *router.Protocol == routerConfig.UPI {
					allowedFieldSource = []string{string(expRequest.HeaderFieldSource),
						string(expRequest.PredictionContextSource)}
				}
				allowedFieldSourceStr := strings.Join(allowedFieldSource, " ")

				for condIdx, cond := range rule.Conditions {
					ns := fmt.Sprintf("TrafficRules[%d].Conditions[%d].FieldSource", ruleIdx, condIdx)
					err := instance.Var(cond.FieldSource, fmt.Sprintf("oneof=%s", allowedFieldSourceStr))
					if err != nil {
						sl.ReportError(router.TrafficRules[ruleIdx].Conditions[condIdx].FieldSource, ns,
							"FieldSource", "oneof", "")
					}
				}
			}
		}
	}

	// Validate dangling routes and traffic rules orthogonality checks
	if router.TrafficRules != nil && len(router.TrafficRules) > 0 {
		checkDanglingRoutes(sl, "Routes", router.Routes, allRuleRoutesSet)
		validateConditionOrthogonality(sl, "TrafficRules", router.TrafficRules)
	}
}
