// +build integration

package service

import (
	"testing"
	"text/template"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/api/client"
	"github.com/gojek/turing/api/turing/models"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestGenerateMonitoringURL(t *testing.T) {
	monitoringURLFormat := "https://www.example.com/{{.ProjectName}}/{{.ClusterName}}/{{.RouterName}}/{{.Version}}"
	var routerVersion uint = 10
	tests := map[string]struct {
		format          *string
		mlpService      func() MLPService
		environmentName string
		projectID       models.ID
		routerName      string
		routerVersion   *uint
		expected        string
	}{
		"success | nominal": {
			format: &monitoringURLFormat,
			mlpService: func() MLPService {
				mlpService := &MockMLPService{}
				mlpService.On(
					"GetEnvironment",
					mock.Anything,
				).Return(&merlin.Environment{Cluster: "cluster-name"}, nil)
				mlpService.On(
					"GetProject",
					mock.Anything,
				).Return(&mlp.Project{Name: "project-name"}, nil)
				return mlpService
			},
			environmentName: "environment",
			projectID:       models.ID(1),
			routerName:      "router-name",
			routerVersion:   &routerVersion,
			expected:        "https://www.example.com/project-name/cluster-name/router-name/10",
		},
		"success | no router version provided": {
			format: &monitoringURLFormat,
			mlpService: func() MLPService {
				mlpService := &MockMLPService{}
				mlpService.On(
					"GetEnvironment",
					mock.Anything,
				).Return(&merlin.Environment{Cluster: "cluster-name"}, nil)
				mlpService.On(
					"GetProject",
					mock.Anything,
				).Return(&mlp.Project{Name: "project-name"}, nil)
				return mlpService
			},
			environmentName: "environment",
			projectID:       models.ID(1),
			routerName:      "router-name",
			routerVersion:   nil,
			expected:        "https://www.example.com/project-name/cluster-name/router-name/$__all",
		},
		"success | no format given": {
			format: nil,
			mlpService: func() MLPService {
				mlpService := &MockMLPService{}
				mlpService.On(
					"GetEnvironment",
					mock.Anything,
				).Return(&merlin.Environment{Cluster: "cluster-name"}, nil)
				mlpService.On(
					"GetProject",
					mock.Anything,
				).Return(&mlp.Project{Name: "project-name"}, nil)
				return mlpService
			},
			environmentName: "environment",
			projectID:       models.ID(1),
			routerName:      "router-name",
			routerVersion:   &routerVersion,
			expected:        "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			var temp *template.Template
			if tt.format != nil {
				var err error
				temp, err = template.New("monitoringURLTemplate").Parse(*tt.format)
				assert.Nil(t, err)
			}
			svc := routerMonitoringService{tt.mlpService(), temp}
			result, err := svc.GenerateMonitoringURL(tt.projectID, tt.environmentName, tt.routerName, tt.routerVersion)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
