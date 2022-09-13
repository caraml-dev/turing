package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
)

type DefaultTrafficRule struct {
	Routes []string `json:"routes" validate:"required,notBlank"`
}

func (r DefaultTrafficRule) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *DefaultTrafficRule) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, r)
}
