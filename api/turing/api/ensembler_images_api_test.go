package api

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/caraml-dev/mlp/api/client"
	"github.com/caraml-dev/turing/api/turing/api/request"
	"github.com/caraml-dev/turing/api/turing/imagebuilder"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/api/turing/service/mocks"
	"github.com/caraml-dev/turing/api/turing/validation"
	"github.com/stretchr/testify/mock"
)

func TestEnsemblerImagesController_ListImages(t *testing.T) {
	projectID := models.ID(1)

	type fields struct {
		BaseController BaseController
	}
	type args struct {
		in0  *http.Request
		vars RequestVars
		in2  interface{}
	}
	tests := []struct {
		name                  string
		mlpService            func() *mocks.MLPService
		ensemblerService      func() *mocks.EnsemblersService
		ensemblerImageService func() *mocks.EnsemblerImagesService
		args                  args
		want                  *Response
	}{
		{
			name: "success - ensembler job image",
			mlpService: func() *mocks.MLPService {
				s := &mocks.MLPService{}
				s.On("GetProject", models.ID(1)).Return(&client.Project{
					ID:   int32(projectID),
					Name: "myproject",
				}, nil)
				return s
			},
			ensemblerService: func() *mocks.EnsemblersService {
				s := &mocks.EnsemblersService{}
				s.On("FindByID", models.ID(1), service.EnsemblersFindByIDOptions{
					ProjectID: &projectID,
				}).Return(&models.PyFuncEnsembler{
					GenericEnsembler: &models.GenericEnsembler{
						Model: models.Model{
							ID: models.ID(1),
						},
						ProjectID: 1,
						Name:      "myensembler",
					},
					RunID: "abc123",
				}, nil)
				return s
			},
			ensemblerImageService: func() *mocks.EnsemblerImagesService {
				s := &mocks.EnsemblerImagesService{}
				s.On("ListImages", mock.Anything, mock.Anything, mock.Anything).
					Return(
						[]imagebuilder.EnsemblerImage{
							{
								ProjectID:           models.ID(1),
								EnsemblerID:         models.ID(1),
								EnsemblerRunnerType: models.EnsemblerRunnerTypeJob,
								ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-jobs/myproject/myensembler-1:abc123",
								Exists:              true,
								JobStatus: imagebuilder.JobStatus{
									State: imagebuilder.JobStateSucceeded,
								},
							},
						}, nil,
					)
				return s
			},
			args: args{
				in0: &http.Request{},
				vars: RequestVars{
					"project_id":   {"1"},
					"ensembler_id": {"1"},
					"runner_type":  {"job"},
				},
				in2: nil,
			},
			want: &Response{
				code: 200,
				data: []imagebuilder.EnsemblerImage{
					{
						ProjectID:           models.ID(1),
						EnsemblerID:         models.ID(1),
						EnsemblerRunnerType: models.EnsemblerRunnerTypeJob,
						ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-jobs/myproject/myensembler-1:abc123",
						Exists:              true,
						JobStatus: imagebuilder.JobStatus{
							State: imagebuilder.JobStateSucceeded,
						},
					},
				},
			},
		},
		{
			name: "success - ensembler service image",
			mlpService: func() *mocks.MLPService {
				s := &mocks.MLPService{}
				s.On("GetProject", models.ID(1)).Return(&client.Project{
					ID:   int32(projectID),
					Name: "myproject",
				}, nil)
				return s
			},
			ensemblerService: func() *mocks.EnsemblersService {
				s := &mocks.EnsemblersService{}
				s.On("FindByID", models.ID(1), service.EnsemblersFindByIDOptions{
					ProjectID: &projectID,
				}).Return(&models.PyFuncEnsembler{
					GenericEnsembler: &models.GenericEnsembler{
						Model: models.Model{
							ID: models.ID(1),
						},
						ProjectID: 1,
						Name:      "myensembler",
					},
					RunID: "abc123",
				}, nil)
				return s
			},
			ensemblerImageService: func() *mocks.EnsemblerImagesService {
				s := &mocks.EnsemblerImagesService{}
				s.On("ListImages", mock.Anything, mock.Anything, mock.Anything).
					Return(
						[]imagebuilder.EnsemblerImage{
							{
								ProjectID:           models.ID(1),
								EnsemblerID:         models.ID(1),
								EnsemblerRunnerType: models.EnsemblerRunnerTypeService,
								ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-services/myproject/myensembler-1:abc123",
								Exists:              true,
								JobStatus: imagebuilder.JobStatus{
									State: imagebuilder.JobStateSucceeded,
								},
							},
						}, nil,
					)
				return s
			},
			args: args{
				in0: &http.Request{},
				vars: RequestVars{
					"project_id":   {"1"},
					"ensembler_id": {"1"},
					"runner_type":  {"service"},
				},
				in2: nil,
			},
			want: &Response{
				code: 200,
				data: []imagebuilder.EnsemblerImage{
					{
						ProjectID:           models.ID(1),
						EnsemblerID:         models.ID(1),
						EnsemblerRunnerType: models.EnsemblerRunnerTypeService,
						ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-services/myproject/myensembler-1:abc123",
						Exists:              true,
						JobStatus: imagebuilder.JobStatus{
							State: imagebuilder.JobStateSucceeded,
						},
					},
				},
			},
		},
		{
			name: "success - both ensembler job & service image",
			mlpService: func() *mocks.MLPService {
				s := &mocks.MLPService{}
				s.On("GetProject", models.ID(1)).Return(&client.Project{
					ID:   int32(projectID),
					Name: "myproject",
				}, nil)
				return s
			},
			ensemblerService: func() *mocks.EnsemblersService {
				s := &mocks.EnsemblersService{}
				s.On("FindByID", models.ID(1), service.EnsemblersFindByIDOptions{
					ProjectID: &projectID,
				}).Return(&models.PyFuncEnsembler{
					GenericEnsembler: &models.GenericEnsembler{
						Model: models.Model{
							ID: models.ID(1),
						},
						ProjectID: 1,
						Name:      "myensembler",
					},
					RunID: "abc123",
				}, nil)
				return s
			},
			ensemblerImageService: func() *mocks.EnsemblerImagesService {
				s := &mocks.EnsemblerImagesService{}
				s.On("ListImages", mock.Anything, mock.Anything, mock.Anything).
					Return(
						[]imagebuilder.EnsemblerImage{
							{
								ProjectID:           models.ID(1),
								EnsemblerID:         models.ID(1),
								EnsemblerRunnerType: models.EnsemblerRunnerTypeJob,
								ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-jobs/myproject/myensembler-1:abc123",
								Exists:              true,
								JobStatus: imagebuilder.JobStatus{
									State: imagebuilder.JobStateSucceeded,
								},
							},
							{
								ProjectID:           models.ID(1),
								EnsemblerID:         models.ID(1),
								EnsemblerRunnerType: models.EnsemblerRunnerTypeService,
								ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-services/myproject/myensembler-1:abc123",
								Exists:              true,
								JobStatus: imagebuilder.JobStatus{
									State: imagebuilder.JobStateSucceeded,
								},
							},
						}, nil,
					)
				return s
			},
			args: args{
				in0: &http.Request{},
				vars: RequestVars{
					"project_id":   {"1"},
					"ensembler_id": {"1"},
					"runner_type":  {""},
				},
				in2: nil,
			},
			want: &Response{
				code: 200,
				data: []imagebuilder.EnsemblerImage{
					{
						ProjectID:           models.ID(1),
						EnsemblerID:         models.ID(1),
						EnsemblerRunnerType: models.EnsemblerRunnerTypeJob,
						ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-jobs/myproject/myensembler-1:abc123",
						Exists:              true,
						JobStatus: imagebuilder.JobStatus{
							State: imagebuilder.JobStateSucceeded,
						},
					},
					{
						ProjectID:           models.ID(1),
						EnsemblerID:         models.ID(1),
						EnsemblerRunnerType: models.EnsemblerRunnerTypeService,
						ImageRef:            "ghcr.io/caraml-dev/turing/ensembler-services/myproject/myensembler-1:abc123",
						Exists:              true,
						JobStatus: imagebuilder.JobStatus{
							State: imagebuilder.JobStateSucceeded,
						},
					},
				},
			},
		},
		{
			name: "failed - invalid query string",
			mlpService: func() *mocks.MLPService {
				s := &mocks.MLPService{}
				return s
			},
			ensemblerService: func() *mocks.EnsemblersService {
				s := &mocks.EnsemblersService{}
				return s
			},
			ensemblerImageService: func() *mocks.EnsemblerImagesService {
				s := &mocks.EnsemblerImagesService{}
				return s
			},
			args: args{
				in0: &http.Request{},
				vars: RequestVars{
					"X-project_id": {"1"},
					"ensembler_id": {"1"},
					"runner_type":  {""},
				},
				in2: nil,
			},
			want: &Response{
				code: 400,
				data: struct {
					Description string `json:"description"`
					Message     string `json:"error"`
				}{
					Description: "failed to list ensembler images",
					Message:     "failed to parse query string: Key: 'EnsemblerImagesListOptions.ProjectID' Error:Field validation for 'ProjectID' failed on the 'required' tag",
				},
			},
		},
		{
			name: "failed - timeout getting MLP project",
			mlpService: func() *mocks.MLPService {
				s := &mocks.MLPService{}
				s.On("GetProject", models.ID(1)).Return(nil, fmt.Errorf("timeout"))
				return s
			},
			ensemblerService: func() *mocks.EnsemblersService {
				s := &mocks.EnsemblersService{}
				return s
			},
			ensemblerImageService: func() *mocks.EnsemblerImagesService {
				s := &mocks.EnsemblerImagesService{}
				return s
			},
			args: args{
				in0: &http.Request{},
				vars: RequestVars{
					"project_id":   {"1"},
					"ensembler_id": {"1"},
					"runner_type":  {""},
				},
				in2: nil,
			},
			want: &Response{
				code: 500,
				data: struct {
					Description string `json:"description"`
					Message     string `json:"error"`
				}{
					Description: "unable to get MLP project for the router",
					Message:     "timeout",
				},
			},
		},
		{
			name: "failed - ensembler not found",
			mlpService: func() *mocks.MLPService {
				s := &mocks.MLPService{}
				s.On("GetProject", models.ID(1)).Return(&client.Project{
					ID:   int32(projectID),
					Name: "myproject",
				}, nil)
				return s
			},
			ensemblerService: func() *mocks.EnsemblersService {
				s := &mocks.EnsemblersService{}
				s.On("FindByID", models.ID(100), service.EnsemblersFindByIDOptions{
					ProjectID: &projectID,
				}).Return(nil, fmt.Errorf("not found"))
				return s
			},
			ensemblerImageService: func() *mocks.EnsemblerImagesService {
				s := &mocks.EnsemblerImagesService{}
				return s
			},
			args: args{
				in0: &http.Request{},
				vars: RequestVars{
					"project_id":   {"1"},
					"ensembler_id": {"100"},
					"runner_type":  {""},
				},
				in2: nil,
			},
			want: &Response{
				code: 404,
				data: struct {
					Description string `json:"description"`
					Message     string `json:"error"`
				}{
					Description: "ensembler not found",
					Message:     "not found",
				},
			},
		},
		{
			name: "failed - invalid ensembler object",
			mlpService: func() *mocks.MLPService {
				s := &mocks.MLPService{}
				s.On("GetProject", models.ID(1)).Return(&client.Project{
					ID:   int32(projectID),
					Name: "myproject",
				}, nil)
				return s
			},
			ensemblerService: func() *mocks.EnsemblersService {
				s := &mocks.EnsemblersService{}
				s.On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
					ProjectID: &projectID,
				}).Return(&models.GenericEnsembler{
					Model: models.Model{
						ID: models.ID(1),
					},
					ProjectID: 1,
					Name:      "myensembler",
				}, nil)
				return s
			},
			ensemblerImageService: func() *mocks.EnsemblerImagesService {
				s := &mocks.EnsemblerImagesService{}
				return s
			},
			args: args{
				in0: &http.Request{},
				vars: RequestVars{
					"project_id":   {"1"},
					"ensembler_id": {"2"},
					"runner_type":  {""},
				},
				in2: nil,
			},
			want: &Response{
				code: 500,
				data: struct {
					Description string `json:"description"`
					Message     string `json:"error"`
				}{
					Description: "unable to list ensembler images",
					Message:     "ensembler is not a PyFuncEnsembler",
				},
			},
		},
		{
			name: "failed - error listing images",
			mlpService: func() *mocks.MLPService {
				s := &mocks.MLPService{}
				s.On("GetProject", models.ID(1)).Return(&client.Project{
					ID:   int32(projectID),
					Name: "myproject",
				}, nil)
				return s
			},
			ensemblerService: func() *mocks.EnsemblersService {
				s := &mocks.EnsemblersService{}
				s.On("FindByID", models.ID(1), service.EnsemblersFindByIDOptions{
					ProjectID: &projectID,
				}).Return(&models.PyFuncEnsembler{
					GenericEnsembler: &models.GenericEnsembler{
						Model: models.Model{
							ID: models.ID(1),
						},
						ProjectID: 1,
						Name:      "myensembler",
					},
					RunID: "abc123",
				}, nil)
				return s
			},
			ensemblerImageService: func() *mocks.EnsemblerImagesService {
				s := &mocks.EnsemblerImagesService{}
				s.On("ListImages", mock.Anything, mock.Anything, mock.Anything).
					Return(
						nil, fmt.Errorf("timeout"),
					)
				return s
			},
			args: args{
				in0: &http.Request{},
				vars: RequestVars{
					"project_id":   {"1"},
					"ensembler_id": {"1"},
					"runner_type":  {"job"},
				},
				in2: nil,
			},
			want: &Response{
				code: 500,
				data: struct {
					Description string `json:"description"`
					Message     string `json:"error"`
				}{
					Description: "unable to list ensembler images",
					Message:     "timeout",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, _ := validation.NewValidator(nil)

			c := EnsemblerImagesController{
				BaseController: NewBaseController(
					&AppContext{
						MLPService:             tt.mlpService(),
						EnsemblersService:      tt.ensemblerService(),
						EnsemblerImagesService: tt.ensemblerImageService(),
					},
					validator,
				),
			}
			if got := c.ListImages(tt.args.in0, tt.args.vars, tt.args.in2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EnsemblerImagesController.ListImages() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnsemblerImagesController_BuildImage(t *testing.T) {
	projectID := models.ID(1)

	type args struct {
		in0  *http.Request
		vars RequestVars
		body interface{}
	}
	tests := []struct {
		name                  string
		mlpService            func() *mocks.MLPService
		ensemblerService      func() *mocks.EnsemblersService
		ensemblerImageService func() *mocks.EnsemblerImagesService
		args                  args
		want                  *Response
	}{
		{
			name: "success - build ensembler job image",
			mlpService: func() *mocks.MLPService {
				s := &mocks.MLPService{}
				s.On("GetProject", models.ID(1)).Return(&client.Project{
					ID:   int32(projectID),
					Name: "myproject",
				}, nil)
				return s
			},
			ensemblerService: func() *mocks.EnsemblersService {
				s := &mocks.EnsemblersService{}
				s.On("FindByID", models.ID(1), service.EnsemblersFindByIDOptions{
					ProjectID: &projectID,
				}).Return(&models.PyFuncEnsembler{
					GenericEnsembler: &models.GenericEnsembler{
						Model: models.Model{
							ID: models.ID(1),
						},
						ProjectID: 1,
						Name:      "myensembler",
					},
					RunID: "abc123",
				}, nil)
				return s
			},
			ensemblerImageService: func() *mocks.EnsemblerImagesService {
				s := &mocks.EnsemblerImagesService{}
				s.On("BuildImage", mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				return s
			},
			args: args{
				in0: &http.Request{},
				vars: RequestVars{
					"project_id":   {"1"},
					"ensembler_id": {"1"},
				},
				body: &request.BuildEnsemblerImageRequest{
					RunnerType: models.EnsemblerRunnerTypeJob,
				},
			},
			want: &Response{
				code: 202,
				data: nil,
			},
		},
		{
			name: "success - build ensembler service image",
			mlpService: func() *mocks.MLPService {
				s := &mocks.MLPService{}
				s.On("GetProject", models.ID(1)).Return(&client.Project{
					ID:   int32(projectID),
					Name: "myproject",
				}, nil)
				return s
			},
			ensemblerService: func() *mocks.EnsemblersService {
				s := &mocks.EnsemblersService{}
				s.On("FindByID", models.ID(1), service.EnsemblersFindByIDOptions{
					ProjectID: &projectID,
				}).Return(&models.PyFuncEnsembler{
					GenericEnsembler: &models.GenericEnsembler{
						Model: models.Model{
							ID: models.ID(1),
						},
						ProjectID: 1,
						Name:      "myensembler",
					},
					RunID: "abc123",
				}, nil)
				return s
			},
			ensemblerImageService: func() *mocks.EnsemblerImagesService {
				s := &mocks.EnsemblerImagesService{}
				s.On("BuildImage", mock.Anything, mock.Anything, mock.Anything).
					Return(nil)
				return s
			},
			args: args{
				in0: &http.Request{},
				vars: RequestVars{
					"project_id":   {"1"},
					"ensembler_id": {"1"},
				},
				body: &request.BuildEnsemblerImageRequest{
					RunnerType: models.EnsemblerRunnerTypeService,
				},
			},
			want: &Response{
				code: 202,
				data: nil,
			},
		},
		{
			name: "failed - invalid query string",
			mlpService: func() *mocks.MLPService {
				s := &mocks.MLPService{}
				return s
			},
			ensemblerService: func() *mocks.EnsemblersService {
				s := &mocks.EnsemblersService{}
				return s
			},
			ensemblerImageService: func() *mocks.EnsemblerImagesService {
				s := &mocks.EnsemblerImagesService{}
				return s
			},
			args: args{
				in0: &http.Request{},
				vars: RequestVars{
					"X-project_id": {"1"},
					"ensembler_id": {"1"},
				},
				body: &request.BuildEnsemblerImageRequest{
					RunnerType: models.EnsemblerRunnerTypeJob,
				},
			},
			want: &Response{
				code: 400,
				data: struct {
					Description string `json:"description"`
					Message     string `json:"error"`
				}{
					Description: "failed to build ensembler image",
					Message:     "failed to parse query string: Key: 'EnsemblerImagesPathOptions.ProjectID' Error:Field validation for 'ProjectID' failed on the 'required' tag",
				},
			},
		},
		{
			name: "failed - timeout getting MLP project",
			mlpService: func() *mocks.MLPService {
				s := &mocks.MLPService{}
				s.On("GetProject", models.ID(1)).Return(nil, fmt.Errorf("timeout"))
				return s
			},
			ensemblerService: func() *mocks.EnsemblersService {
				s := &mocks.EnsemblersService{}
				return s
			},
			ensemblerImageService: func() *mocks.EnsemblerImagesService {
				s := &mocks.EnsemblerImagesService{}
				return s
			},
			args: args{
				in0: &http.Request{},
				vars: RequestVars{
					"project_id":   {"1"},
					"ensembler_id": {"1"},
				},
				body: &request.BuildEnsemblerImageRequest{
					RunnerType: models.EnsemblerRunnerTypeJob,
				},
			},
			want: &Response{
				code: 500,
				data: struct {
					Description string `json:"description"`
					Message     string `json:"error"`
				}{
					Description: "unable to get MLP project for the router",
					Message:     "timeout",
				},
			},
		},
		{
			name: "failed - ensembler not found",
			mlpService: func() *mocks.MLPService {
				s := &mocks.MLPService{}
				s.On("GetProject", models.ID(1)).Return(&client.Project{
					ID:   int32(projectID),
					Name: "myproject",
				}, nil)
				return s
			},
			ensemblerService: func() *mocks.EnsemblersService {
				s := &mocks.EnsemblersService{}
				s.On("FindByID", models.ID(100), service.EnsemblersFindByIDOptions{
					ProjectID: &projectID,
				}).Return(nil, fmt.Errorf("not found"))
				return s
			},
			ensemblerImageService: func() *mocks.EnsemblerImagesService {
				s := &mocks.EnsemblerImagesService{}
				return s
			},
			args: args{
				in0: &http.Request{},
				vars: RequestVars{
					"project_id":   {"1"},
					"ensembler_id": {"100"},
				},
				body: &request.BuildEnsemblerImageRequest{
					RunnerType: models.EnsemblerRunnerTypeJob,
				},
			},
			want: &Response{
				code: 404,
				data: struct {
					Description string `json:"description"`
					Message     string `json:"error"`
				}{
					Description: "ensembler not found",
					Message:     "not found",
				},
			},
		},
		{
			name: "failed - invalid ensembler object",
			mlpService: func() *mocks.MLPService {
				s := &mocks.MLPService{}
				s.On("GetProject", models.ID(1)).Return(&client.Project{
					ID:   int32(projectID),
					Name: "myproject",
				}, nil)
				return s
			},
			ensemblerService: func() *mocks.EnsemblersService {
				s := &mocks.EnsemblersService{}
				s.On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
					ProjectID: &projectID,
				}).Return(&models.GenericEnsembler{
					Model: models.Model{
						ID: models.ID(1),
					},
					ProjectID: 1,
					Name:      "myensembler",
				}, nil)
				return s
			},
			ensemblerImageService: func() *mocks.EnsemblerImagesService {
				s := &mocks.EnsemblerImagesService{}
				return s
			},
			args: args{
				in0: &http.Request{},
				vars: RequestVars{
					"project_id":   {"1"},
					"ensembler_id": {"2"},
				},
				body: &request.BuildEnsemblerImageRequest{
					RunnerType: models.EnsemblerRunnerTypeJob,
				},
			},
			want: &Response{
				code: 500,
				data: struct {
					Description string `json:"description"`
					Message     string `json:"error"`
				}{
					Description: "unable to build ensembler image",
					Message:     "ensembler is not a PyFuncEnsembler",
				},
			},
		},
		{
			name: "failed - error building image",
			mlpService: func() *mocks.MLPService {
				s := &mocks.MLPService{}
				s.On("GetProject", models.ID(1)).Return(&client.Project{
					ID:   int32(projectID),
					Name: "myproject",
				}, nil)
				return s
			},
			ensemblerService: func() *mocks.EnsemblersService {
				s := &mocks.EnsemblersService{}
				s.On("FindByID", models.ID(2), service.EnsemblersFindByIDOptions{
					ProjectID: &projectID,
				}).Return(&models.PyFuncEnsembler{
					GenericEnsembler: &models.GenericEnsembler{
						Model: models.Model{
							ID: models.ID(1),
						},
						ProjectID: 1,
						Name:      "myensembler",
					},
					RunID: "abc123",
				}, nil)
				return s
			},
			ensemblerImageService: func() *mocks.EnsemblerImagesService {
				s := &mocks.EnsemblerImagesService{}
				s.On("BuildImage", mock.Anything, mock.Anything, mock.Anything).
					Return(fmt.Errorf("timeout"))

				return s
			},
			args: args{
				in0: &http.Request{},
				vars: RequestVars{
					"project_id":   {"1"},
					"ensembler_id": {"2"},
				},
				body: &request.BuildEnsemblerImageRequest{
					RunnerType: models.EnsemblerRunnerTypeJob,
				},
			},
			want: &Response{
				code: 202,
				data: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator, _ := validation.NewValidator(nil)

			c := EnsemblerImagesController{
				BaseController: NewBaseController(
					&AppContext{
						MLPService:             tt.mlpService(),
						EnsemblersService:      tt.ensemblerService(),
						EnsemblerImagesService: tt.ensemblerImageService(),
					},
					validator,
				),
			}
			if got := c.BuildImage(tt.args.in0, tt.args.vars, tt.args.body); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("EnsemblerImagesController.BuildImage() = %v, want %v", got, tt.want)
			}
		})
	}
}
