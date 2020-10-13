package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/gojek/turing/engines/experiment/common"
)

type TrafficRule struct {
	Conditions []*common.TrafficRuleCondition `json:"conditions" validate:"required,notBlank,dive"`
	Routes     []string                       `json:"routes" validate:"required,notBlank"`
}

type TrafficRules []*TrafficRule

func (r TrafficRules) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *TrafficRules) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, r)
}

func (r *TrafficRules) ConditionalRouteIds() map[string]bool {
	distinctRouteIds := map[string]bool{}

	for _, rule := range *r {
		for _, rID := range rule.Routes {
			distinctRouteIds[rID] = true
		}
	}
	return distinctRouteIds
}
