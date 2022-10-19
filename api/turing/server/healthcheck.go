package server

import (
	"database/sql"
	"time"

	"github.com/gorilla/mux"
	"github.com/heptiolabs/healthcheck"
)

func AddHealthCheckHandler(r *mux.Router, path string, db *sql.DB) {
	sub := r.PathPrefix(path).Subrouter()

	health := healthcheck.NewHandler()
	health.AddReadinessCheck(
		"database",
		healthcheck.DatabasePingCheck(db, 1*time.Second))

	sub.Path("/live").HandlerFunc(health.LiveEndpoint)
	sub.Path("/ready").HandlerFunc(health.ReadyEndpoint)
}
