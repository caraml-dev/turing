//go:build integration

package service_test

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/api/turing/service/mocks"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/api/client"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/caraml-dev/turing/api/turing/database"
	"github.com/caraml-dev/turing/api/turing/models"
)

func TestRouterVersionsServiceIntegration(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		// Monitoring URL Deps
		monitoringURLFormat := "https://www.example.com/{{.ProjectName}}/{{.ClusterName}}/{{.RouterName}}/{{.Version}}"
		mlpService := &mocks.MLPService{}
		mlpService.On(
			"GetEnvironment",
			mock.Anything,
		).Return(&merlin.Environment{Cluster: "cluster-name"}, nil)
		mlpService.On(
			"GetProject",
			mock.Anything,
		).Return(&mlp.Project{Name: "project-name"}, nil)

		// create router first
		router := &models.Router{
			ProjectID:       1,
			EnvironmentName: "johto",
			Name:            "wooper",
			Status:          models.RouterStatusPending,
		}
		router, err := service.NewRoutersService(db, mlpService, &monitoringURLFormat).Save(router)
		assert.NoError(t, err)

		// populate database
		svc := service.NewRouterVersionsService(db, mlpService, &monitoringURLFormat)
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
			AutoscalingPolicy: &models.AutoscalingPolicy{
				Metric: "concurrency",
				Target: "1",
			},
			Timeout:  "5s",
			Protocol: "HTTP_JSON",
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
				AutoscalingPolicy: &models.AutoscalingPolicy{
					Metric: "concurrency",
					Target: "1",
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
					AutoscalingPolicy: &models.AutoscalingPolicy{
						Metric: "concurrency",
						Target: "1",
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
		router.MonitoringURL = ""
		routerVersion.Router = router
		routerVersion.Router.CreatedAt = found.Router.CreatedAt
		routerVersion.Router.UpdatedAt = found.Router.UpdatedAt
		routerVersion.Enricher.CreatedAt = found.Enricher.CreatedAt
		routerVersion.Enricher.UpdatedAt = found.Enricher.UpdatedAt
		routerVersion.Ensembler.CreatedAt = found.Ensembler.CreatedAt
		routerVersion.Ensembler.UpdatedAt = found.Ensembler.UpdatedAt
		routerVersion.MonitoringURL = "https://www.example.com/project-name/cluster-name/wooper/1"
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
		found, err = svc.FindByRouterIDAndVersion(router.ID, 1)
		assert.NoError(t, err)
		assert.Equal(t, models.ID(1), found.ID)

		// list with versions
		list, err = svc.ListRouterVersionsWithStatus(router.ID, models.RouterVersionStatusDeployed)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(list))
		assert.Equal(t, models.RouterVersionStatusDeployed, list[0].Status)

		// delete with upstream reference
		router.SetCurrRouterVersion(routerVersion)
		_, err = service.NewRoutersService(db, mlpService, &monitoringURLFormat).Save(router)
		assert.NoError(t, err)
		err = svc.Delete(routerVersion)
		assert.Error(t, err, "unable to delete router version - there exists a router that is currently using this version")

		// clear reference and delete
		router.CurrRouterVersionID = sql.NullInt32{}
		router.CurrRouterVersion = nil
		_, err = service.NewRoutersService(db, mlpService, &monitoringURLFormat).Save(router)
		assert.NoError(t, err)
		err = svc.Delete(routerVersion)
		found, err = svc.FindByID(1)
		assert.Error(t, err)
		assert.Nil(t, found)

		var count int64 = -1
		db.Model(&models.Ensembler{}).Count(&count)
		assert.Equal(t, int64(0), count)
		// reset count
		count = -1
		db.Model(&models.Enricher{}).Count(&count)
		assert.Equal(t, int64(0), count)

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
