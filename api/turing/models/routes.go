package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	fiberConfig "github.com/gojek/fiber/config"
	fiberProtocol "github.com/gojek/fiber/protocol"
)

// RouteProtocol is the type used to capture the supported response
// logging destinations
type RouteProtocol fiberProtocol.Protocol

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
	// Fiber protocol
	Protocol RouteProtocol `json:"protocol"`
	// Grpc ServiceMethod name
	ServiceMethod string `json:"service_method,omitempty"`
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
func (r *Routes) ToFiberRoutes() (*fiberConfig.Routes, error) {
	routes := fiberConfig.Routes{}
	for _, route := range *r {
		timeout, err := time.ParseDuration(route.Timeout)
		if err != nil {
			return nil, err
		}
		if string(route.Protocol) != string(fiberProtocol.HTTP) &&
			string(route.Protocol) != string(fiberProtocol.GRPC) {
			return nil, fmt.Errorf("invalid route protocol for %s", route.ID)
		}
		routes = append(routes, &fiberConfig.ProxyConfig{
			ComponentConfig: fiberConfig.ComponentConfig{
				ID:   route.ID,
				Type: route.Type,
			},
			Endpoint: route.Endpoint,
			Protocol: fiberProtocol.Protocol(route.Protocol),
			Timeout:  fiberConfig.Duration(timeout),
			GrpcConfig: fiberConfig.GrpcConfig{
				ServiceMethod: route.ServiceMethod,
			},
		})
	}
	return &routes, nil
}
