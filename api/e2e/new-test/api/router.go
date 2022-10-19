//go:build e2e

package api

import (
	"net/http"

	"github.com/gavv/httpexpect/v2"
)

var Status = struct {
	Deployed   string
	Failed     string
	Pending    string
	Undeployed string
}{
	Deployed:   "deployed",
	Failed:     "failed",
	Pending:    "pending",
	Undeployed: "undeployed",
}

func GetRouter(e *httpexpect.Expect, projectID, routerID interface{}) *httpexpect.Object {
	return e.GET("/projects/{projectId}/routers/{routerId}").
		WithPath("projectId", projectID).
		WithPath("routerId", routerID).
		Expect().Status(http.StatusOK).
		JSON().Object()
}

func GetRouterVersion(e *httpexpect.Expect, projectID, routerID, version interface{}) *httpexpect.Object {
	return e.GET("/projects/{projectId}/routers/{routerId}/versions/{version}").
		WithPath("projectId", projectID).
		WithPath("routerId", routerID).
		WithPath("version", version).
		Expect().Status(http.StatusOK).
		JSON().Object()
}
