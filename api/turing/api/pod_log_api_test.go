package api

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"

	"github.com/caraml-dev/turing/api/turing/batch"
	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/cluster/servicebuilder"
	openapi "github.com/caraml-dev/turing/api/turing/generated"
	"github.com/caraml-dev/turing/api/turing/internal/ref"
	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/api/turing/validation"

	"github.com/gojek/mlp/api/client"

	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service/mocks"
)

func TestPodLogControllerListEnsemblingPodLogs(t *testing.T) {
	namespace := "foo"
	environment := "dev"
	loggingURL := "https://www.example.com/hello/world"
	podLogsLegacyFormat := []*service.PodLog{
		{
			Environment: environment,
			Namespace:   namespace,
			Timestamp:   time.Date(2020, 7, 7, 7, 0, 5, 0, time.UTC),
			PodName:     "bar",
			TextPayload: "[INFO] Taking snapshot of full filesystem...",
		},
	}
	ensemblingPodLogs := &service.PodLogsV2{
		Environment: environment,
		Namespace:   namespace,
		LoggingURL:  loggingURL,
		Logs: []*service.PodLogV2{
			{
				Timestamp:   time.Date(2020, 7, 7, 7, 0, 5, 0, time.UTC),
				PodName:     "bar",
				TextPayload: "[INFO] Taking snapshot of full filesystem...",
			},
		},
	}
	ensemblingJob := &models.EnsemblingJob{
		InfraConfig: &models.InfraConfig{
			EnsemblerInfraConfig: openapi.EnsemblerInfraConfig{
				EnsemblerName: ref.String("hello"),
			},
		},
	}

	tests := map[string]struct {
		mlpService           func() service.MLPService
		ensemblingJobService func() service.EnsemblingJobService
		podLogService        func() service.PodLogService
		componentType        string
		vars                 RequestVars
		expected             *Response
	}{
		"success | nominal": {
			mlpService: func() service.MLPService {
				s := &mocks.MLPService{}
				project := &client.Project{Name: "project"}
				s.On("GetProject", models.ID(1)).Return(project, nil)
				return s
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				s := &mocks.EnsemblingJobService{}
				s.On("FindByID", mock.Anything, mock.Anything).Return(ensemblingJob, nil)
				s.On("GetNamespaceByComponent", mock.Anything, mock.Anything).Return(namespace)
				s.On("GetDefaultEnvironment").Return(environment)
				s.On("CreatePodLabelSelector", mock.Anything, mock.Anything).Return([]service.LabelSelector{})
				s.On("FormatLoggingURL", mock.Anything, mock.Anything, mock.Anything).Return(loggingURL, nil)
				return s
			},
			podLogService: func() service.PodLogService {
				s := &mocks.PodLogService{}
				s.On("ListPodLogs", mock.Anything).Return(podLogsLegacyFormat, nil)
				return s
			},
			componentType: "",
			vars: RequestVars{
				"job_id":         {"1"},
				"project_id":     {"1"},
				"previous":       {"true"},
				"since_time":     {"2020-12-05T08:00:00Z"},
				"tail_lines":     {"5"},
				"head_lines":     {"3"},
				"component_type": {batch.ImageBuilderPodType},
			},
			expected: Ok(ensemblingPodLogs),
		},
		"success | head lines empty": {
			mlpService: func() service.MLPService {
				s := &mocks.MLPService{}
				project := &client.Project{Name: "project"}
				s.On("GetProject", models.ID(1)).Return(project, nil)
				return s
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				s := &mocks.EnsemblingJobService{}
				s.On("FindByID", mock.Anything, mock.Anything).Return(ensemblingJob, nil)
				s.On("GetNamespaceByComponent", mock.Anything, mock.Anything).Return(namespace)
				s.On("GetDefaultEnvironment").Return(environment)
				s.On("CreatePodLabelSelector", mock.Anything, mock.Anything).Return([]service.LabelSelector{})
				s.On("FormatLoggingURL", mock.Anything, mock.Anything, mock.Anything).Return(loggingURL, nil)
				return s
			},
			podLogService: func() service.PodLogService {
				s := &mocks.PodLogService{}
				s.On("ListPodLogs", mock.Anything).Return(podLogsLegacyFormat, nil)
				return s
			},
			componentType: "",
			vars: RequestVars{
				"job_id":         {"1"},
				"project_id":     {"1"},
				"previous":       {"true"},
				"tail_lines":     {"5"},
				"component_type": {batch.ImageBuilderPodType},
			},
			expected: Ok(ensemblingPodLogs),
		},
		"success | since date missing": {
			mlpService: func() service.MLPService {
				s := &mocks.MLPService{}
				project := &client.Project{Name: "project"}
				s.On("GetProject", models.ID(1)).Return(project, nil)
				return s
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				s := &mocks.EnsemblingJobService{}
				s.On("FindByID", mock.Anything, mock.Anything).Return(ensemblingJob, nil)
				s.On("GetNamespaceByComponent", mock.Anything, mock.Anything).Return(namespace)
				s.On("GetDefaultEnvironment").Return(environment)
				s.On("CreatePodLabelSelector", mock.Anything, mock.Anything).Return([]service.LabelSelector{})
				s.On("FormatLoggingURL", mock.Anything, mock.Anything, mock.Anything).Return(loggingURL, nil)
				return s
			},
			podLogService: func() service.PodLogService {
				s := &mocks.PodLogService{}
				s.On("ListPodLogs", mock.Anything).Return(podLogsLegacyFormat, nil)
				return s
			},
			componentType: "",
			vars: RequestVars{
				"job_id":         {"1"},
				"project_id":     {"1"},
				"previous":       {"true"},
				"since_time":     {"2020-12-05T08:00:00Z"},
				"tail_lines":     {"5"},
				"component_type": {batch.ImageBuilderPodType},
			},
			expected: Ok(ensemblingPodLogs),
		},
		"failure | negative tail lines": {
			mlpService: func() service.MLPService {
				s := &mocks.MLPService{}
				project := &client.Project{Name: "project"}
				s.On("GetProject", models.ID(1)).Return(project, nil)
				return s
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				s := &mocks.EnsemblingJobService{}
				s.On("FindByID", mock.Anything, mock.Anything).Return(ensemblingJob, nil)
				s.On("GetNamespaceByComponent", mock.Anything, mock.Anything).Return(namespace)
				s.On("GetDefaultEnvironment").Return(environment)
				s.On("CreatePodLabelSelector", mock.Anything, mock.Anything).Return([]service.LabelSelector{})
				s.On("FormatLoggingURL", mock.Anything, mock.Anything, mock.Anything).Return(loggingURL, nil)
				return s
			},
			podLogService: func() service.PodLogService {
				s := &mocks.PodLogService{}
				s.On("ListPodLogs", mock.Anything).Return(podLogsLegacyFormat, nil)
				return s
			},
			componentType: "",
			vars: RequestVars{
				"job_id":         {"1"},
				"project_id":     {"1"},
				"previous":       {"true"},
				"since_time":     {"2020-12-05T08:00:00Z"},
				"tail_lines":     {"-5"},
				"head_lines":     {"3"},
				"component_type": {batch.ImageBuilderPodType},
			},
			expected: BadRequest(
				"failed to fetch ensembling job pod logs",
				"failed to parse query string: Key: 'listEnsemblingPodLogsOptions.podLogOptions.TailLines'"+
					" Error:Field validation for 'TailLines' failed on the 'gte' tag",
			),
		},
		"failure | wrong component type": {
			mlpService: func() service.MLPService {
				s := &mocks.MLPService{}
				project := &client.Project{Name: "project"}
				s.On("GetProject", models.ID(1)).Return(project, nil)
				return s
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				s := &mocks.EnsemblingJobService{}
				s.On("FindByID", mock.Anything, mock.Anything).Return(ensemblingJob, nil)
				s.On("GetNamespaceByComponent", mock.Anything, mock.Anything).Return(namespace)
				s.On("GetDefaultEnvironment").Return(environment)
				s.On("CreatePodLabelSelector", mock.Anything, mock.Anything).Return([]service.LabelSelector{})
				s.On("FormatLoggingURL", mock.Anything, mock.Anything, mock.Anything).Return(loggingURL, nil)
				return s
			},
			podLogService: func() service.PodLogService {
				return &mocks.PodLogService{}
			},
			componentType: "",
			vars: RequestVars{
				"job_id":         {"1"},
				"project_id":     {"1"},
				"previous":       {"true"},
				"since_time":     {"2020-12-05T08:00:00Z"},
				"tail_lines":     {"5"},
				"head_lines":     {"3"},
				"component_type": {"broken_comp"},
			},
			expected: BadRequest(
				"failed to fetch ensembling job pod logs",
				"failed to parse query string: Key: 'listEnsemblingPodLogsOptions.ComponentType'"+
					" Error:Field validation for 'ComponentType' failed on the 'oneof' tag",
			),
		},
		"failure | ensembling job not found": {
			mlpService: func() service.MLPService {
				s := &mocks.MLPService{}
				project := &client.Project{Name: "project"}
				s.On("GetProject", models.ID(1)).Return(project, nil)
				return s
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				s := &mocks.EnsemblingJobService{}
				s.On("FindByID", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("not found"))
				return s
			},
			podLogService: func() service.PodLogService {
				return &mocks.PodLogService{}
			},
			componentType: "",
			vars: RequestVars{
				"job_id":         {"1"},
				"project_id":     {"1"},
				"previous":       {"true"},
				"since_time":     {"2020-12-05T08:00:00Z"},
				"tail_lines":     {"5"},
				"head_lines":     {"3"},
				"component_type": {batch.ImageBuilderPodType},
			},
			expected: NotFound("ensembling job not found", "not found"),
		},
		"failure | fail to list logs": {
			mlpService: func() service.MLPService {
				s := &mocks.MLPService{}
				project := &client.Project{Name: "project"}
				s.On("GetProject", models.ID(1)).Return(project, nil)
				return s
			},
			ensemblingJobService: func() service.EnsemblingJobService {
				s := &mocks.EnsemblingJobService{}
				s.On("FindByID", mock.Anything, mock.Anything).Return(ensemblingJob, nil)
				s.On("GetNamespaceByComponent", mock.Anything, mock.Anything).Return(namespace)
				s.On("GetDefaultEnvironment").Return(environment)
				s.On("CreatePodLabelSelector", mock.Anything, mock.Anything).Return([]service.LabelSelector{})
				return s
			},
			podLogService: func() service.PodLogService {
				s := &mocks.PodLogService{}
				s.On("ListPodLogs", mock.Anything).Return(nil, fmt.Errorf("error"))
				return s
			},
			componentType: "",
			vars: RequestVars{
				"job_id":         {"1"},
				"project_id":     {"1"},
				"previous":       {"true"},
				"since_time":     {"2020-12-05T08:00:00Z"},
				"tail_lines":     {"5"},
				"head_lines":     {"3"},
				"component_type": {batch.ImageBuilderPodType},
			},
			expected: InternalServerError("Failed to list logs", "error"),
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			validator, _ := validation.NewValidator(nil)
			c := PodLogController{
				NewBaseController(
					&AppContext{
						PodLogService:        tt.podLogService(),
						MLPService:           tt.mlpService(),
						EnsemblingJobService: tt.ensemblingJobService(),
					},
					validator,
				),
			}
			if got := c.ListEnsemblingJobPodLogs(nil, tt.vars, nil); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("ListEnsemblingJobPodLogs() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPodLogControllerListRouterPodLogs(t *testing.T) {

	project := &client.Project{Name: "project1"}

	router1 := &models.Router{Model: models.Model{ID: 1}, CurrRouterVersionID: sql.NullInt32{Int32: 1, Valid: true}}
	// Simulate router where the CurrRouterVersionID value is invalid
	router2 := &models.Router{Model: models.Model{ID: 2}, CurrRouterVersionID: sql.NullInt32{Int32: 2, Valid: false}}

	routerVersion1 := &models.RouterVersion{Model: models.Model{ID: 1}, Router: &models.Router{Name: "hello"}}
	routerVersion2 := &models.RouterVersion{Model: models.Model{ID: 2}, Router: &models.Router{Name: "hello"}}
	// Simulate error in retrieving router's current version
	router3 := &models.Router{Model: models.Model{ID: 3}, CurrRouterVersionID: sql.NullInt32{Int32: 3, Valid: true}}

	sinceTime := time.Date(2020, 12, 5, 8, 0, 0, 0, time.UTC)
	tailLines := int64(5)
	headLines := int64(3)

	type args struct {
		r    *http.Request
		vars RequestVars
		body interface{}
	}
	tests := []struct {
		name                  string
		args                  args
		want                  *Response
		mlpService            func() service.MLPService
		routersService        func() service.RoutersService
		routerVersionsService func() service.RouterVersionsService
		podLogService         func() service.PodLogService
	}{
		{
			name: "missing project_id",
			args: args{
				vars: RequestVars{},
			},
			want: BadRequest("invalid project id", "key project_id not found in vars"),
			mlpService: func() service.MLPService {
				return nil
			},
			routersService: func() service.RoutersService {
				return nil
			},
			routerVersionsService: func() service.RouterVersionsService {
				return nil
			},
			podLogService: func() service.PodLogService {
				return nil
			},
		},
		{
			name: "missing router_id",
			args: args{
				vars: RequestVars{
					"project_id": {"1"},
				},
			},
			want: BadRequest("invalid router id", "key router_id not found in vars"),
			mlpService: func() service.MLPService {
				svc := &mocks.MLPService{}
				svc.On("GetProject", models.ID(1)).Return(project, nil)
				return svc
			},
			routersService: func() service.RoutersService {
				return nil
			},
			routerVersionsService: func() service.RouterVersionsService {
				return nil
			},
			podLogService: func() service.PodLogService {
				return nil
			},
		},
		{
			name: "expected args",
			args: args{
				vars: RequestVars{
					"project_id":     {"1"},
					"router_id":      {"1"},
					"component_type": {"router"},
				},
			},
			want: Ok([]*service.PodLog{{TextPayload: "routerVersion1"}}),
			mlpService: func() service.MLPService {
				svc := &mocks.MLPService{}
				svc.On("GetProject", models.ID(1)).Return(project, nil)
				return svc
			},
			routersService: func() service.RoutersService {
				svc := &mocks.RoutersService{}
				svc.On("FindByID", models.ID(1)).Return(router1, nil)
				return svc
			},
			routerVersionsService: func() service.RouterVersionsService {
				svc := &mocks.RouterVersionsService{}
				svc.On("FindByID", models.ID(1)).Return(routerVersion1, nil)
				return svc
			},
			podLogService: func() service.PodLogService {
				svc := &mocks.PodLogService{}
				svc.On("ListPodLogs", service.PodLogRequest{
					Namespace:        servicebuilder.GetNamespace(project),
					DefaultContainer: cluster.KnativeUserContainerName,
					Environment:      router1.EnvironmentName,
					LabelSelectors: []service.LabelSelector{
						{
							Key:   cluster.KnativeServiceLabelKey,
							Value: servicebuilder.GetComponentName(routerVersion1, "router"),
						},
					},
				}).Return([]*service.PodLog{{TextPayload: "routerVersion1"}}, nil)
				return svc
			},
		},
		{
			name: "invalid router version id",
			args: args{
				vars: RequestVars{
					"project_id":     {"1"},
					"router_id":      {"1"},
					"version":        {"3"},
					"component_type": {"router"},
				},
			},
			want: NotFound("router version not found", ""),
			mlpService: func() service.MLPService {
				svc := &mocks.MLPService{}
				svc.On("GetProject", models.ID(1)).Return(project, nil)
				return svc
			},
			routersService: func() service.RoutersService {
				svc := &mocks.RoutersService{}
				svc.On("FindByID", models.ID(1)).Return(router1, nil)
				return svc
			},
			routerVersionsService: func() service.RouterVersionsService {
				svc := &mocks.RouterVersionsService{}
				svc.On("FindByRouterIDAndVersion", models.ID(1), uint(3)).Return(nil, errors.New(""))
				return svc
			},
			podLogService: func() service.PodLogService {
				return nil
			},
		},
		{
			name: "invalid router version id reference in router",
			args: args{
				vars: RequestVars{
					"project_id":     {"1"},
					"router_id":      {"2"},
					"component_type": {"router"},
				},
			},
			want: BadRequest("Current router version id is invalid", "Make sure current router is deployed"),
			mlpService: func() service.MLPService {
				svc := &mocks.MLPService{}
				svc.On("GetProject", models.ID(1)).Return(project, nil)
				return svc
			},
			routersService: func() service.RoutersService {
				svc := &mocks.RoutersService{}
				svc.On("FindByID", models.ID(2)).Return(router2, nil)
				return svc
			},
			routerVersionsService: func() service.RouterVersionsService {
				svc := &mocks.RouterVersionsService{}
				svc.On("FindByID", models.ID(2)).Return(routerVersion2, nil)
				return svc
			},
			podLogService: func() service.PodLogService {
				return nil
			},
		},
		{
			name: "current version not found",
			args: args{
				vars: RequestVars{
					"project_id":     {"1"},
					"router_id":      {"3"},
					"component_type": {"router"},
				},
			},
			want: InternalServerError("Failed to find current router version", "test router version error"),
			mlpService: func() service.MLPService {
				svc := &mocks.MLPService{}
				svc.On("GetProject", models.ID(1)).Return(project, nil)
				return svc
			},
			routersService: func() service.RoutersService {
				svc := &mocks.RoutersService{}
				svc.On("FindByID", models.ID(3)).Return(router3, nil)
				return svc
			},
			routerVersionsService: func() service.RouterVersionsService {
				svc := &mocks.RouterVersionsService{}
				svc.On("FindByID", models.ID(3)).Return(nil, errors.New("test router version error"))
				return svc
			},
			podLogService: func() service.PodLogService {
				return nil
			},
		},
		{
			name: "valid optional args",
			args: args{
				vars: RequestVars{
					"project_id":     {"1"},
					"router_id":      {"1"},
					"version":        {"1"},
					"component_type": {"enricher"},
					"container":      {"mycontainer"},
					"previous":       {"true"},
					"since_time":     {"2020-12-05T08:00:00Z"},
					"tail_lines":     {"5"},
					"head_lines":     {"3"},
				},
			},
			want: Ok([]*service.PodLog{{TextPayload: "valid optional args"}}),
			mlpService: func() service.MLPService {
				svc := &mocks.MLPService{}
				svc.On("GetProject", models.ID(1)).Return(project, nil)
				return svc
			},
			routersService: func() service.RoutersService {
				svc := &mocks.RoutersService{}
				svc.On("FindByID", models.ID(1)).Return(router1, nil)
				return svc
			},
			routerVersionsService: func() service.RouterVersionsService {
				svc := &mocks.RouterVersionsService{}
				svc.On("FindByRouterIDAndVersion", models.ID(1), uint(1)).Return(routerVersion1, nil)
				return svc
			},
			podLogService: func() service.PodLogService {
				svc := &mocks.PodLogService{}
				svc.On("ListPodLogs", service.PodLogRequest{
					Namespace:        servicebuilder.GetNamespace(project),
					DefaultContainer: cluster.KnativeUserContainerName,
					Environment:      router1.EnvironmentName,
					LabelSelectors: []service.LabelSelector{
						{
							Key:   cluster.KnativeServiceLabelKey,
							Value: servicebuilder.GetComponentName(routerVersion1, "enricher"),
						},
					},
					SinceTime: &sinceTime,
					TailLines: &tailLines,
					HeadLines: &headLines,
					Previous:  true,
					Container: "mycontainer",
				}).Return([]*service.PodLog{{TextPayload: "valid optional args"}}, nil)
				return svc
			},
		},
		{
			name: "invalid component_type",
			args: args{
				vars: RequestVars{
					"project_id":     {"1"},
					"router_id":      {"1"},
					"version":        {"1"},
					"component_type": {"invalidcomponenttype"},
					"container":      {"mycontainer"},
					"previous":       {"true"},
					"since_time":     {"2020-12-05T08:00:00Z"},
					"tail_lines":     {"5"},
				},
			},
			want: BadRequest(
				"failed to fetch router pod logs",
				"failed to parse query string: Key: 'listRouterPodLogsOptions.ComponentType'"+
					" Error:Field validation for 'ComponentType' failed on the 'oneof' tag",
			),
			mlpService: func() service.MLPService {
				svc := &mocks.MLPService{}
				svc.On("GetProject", models.ID(1)).Return(project, nil)
				return svc
			},
			routersService: func() service.RoutersService {
				svc := &mocks.RoutersService{}
				svc.On("FindByID", models.ID(1)).Return(router1, nil)
				return svc
			},
			routerVersionsService: func() service.RouterVersionsService {
				return nil
			},
			podLogService: func() service.PodLogService {
				return nil
			},
		},
		{
			name: "list logs error",
			args: args{
				vars: RequestVars{
					"project_id":     {"1"},
					"router_id":      {"1"},
					"component_type": {"router"},
				},
			},
			want: InternalServerError("Failed to list logs", "test pod log error"),
			mlpService: func() service.MLPService {
				svc := &mocks.MLPService{}
				svc.On("GetProject", models.ID(1)).Return(project, nil)
				return svc
			},
			routersService: func() service.RoutersService {
				svc := &mocks.RoutersService{}
				svc.On("FindByID", models.ID(1)).Return(router1, nil)
				return svc
			},
			routerVersionsService: func() service.RouterVersionsService {
				svc := &mocks.RouterVersionsService{}
				svc.On("FindByID", models.ID(1)).Return(routerVersion1, nil)
				return svc
			},
			podLogService: func() service.PodLogService {
				svc := &mocks.PodLogService{}
				svc.On("ListPodLogs", service.PodLogRequest{
					Namespace:        servicebuilder.GetNamespace(project),
					DefaultContainer: cluster.KnativeUserContainerName,
					Environment:      router1.EnvironmentName,
					LabelSelectors: []service.LabelSelector{
						{
							Key:   cluster.KnativeServiceLabelKey,
							Value: servicebuilder.GetComponentName(routerVersion1, "router"),
						},
					},
				}).Return(nil, fmt.Errorf("test pod log error"))
				return svc
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, _ := validation.NewValidator(nil)
			c := PodLogController{
				NewBaseController(
					&AppContext{
						PodLogService:         tt.podLogService(),
						MLPService:            tt.mlpService(),
						RoutersService:        tt.routersService(),
						RouterVersionsService: tt.routerVersionsService(),
					},
					validator,
				),
			}
			if got := c.ListRouterPodLogs(tt.args.r, tt.args.vars, tt.args.body); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ListRouterPodLogs() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
