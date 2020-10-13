package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/heptiolabs/healthcheck"

	"github.com/gojek/turing/engines/router/missionctl/internal"
)

func versionAPI(w http.ResponseWriter, r *http.Request) {
	dec := json.NewEncoder(w)
	dec.SetIndent("", "  ")
	if err := dec.Encode(internal.VersionInfo); err != nil {
		http.Error(w, fmt.Sprintf("error encoding JSON: %s", err), http.StatusInternalServerError)
	}
}

// NewInternalAPIHandler creates an instance of the internal api handler
func NewInternalAPIHandler() http.Handler {
	h := http.NewServeMux()

	h.Handle("/", healthcheck.NewHandler())
	h.HandleFunc("/version", versionAPI)

	return h
}
