package api

import (
	"database/sql"
	"errors"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/gojek/turing/api/turing/service"

	"github.com/gojek/mlp/client"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service/mocks"
)

func TestPodLogControllerListPodLogs(t *testing.T) {
	podLogService := &mocks.PodLogService{}
	mlpService := &mocks.MLPService{}
	routersService := &mocks.RoutersService{}
	routerVersionsService := &mocks.RouterVersionsService{}

	project := &client.Project{Name: "project1"}
	router1 := &models.Router{Model: models.Model{ID: 1}, CurrRouterVersionID: sql.NullInt32{Int32: 1, Valid: true}}
	// Simulate router where the CurrRouterVersionID value is invalid
	router2 := &models.Router{Model: models.Model{ID: 2}, CurrRouterVersionID: sql.NullInt32{Int32: 2, Valid: false}}
	routerVersion1 := &models.RouterVersion{Model: models.Model{ID: 1}}
	routerVersion2 := &models.RouterVersion{Model: models.Model{ID: 2}}
	// Simulate error in retrieving router's current version
	router3 := &models.Router{Model: models.Model{ID: 3}, CurrRouterVersionID: sql.NullInt32{Int32: 3, Valid: true}}

	sinceTime := time.Date(2020, 12, 5, 8, 0, 0, 0, time.UTC)
	tailLines := int64(5)
	headLines := int64(3)
	podLogOptions := &service.PodLogOptions{
		Container: "mycontainer",
		Previous:  true,
		SinceTime: &sinceTime,
		TailLines: &tailLines,
		HeadLines: &headLines,
	}

	mlpService.On("GetProject", 1).Return(project, nil)
	routersService.On("FindByID", uint(1)).Return(router1, nil)
	routersService.On("FindByID", uint(2)).Return(router2, nil)
	routersService.On("FindByID", uint(3)).Return(router3, nil)
	routerVersionsService.On("FindByID", uint(1)).Return(routerVersion1, nil)
	routerVersionsService.On("FindByID", uint(2)).Return(routerVersion2, nil)
	routerVersionsService.On("FindByID", uint(3)).Return(nil, errors.New("Test router version error"))
	// Simulate error when router with id 3 is requested
	routerVersionsService.On("FindByID", uint(3)).Return(nil, errors.New(""))
	routerVersionsService.On("FindByRouterIDAndVersion", uint(1), uint(1)).Return(routerVersion1, nil)
	routerVersionsService.On("FindByRouterIDAndVersion", uint(1), uint(2)).Return(routerVersion2, nil)
	// Simulate error when router with id 1 and version 3 is requested
	routerVersionsService.On("FindByRouterIDAndVersion", uint(1), uint(3)).Return(nil, errors.New(""))
	podLogService.
		On("ListPodLogs", project, router1, routerVersion1, "router", &service.PodLogOptions{}).
		Return([]*service.PodLog{{TextPayload: "routerVersion1"}}, nil)
	podLogService.
		On("ListPodLogs", project, router1, routerVersion2, "router", &service.PodLogOptions{}).
		Return([]*service.PodLog{{TextPayload: "routerVersion2"}}, nil)
	podLogService.
		On("ListPodLogs", project, router1, routerVersion1, "enricher", podLogOptions).
		Return([]*service.PodLog{{TextPayload: "valid optional args"}}, nil)
	// Simulate error when logs for router with component 'ensembler' is requested
	podLogService.
		On("ListPodLogs", project, router1, routerVersion2, "ensembler", &service.PodLogOptions{}).
		Return([]*service.PodLog{}, errors.New("Test Pod Log error"))

	type args struct {
		r    *http.Request
		vars map[string]string
		body interface{}
	}
	tests := []struct {
		name string
		args args
		want *Response
	}{
		{
			name: "missing project_id",
			args: args{
				vars: map[string]string{},
			},
			want: BadRequest("invalid project id", "key project_id not found in vars"),
		},
		{
			name: "missing router_id",
			args: args{
				vars: map[string]string{
					"project_id": "1",
				},
			},
			want: BadRequest("invalid router id", "key router_id not found in vars"),
		},
		{
			name: "expected args",
			args: args{
				vars: map[string]string{
					"project_id": "1",
					"router_id":  "1",
				},
			},
			want: Ok([]*service.PodLog{{TextPayload: "routerVersion1"}}),
		},
		{
			name: "specific router version id",
			args: args{
				vars: map[string]string{
					"project_id": "1",
					"router_id":  "1",
					"version":    "2",
				},
			},
			want: Ok([]*service.PodLog{{TextPayload: "routerVersion2"}}),
		},
		{
			name: "invalid router version id",
			args: args{
				vars: map[string]string{
					"project_id": "1",
					"router_id":  "1",
					"version":    "3",
				},
			},
			want: NotFound("router version not found", ""),
		},
		{
			name: "invalid router version id reference in router",
			args: args{
				vars: map[string]string{
					"project_id": "1",
					"router_id":  "2",
				},
			},
			want: BadRequest("Current router version id is invalid", "Make sure current router is deployed"),
		},
		{
			name: "current version not found",
			args: args{
				vars: map[string]string{
					"project_id": "1",
					"router_id":  "3",
				},
			},
			want: InternalServerError("Failed to find current router version", "Test router version error"),
		},
		{
			name: "valid optional args",
			args: args{
				vars: map[string]string{
					"project_id":     "1",
					"router_id":      "1",
					"version":        "1",
					"component_type": "enricher",
					"container":      "mycontainer",
					"previous":       "true",
					"since_time":     "2020-12-05T08:00:00Z",
					"tail_lines":     "5",
					"head_lines":     "3",
				},
			},
			want: Ok([]*service.PodLog{{TextPayload: "valid optional args"}}),
		},
		{
			name: "invalid component_type",
			args: args{
				vars: map[string]string{
					"project_id":     "1",
					"router_id":      "1",
					"version":        "1",
					"component_type": "invalidcomponenttype",
					"container":      "mycontainer",
					"previous":       "true",
					"since_time":     "2020-12-05T08:00:00Z",
					"tail_lines":     "5",
				},
			},
			want: BadRequest("Invalid component type 'invalidcomponenttype'",
				"must be one of router, enricher or ensembler"),
		},
		{
			name: "invalid since_time time format",
			args: args{
				vars: map[string]string{
					"project_id": "1",
					"router_id":  "1",
					"version":    "1",
					"since_time": "2020-1205T08:00:00Z",
				},
			},
			want: BadRequest("Query string 'since_time' must be in RFC3339 format",
				`parsing time "2020-1205T08:00:00Z" as "2006-01-02T15:04:05Z07:00": cannot parse "05T08:00:00Z" as "-"`),
		},
		{
			name: "invalid previous arg",
			args: args{
				vars: map[string]string{
					"project_id": "1",
					"router_id":  "1",
					"version":    "1",
					"previous":   "yes",
				},
			},
			want: BadRequest("Query string 'previous' must be a truthy value",
				`strconv.ParseBool: parsing "yes": invalid syntax`),
		},
		{
			name: "invalid tail_lines arg",
			args: args{
				vars: map[string]string{
					"project_id": "1",
					"router_id":  "1",
					"version":    "1",
					"tail_lines": "five",
				},
			},
			want: BadRequest("Query string 'tail_lines' must be a positive number",
				`strconv.ParseInt: parsing "five": invalid syntax`),
		},
		{
			name: "negative tail_lines arg",
			args: args{
				vars: map[string]string{
					"project_id": "1",
					"router_id":  "1",
					"version":    "1",
					"tail_lines": "-1",
				},
			},
			want: BadRequest("Query string 'tail_lines' must be a positive number", ""),
		},
		{
			name: "invalid head_lines arg",
			args: args{
				vars: map[string]string{
					"project_id": "1",
					"router_id":  "1",
					"version":    "1",
					"head_lines": "five",
				},
			},
			want: BadRequest("Query string 'head_lines' must be a positive number",
				`strconv.ParseInt: parsing "five": invalid syntax`),
		},
		{
			name: "negative head_lines arg",
			args: args{
				vars: map[string]string{
					"project_id": "1",
					"router_id":  "1",
					"version":    "1",
					"head_lines": "-10",
				},
			},
			want: BadRequest("Query string 'head_lines' must be a positive number", ""),
		},
		{
			name: "list logs error",
			args: args{
				vars: map[string]string{
					"project_id":     "1",
					"router_id":      "1",
					"version":        "2",
					"component_type": "ensembler",
				},
			},
			want: InternalServerError("Failed to list logs", "Test Pod Log error"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := PodLogController{
				&baseController{
					&AppContext{
						PodLogService:         podLogService,
						MLPService:            mlpService,
						RoutersService:        routersService,
						RouterVersionsService: routerVersionsService,
					},
				},
			}
			if got := c.ListPodLogs(tt.args.r, tt.args.vars, tt.args.body); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListPodLogs() = %v, want %v", got, tt.want)
			}
		})
	}
}
