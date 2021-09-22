// +build integration

package service

import (
	"database/sql"
	"encoding/json"
	"testing"

	merlin "github.com/gojek/merlin/client"
	"github.com/gojek/turing/api/turing/it/database"
	"github.com/gojek/turing/api/turing/models"
	"github.com/google/go-cmp/cmp"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestRouterVersionsServiceIntegration(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		// create router first
		router := &models.Router{
			ProjectID:       1,
			EnvironmentName: "johto",
			Name:            "wooper",
			Status:          models.RouterStatusPending,
		}
		router, err := NewRoutersService(db).Save(router)
		assert.NoError(t, err)

		// populate database
		svc := NewRouterVersionsService(db, nil, nil)
		routerVersion := &models.RouterVersion{
			RouterID: router.ID,
			Status:   models.RouterVersionStatusPending,
			Image:    "asia.gcr.io/myimage:1.0.0",
			Routes: []*models.Route{
				{
					ID:       "bulbasaur",
					Type:     "PROXY",
					Endpoint: "bulbasaur.default.svc.cluster.local:80",
					Timeout:  "5s",
				},
			},
			DefaultRouteID: "bulbasaur",
			ExperimentEngine: &models.ExperimentEngine{
				Type: models.ExperimentEngineTypeNop,
			},
			ResourceRequest: &models.ResourceRequest{
				MinReplica: 0,
				MaxReplica: 10,
				CPURequest: resource.Quantity{
					Format: "100m",
				},
				MemoryRequest: resource.Quantity{
					Format: "1G",
				},
			},
			Timeout: "5s",
			LogConfig: &models.LogConfig{
				LogLevel:             "DEBUG",
				CustomMetricsEnabled: true,
				FiberDebugLogEnabled: true,
			},
			Enricher: &models.Enricher{
				Image: "enricher:1.0.0",
				ResourceRequest: &models.ResourceRequest{
					MinReplica: 0,
					MaxReplica: 10,
					CPURequest: resource.Quantity{
						Format: "100m",
					},
					MemoryRequest: resource.Quantity{
						Format: "1G",
					},
				},
				Endpoint: "/enrich",
				Timeout:  "5s",
				Port:     8080,
				Env: []*models.EnvVar{
					{
						Name:  "KEY",
						Value: "VALUE",
					},
				},
			},
			Ensembler: &models.Ensembler{
				Type: "docker",
				DockerConfig: &models.EnsemblerDockerConfig{
					Image: "ensembler:1.0.0",
					ResourceRequest: &models.ResourceRequest{
						MinReplica: 0,
						MaxReplica: 10,
						CPURequest: resource.Quantity{
							Format: "100m",
						},
						MemoryRequest: resource.Quantity{
							Format: "1G",
						},
					},
					Endpoint: "/ensemble",
					Timeout:  "5s",
					Port:     8080,
					Env: []*models.EnvVar{
						{
							Name:  "KEY",
							Value: "VALUE",
						},
					},
				},
			},
		}
		saved, err := svc.Save(routerVersion)
		assert.NoError(t, err)
		assert.Equal(t, models.ID(1), saved.ID)
		assert.NotNil(t, saved.CreatedAt)
		assert.NotNil(t, saved.UpdatedAt)
		assert.Equal(t, 1, int(saved.EnsemblerID.Int32))
		assert.Equal(t, 1, int(saved.EnricherID.Int32))

		// find by id
		found, err := svc.FindByID(1)
		assert.NoError(t, err)
		assert.Equal(t, uint(1), found.Version)
		routerVersion.Version = found.Version
		routerVersion.CreatedAt = found.CreatedAt
		routerVersion.UpdatedAt = found.UpdatedAt
		routerVersion.Router = router
		routerVersion.Router.CreatedAt = found.Router.CreatedAt
		routerVersion.Router.UpdatedAt = found.Router.UpdatedAt
		routerVersion.Enricher.CreatedAt = found.Enricher.CreatedAt
		routerVersion.Enricher.UpdatedAt = found.Enricher.UpdatedAt
		routerVersion.Ensembler.CreatedAt = found.Ensembler.CreatedAt
		routerVersion.Ensembler.UpdatedAt = found.Ensembler.UpdatedAt
		if !cmp.Equal(routerVersion, found) {
			routerVersionJson, _ := json.Marshal(routerVersion)
			foundJson, _ := json.Marshal(found)
			t.Errorf("Not equal: \n expected: %s\n actual: %s\n", routerVersionJson, foundJson)
		}

		// find by version
		found, err = svc.FindByRouterIDAndVersion(router.ID, 1)
		assert.NoError(t, err)
		if !cmp.Equal(routerVersion, found) {
			routerVersionJson, _ := json.Marshal(routerVersion)
			foundJson, _ := json.Marshal(found)
			t.Errorf("Not equal: \n expected: %s\n actual: %s\n", routerVersionJson, foundJson)
		}

		// list
		list, err := svc.ListRouterVersions(router.ID)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []*models.RouterVersion{found}, list)

		// update
		found.Status = models.RouterVersionStatusDeployed
		saved, err = svc.Save(found)
		assert.NoError(t, err)
		assert.Equal(t, models.ID(1), saved.ID)
		assert.Equal(t, models.RouterVersionStatusDeployed, saved.Status)

		// list with versions
		list, err = svc.ListRouterVersionsWithStatus(router.ID, models.RouterVersionStatusDeployed)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(list))
		assert.Equal(t, models.RouterVersionStatusDeployed, list[0].Status)

		// delete with upstream reference
		router.SetCurrRouterVersion(routerVersion)
		_, err = NewRoutersService(db).Save(router)
		assert.NoError(t, err)
		err = svc.Delete(routerVersion)
		assert.Error(t, err, "unable to delete router version - there exists a router that is currently using this version")

		// clear reference and delete
		router.CurrRouterVersionID = sql.NullInt32{}
		router.CurrRouterVersion = nil
		_, err = NewRoutersService(db).Save(router)
		assert.NoError(t, err)
		err = svc.Delete(routerVersion)
		found, err = svc.FindByID(1)
		assert.Error(t, err)
		assert.Nil(t, found)

		count := -1
		db.Model(&models.Ensembler{}).Count(&count)
		assert.Equal(t, 0, count)
		// reset count
		count = -1
		db.Model(&models.Enricher{}).Count(&count)
		assert.Equal(t, 0, count)

		// create router again without ensembler and enricher
		routerVersion.EnricherID = sql.NullInt32{}
		routerVersion.Enricher = nil
		routerVersion.EnsemblerID = sql.NullInt32{}
		routerVersion.Ensembler = nil
		routerVersion, err = svc.Save(routerVersion)
		assert.NoError(t, err)
		assert.Nil(t, routerVersion.Enricher)
		assert.Nil(t, routerVersion.Ensembler)
	})
}

func TestGenerateMonitoringURL(t *testing.T) {
	monitoringURLFormat := "https://www.example.com/{{.ProjectName}}/{{.ClusterName}}/{{.RouterName}}/{{.Version}}"
	var routerVersion uint = 10
	tests := map[string]struct {
		format          *string
		mlpService      func() MLPService
		environmentName string
		projectName     string
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
				return mlpService
			},
			environmentName: "environment",
			projectName:     "project-name",
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
				return mlpService
			},
			environmentName: "environment",
			projectName:     "project-name",
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
				return mlpService
			},
			environmentName: "environment",
			projectName:     "project-name",
			routerName:      "router-name",
			routerVersion:   &routerVersion,
			expected:        "",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			svc := NewRouterVersionsService(nil, tt.mlpService(), tt.format)
			result, err := svc.GenerateMonitoringURL(tt.projectName, tt.environmentName, tt.routerName, tt.routerVersion)
			assert.Nil(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
