package server

import (
	"time"

	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
	"github.com/jinzhu/gorm"
)

func AddHealthCheckHandler(r *mux.Router, path string, db *gorm.DB) {
	sub := r.PathPrefix(path).Subrouter()

	health := healthcheck.NewHandler()
	health.AddReadinessCheck(
		"database",
		healthcheck.DatabasePingCheck(db.DB(), 1*time.Second))

	sub.Path("/live").HandlerFunc(health.LiveEndpoint)
	sub.Path("/ready").HandlerFunc(health.ReadyEndpoint)
}
