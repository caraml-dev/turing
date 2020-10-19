package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

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

// NewInternalAPIHandler creates an instance of the internal api handler.
//
// It accepts a list of service URLs that Turing router will accept or access.
// These service URLs will be checked, to ensure they are resolvable, as part of the
// readiness check. This is because sometimes it can take several seconds for the URL host to
// be resolvable.
func NewInternalAPIHandler(serviceURLs []string) http.Handler {
	h := http.NewServeMux()

	h.Handle("/", newHealthcheckHandler(serviceURLs))
	h.HandleFunc("/version", versionAPI)

	return h
}

func newHealthcheckHandler(serviceURLs []string) healthcheck.Handler {
	health := healthcheck.NewHandler()
	for i, serviceURL := range serviceURLs {
		checkName := fmt.Sprintf("url-resolvable-%d", i)
		health.AddReadinessCheck(checkName, checkURLResolvable(serviceURL))
	}
	return health
}

var defaultTimeoutForDNSLookup = 100 * time.Millisecond

func checkURLResolvable(rawURL string) healthcheck.Check {
	resolver := net.Resolver{}
	return func() error {
		if rawURL == "" {
			return nil
		}

		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutForDNSLookup)
		defer cancel()

		parsedURL, err := url.Parse(rawURL)
		if err != nil {
			return err
		}

		addrs, err := resolver.LookupHost(ctx, parsedURL.Host)
		if err != nil {
			return err
		}

		if len(addrs) < 1 {
			return fmt.Errorf("could not resolve host for URL: %s", rawURL)
		}

		return nil
	}
}

