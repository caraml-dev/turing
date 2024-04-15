package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/caraml-dev/turing/engines/router"
)

type TrafficRule struct {
	Name       string                         `json:"name" validate:"required,notBlank"`
	Conditions []*router.TrafficRuleCondition `json:"conditions" validate:"required,notBlank,dive"`
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

func (r *TrafficRules) ConditionalRouteIDs() map[string]bool {
	distinctRouteIDs := map[string]bool{}

	for _, rule := range *r {
		for _, rID := range rule.Routes {
			distinctRouteIDs[rID] = true
		}
	}
	return distinctRouteIDs
}
