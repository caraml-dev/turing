//go:build integration

package repository_test

import (
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/caraml-dev/turing/api/turing/database"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/repository"
)

func TestRouterVersionsServiceIntegration(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		rRepo := repository.NewRoutersRepository(db)
		rvRepo := repository.NewRouterVersionsRepository(db)

		// Create router first
		router := &models.Router{
			ProjectID:       1,
			EnvironmentName: "johto",
			Name:            "wooper",
			Status:          models.RouterStatusPending,
		}
		router, err := rRepo.Save(router)
		assert.NoError(t, err)

		// Populate database
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
		saved, err := rvRepo.Save(routerVersion)
		assert.NoError(t, err)
		assert.Equal(t, models.ID(1), saved.ID)
		assert.NotNil(t, saved.CreatedAt)
		assert.NotNil(t, saved.UpdatedAt)
		assert.Equal(t, 1, int(saved.EnsemblerID.Int32))
		assert.Equal(t, 1, int(saved.EnricherID.Int32))

		// Find by id
		found, err := rvRepo.FindByID(1)
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
		if !cmp.Equal(routerVersion, found) {
			routerVersionJson, _ := json.Marshal(routerVersion)
			foundJson, _ := json.Marshal(found)
			t.Errorf("Not equal: \n expected: %s\n actual: %s\n", routerVersionJson, foundJson)
		}

		// Find by version
		found, err = rvRepo.FindByRouterIDAndVersion(router.ID, 1)
		assert.NoError(t, err)
		if !cmp.Equal(routerVersion, found) {
			routerVersionJson, _ := json.Marshal(routerVersion)
			foundJson, _ := json.Marshal(found)
			t.Errorf("Not equal: \n expected: %s\n actual: %s\n", routerVersionJson, foundJson)
		}

		// List
		list, err := rvRepo.List(router.ID)
		assert.NoError(t, err)
		assert.ElementsMatch(t, []*models.RouterVersion{found}, list)

		// Update
		found.Status = models.RouterVersionStatusDeployed
		saved, err = rvRepo.Save(found)
		assert.NoError(t, err)
		assert.Equal(t, models.ID(1), saved.ID)
		assert.Equal(t, models.RouterVersionStatusDeployed, saved.Status)
		found, err = rvRepo.FindByRouterIDAndVersion(router.ID, 1)
		assert.NoError(t, err)
		assert.Equal(t, models.ID(1), found.ID)

		// List with versions
		list, err = rvRepo.ListByStatus(router.ID, models.RouterVersionStatusDeployed)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(list))
		assert.Equal(t, models.RouterVersionStatusDeployed, list[0].Status)

		// Clear reference and delete
		router.CurrRouterVersionID = sql.NullInt32{}
		router.CurrRouterVersion = nil
		_, err = rRepo.Save(router)
		assert.NoError(t, err)
		err = rvRepo.Delete(routerVersion)
		found, err = rvRepo.FindByID(1)
		t.Log(found, err)
		assert.Error(t, err)

		var count int64 = -1
		db.Model(&models.Ensembler{}).Count(&count)
		assert.Equal(t, int64(0), count)
		// Reset count
		count = -1
		db.Model(&models.Enricher{}).Count(&count)
		assert.Equal(t, int64(0), count)

		// Create router again without ensembler and enricher
		routerVersion.EnricherID = sql.NullInt32{}
		routerVersion.Enricher = nil
		routerVersion.EnsemblerID = sql.NullInt32{}
		routerVersion.Ensembler = nil
		routerVersion, err = rvRepo.Save(routerVersion)
		assert.NoError(t, err)
		assert.Nil(t, routerVersion.Enricher)
		assert.Nil(t, routerVersion.Ensembler)
	})
}
