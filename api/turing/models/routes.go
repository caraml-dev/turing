package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	fiberconfig "github.com/gojek/fiber/config"
)

// Route maps onto the fiber.Component.
type Route struct {
	// ID of the route
	ID string `json:"id"`
	// Type of the route
	Type string `json:"type"`
	// Endpoint to query
	Endpoint string `json:"endpoint"`
	// Annotations (optional) holds extra information about the route
	Annotations map[string]string `json:"annotations"`
	// Request timeout as a valid quantity string.
	Timeout string `json:"timeout"`
}

type Routes []*Route

func (r Routes) Value() (driver.Value, error) {
	return json.Marshal(r)
}

func (r *Routes) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}

	return json.Unmarshal(b, &r)
}

// ToFiberRoutes converts routes to a type compatible with Fiber's config
func (r *Routes) ToFiberRoutes() (*fiberconfig.Routes, error) {
	routes := fiberconfig.Routes{}
	for _, route := range *r {
		timeout, err := time.ParseDuration(route.Timeout)
		if err != nil {
			return nil, err
		}
		routes = append(routes, &fiberconfig.ProxyConfig{
			ComponentConfig: fiberconfig.ComponentConfig{
				ID:   route.ID,
				Type: route.Type,
			},
			Endpoint: route.Endpoint,
			Timeout:  fiberconfig.Duration(timeout),
		})
	}
	return &routes, nil
}
