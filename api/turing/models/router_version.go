package models

import (
	"database/sql"

	"github.com/jinzhu/gorm"
)

type RouterVersionStatus string

const (
	RouterVersionStatusPending    RouterVersionStatus = "pending"
	RouterVersionStatusFailed     RouterVersionStatus = "failed"
	RouterVersionStatusDeployed   RouterVersionStatus = "deployed"
	RouterVersionStatusUndeployed RouterVersionStatus = "undeployed"
)

// RouterVersion contains the configuration of a version of a router.
// Every change in configuration should always result in a new instance of
// RouterVersion.
type RouterVersion struct {
	Model
	// Router this RouterVersion is associated with.
	RouterID ID      `json:"-"`
	Router   *Router `json:"router" gorm:"association_autoupdate:false"`

	// Version of Router configuration.
	Version uint `json:"version"`

	// Status of the RouterVersion. Indicates the deployment status of the configuration.
	Status RouterVersionStatus `json:"status"`
	// Last known error if the status is error
	Error string `json:"error,omitempty"`
	// Image of the router deployed
	Image string `json:"image"`
	// Downstream endpoints for the router
	Routes Routes `json:"routes"`
	// Default route
	DefaultRouteID string `json:"default_route_id"`
	// Rules for activating some routes based on request conditions.
	TrafficRules TrafficRules `json:"rules,omitempty"`
	// Configuration for the experiment engine queried by the router.
	ExperimentEngine *ExperimentEngine `json:"experiment_engine"`
	// Resource requests for deployment
	ResourceRequest *ResourceRequest `json:"resource_request"`
	// Request timeout as a valid quantity string
	Timeout string `json:"timeout"`
	// Logging configuration for the router
	LogConfig *LogConfig `json:"log_config"`

	// The enricher used by the router
	EnricherID sql.NullInt32 `json:"-"`
	Enricher   *Enricher     `json:"enricher,omitempty"`

	// The ensembler used by the router
	EnsemblerID sql.NullInt32 `json:"-"`
	Ensembler   *Ensembler    `json:"ensembler,omitempty"`

	// Monitoring URL used in the monitoring tab
	MonitoringURL string `json:"monitoring_url" gorm:"-"`
}

// SetEnsemblerID Sets the id of the associated Ensembler
func (r *RouterVersion) SetEnsemblerID(ensemblerID ID) {
	r.EnsemblerID = sql.NullInt32{Int32: int32(ensemblerID), Valid: true}
}

// SetEnricherID Sets the id of the associated Enricher
func (r *RouterVersion) SetEnricherID(enricherID ID) {
	r.EnricherID = sql.NullInt32{Int32: int32(enricherID), Valid: true}
}

// BeforeCreate Sets version before creating
func (r *RouterVersion) BeforeCreate(tx *gorm.DB) error {
	var latestVersion RouterVersion
	err := tx.Select("router_versions.*").
		Where("router_id = ?", r.RouterID).
		Order("version desc").
		FirstOrInit(&latestVersion, &RouterVersion{Version: 0}).Error
	if err != nil {
		return err
	}
	r.Version = latestVersion.Version + 1
	return nil
}
