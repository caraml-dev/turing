package models

import (
	"database/sql"
	"encoding/json"
	"fmt"
)

type RouterStatus string

const (
	RouterStatusPending    RouterStatus = "pending"
	RouterStatusFailed     RouterStatus = "failed"
	RouterStatusDeployed   RouterStatus = "deployed"
	RouterStatusUndeployed RouterStatus = "undeployed"
)

// Router holds the information as to which versions of Router have been deployed,
// and acts as the top level object for an instance of the Turing Router.
type Router struct {
	Model
	// Project id of the project this router belongs to, as retrieved from
	// the MLP API.
	ProjectID ID `json:"project_id"`
	// Environment name of the environment this router belongs to, as retrieved
	// from the MLP API.
	EnvironmentName string `json:"environment_name"`
	// Name of the router. Must be unique within the given project and environment.
	Name string `json:"name"`
	// Status of the Router. Indicates the deployment status of the router.
	Status RouterStatus `json:"status"`
	// Endpoint URL where the currently deployed router version is accessible at
	Endpoint string `json:"endpoint,omitempty"`

	// The current version (may be deployed or undeployed)
	CurrRouterVersionID sql.NullInt32  `json:"-"`
	CurrRouterVersion   *RouterVersion `json:"config,omitempty" gorm:"foreignkey:CurrRouterVersionID"`

	// MonitoringURL is for all router versions
	MonitoringURL string `json:"monitoring_url" gorm:"-"`
}

func (r *Router) SetCurrRouterVersionID(routerVersionID ID) {
	r.CurrRouterVersionID = sql.NullInt32{Int32: int32(routerVersionID), Valid: true}
}

func (r *Router) ClearCurrRouterVersionID() {
	r.CurrRouterVersionID = sql.NullInt32{Int32: 0, Valid: false}
}

// SetCurrRouterVersion sets the currently version for this router to the provided
// routerVersion.
func (r *Router) SetCurrRouterVersion(routerVersion *RouterVersion) {
	r.SetCurrRouterVersionID(routerVersion.ID)
	r.CurrRouterVersion = routerVersion
}

// ClearCurrRouterVersion clears the current version for this router.
func (r *Router) ClearCurrRouterVersion() {
	r.ClearCurrRouterVersionID()
	r.CurrRouterVersion = nil
}

// RouterResponse is an alias for the Router, to enable custom marshaling
type RouterResponse Router

// MarshalJSON is a custom marshaling function for the Router, to manipulate the
// endpoint.
func (r *Router) MarshalJSON() ([]byte, error) {
	response := RouterResponse(*r)
	// Append the path to the endpoint
	if response.Endpoint != "" {
		response.Endpoint = fmt.Sprintf("%s/v1/predict", response.Endpoint)
	}
	return json.Marshal(response)
}
