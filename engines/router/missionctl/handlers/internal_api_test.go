package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/caraml-dev/turing/engines/router/missionctl/internal"
	tu "github.com/caraml-dev/turing/engines/router/missionctl/internal/testutils"
)

func TestVersionAPI(t *testing.T) {
	// Patch internal.VersionInfo
	currentVersionInfo := internal.VersionInfo
	internal.VersionInfo = &internal.Info{
		Version:   "1.2.3",
		Branch:    "test_branch",
		BuildUser: "test_user",
		BuildDate: "01.01.2020",
		GoVersion: "1.12",
	}
	// Reset internal.VersionInfo
	defer func() {
		internal.VersionInfo = currentVersionInfo
	}()

	handler := NewInternalAPIHandler(nil)
	// Request the version API
	req, err := http.NewRequest(http.MethodGet, "/version", nil)
	tu.FailOnError(t, err)
	rr := httptest.NewRecorder()
	http.HandlerFunc(handler.ServeHTTP).ServeHTTP(rr, req)

	// Validate
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.JSONEq(t, `
		{
			"version": "1.2.3",
			"branch": "test_branch",
			"build_user": "test_user",
			"build_date": "01.01.2020",
			"go_version": "1.12"
		}
	`, rr.Body.String())
}

func TestNewHealthcheckHandler(t *testing.T) {
	tests := map[string]struct {
		serviceURLs []string
		wantCode    int
	}{
		"Nil seviceURLs": {
			serviceURLs: nil,
			wantCode:    http.StatusOK,
		},
		"Empty seviceURLs": {
			serviceURLs: []string{},
			wantCode:    http.StatusOK,
		},
		"Empty string seviceURLs": {
			serviceURLs: []string{""},
			wantCode:    http.StatusOK,
		},
		"Resolvable seviceURLs": {
			serviceURLs: []string{""},
			wantCode:    http.StatusOK,
		},
		"All non-resolvable seviceURLs": {
			serviceURLs: []string{"http://ttthis-host-should-not-exist.com"},
			wantCode:    http.StatusServiceUnavailable,
		},
		"Some non-resolvable seviceURLs": {
			serviceURLs: []string{"http://ttthis-host-should-not-exist.com", "http://google.com"},
			wantCode:    http.StatusServiceUnavailable,
		},
		"Invalid seviceURLs": {
			serviceURLs: []string{"invalid-url"},
			wantCode:    http.StatusServiceUnavailable,
		},
	}
	for testName, tt := range tests {
		t.Run(testName, func(t *testing.T) {
			handler := newHealthcheckHandler(tt.serviceURLs).(http.Handler)
			req, err := http.NewRequest("GET", "/ready", nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.wantCode {
				t.Errorf("newHealthcheckHandler() = %v, want %v", rr.Code, tt.wantCode)
			}
		})
	}
}
