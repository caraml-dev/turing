package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"

	"github.com/gojek/turing/engines/experiment/common"
)

type TrafficRuleCondition struct {
	FieldSource common.FieldSource    `json:"field_source" validate:"required,oneof=header payload"`
	Field       string                `json:"field" validate:"required"`
	Operator    RuleConditionOperator `json:"operator" validate:"required,oneof=in"`
	Values      []string              `json:"values" validate:"required,notBlank"`
}

func (c *TrafficRuleCondition) TestRequest(reqHeader http.Header, bodyBytes []byte) (bool, error) {
	fieldValue, err := common.GetValueFromRequest(reqHeader, bodyBytes, c.FieldSource, c.Field)
	if err != nil {
		return false, err
	}
	return c.Operator.Test(fieldValue, c.Values)
}

type RuleConditionOperator struct {
	Operator
}

func (o RuleConditionOperator) String() string {
	if o.Operator != nil {
		return o.Operator.String()
	}
	return ""
}

func (o RuleConditionOperator) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.String())
}

func (o *RuleConditionOperator) UnmarshalJSON(data []byte) error {
	var operatorName string
	if err := json.Unmarshal(data, &operatorName); err != nil {
		return err
	}

	switch operatorName {
	case InConditionOperator.String():
		*o = InConditionOperator
	default:
		return fmt.Errorf("unknown operator: %s", operatorName)
	}

	return nil
}

type Operator interface {
	String() string
	Test(left interface{}, right interface{}) (bool, error)
}

type inConditionOperator struct{}

func (o *inConditionOperator) String() string {
	return "in"
}

func (o *inConditionOperator) Test(left interface{}, right interface{}) (bool, error) {
	typeOf := reflect.ValueOf(right)

	switch typeOf.Kind() {
	case reflect.Slice:
		for i := 0; i < typeOf.Len(); i++ {
			if left == typeOf.Index(i).Interface() {
				return true, nil
			}
		}
	default:
		return false, fmt.Errorf("invalid type of right argument: slice is expected")
	}

	return false, nil
}

var (
	InConditionOperator RuleConditionOperator = RuleConditionOperator{&inConditionOperator{}}
)
