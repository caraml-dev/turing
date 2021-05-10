package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	mlp "github.com/gojek/mlp/client"
	"github.com/gojek/mlp/pkg/authz/enforcer"
)

const (
	resourceExperimentEngines = "experiment-engines"
)

// NewAuthorizer initialises the Turing specific policies on the given auth enforcer
// and creates a new authorization middleware.
func NewAuthorizer(enforcer enforcer.Enforcer) (*Authorizer, error) {
	// Set up Turing API specific policies
	err := upsertExperimentEnginesListAllPolicy(enforcer)
	if err != nil {
		return nil, err
	}

	return &Authorizer{authEnforcer: enforcer}, nil
}

type Authorizer struct {
	authEnforcer enforcer.Enforcer
}

var methodMapping = map[string]string{
	http.MethodGet:     enforcer.ActionRead,
	http.MethodHead:    enforcer.ActionRead,
	http.MethodPost:    enforcer.ActionCreate,
	http.MethodPut:     enforcer.ActionUpdate,
	http.MethodPatch:   enforcer.ActionUpdate,
	http.MethodDelete:  enforcer.ActionDelete,
	http.MethodConnect: enforcer.ActionRead,
	http.MethodOptions: enforcer.ActionRead,
	http.MethodTrace:   enforcer.ActionRead,
}

func (a *Authorizer) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		resource := getResourceFromPath(r.URL.Path)
		action := getActionFromMethod(r.Method)
		user := r.Header.Get("User-Email")

		allowed, err := a.authEnforcer.Enforce(user, resource, action)
		if err != nil {
			jsonError(w, fmt.Sprintf("Error while checking authorization: %s", err), http.StatusInternalServerError)
			return
		}
		if !*allowed {
			jsonError(w,
				fmt.Sprintf("%s is not authorized to execute %s on %s", user, action, resource),
				http.StatusUnauthorized,
			)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (a *Authorizer) FilterAuthorizedProjects(
	user string,
	projects []mlp.Project,
	action string,
) ([]mlp.Project, error) {
	projectIDs := make([]string, 0)
	var allowedProjects []mlp.Project
	projectMap := make(map[string]mlp.Project)
	for _, project := range projects {
		projectID := fmt.Sprintf("projects:%d", project.Id)
		projectIDs = append(projectIDs, projectID)
		projectMap[projectID] = project
	}

	allowedProjectIds, err := a.authEnforcer.FilterAuthorizedResource(user, projectIDs, action)
	if err != nil {
		return nil, err
	}

	for _, projectID := range allowedProjectIds {
		allowedProjects = append(allowedProjects, projectMap[projectID])
	}

	return allowedProjects, nil
}

func getResourceFromPath(path string) string {
	return strings.Replace(strings.TrimPrefix(path, "/"), "/", ":", -1)
}

func getActionFromMethod(method string) string {
	return methodMapping[method]
}

func upsertExperimentEnginesListAllPolicy(authEnforcer enforcer.Enforcer) error {
	subresource := fmt.Sprintf("%s:**", resourceExperimentEngines)

	// Upsert policy
	policyName := fmt.Sprintf("allow-all-list-%s", resourceExperimentEngines)
	_, err := authEnforcer.UpsertPolicy(
		policyName,
		[]string{},
		[]string{"**"},
		[]string{resourceExperimentEngines, subresource},
		[]string{enforcer.ActionRead},
	)
	return err
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	if len(msg) > 0 {
		errJSON, _ := json.Marshal(struct {
			Error string `json:"error"`
		}{msg})

		_, _ = w.Write(errJSON)
	}
}
