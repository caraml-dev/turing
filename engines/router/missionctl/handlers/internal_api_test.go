package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gojek/turing/engines/router/missionctl/internal"
	tu "github.com/gojek/turing/engines/router/missionctl/internal/testutils"
	"github.com/stretchr/testify/assert"
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

	handler := NewInternalAPIHandler()
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
