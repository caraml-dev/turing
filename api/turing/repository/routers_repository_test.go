//go:build integration

package repository_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"

	"github.com/caraml-dev/turing/api/turing/database"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/repository"
)

func TestRoutersServiceIntegration(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		rRepo := repository.NewRoutersRepository(db)
		rvRepo := repository.NewRouterVersionsRepository(db)

		// Create routers
		routers := []*models.Router{
			{
				ProjectID:       1,
				EnvironmentName: "env",
				Name:            "hamburger",
				Status:          models.RouterStatusPending,
			},
			{
				ProjectID:       1,
				EnvironmentName: "env",
				Name:            "pasta",
				Status:          models.RouterStatusFailed,
			},
		}
		for i, router := range routers {
			router, err := rRepo.Save(router)
			assert.NoError(t, err)
			assert.Equal(t, models.ID(i+1), router.ID)
			assert.NotNil(t, router.CreatedAt)
			assert.NotNil(t, router.UpdatedAt)
		}

		// Create router version
		router := routers[0]
		routerVersion := &models.RouterVersion{
			RouterID: router.ID,
			Status:   models.RouterVersionStatusPending,
			Image:    "asia.gcr.io/myimage:1.0.0",
			Routes: []*models.Route{
				{
					ID:       "lala",
					Type:     "PROXY",
					Endpoint: "lalal.com",
					Timeout:  "5s",
				},
			},
			DefaultRouteID: "lala",
			ExperimentEngine: &models.ExperimentEngine{
				Type: models.ExperimentEngineTypeNop,
			},
			ResourceRequest: &models.ResourceRequest{},
			AutoscalingPolicy: &models.AutoscalingPolicy{
				Metric: models.AutoscalingMetricConcurrency,
				Target: "1",
			},
			Timeout:   "5s",
			Protocol:  "HTTP_JSON",
			LogConfig: &models.LogConfig{},
			Ensembler: &models.Ensembler{
				Type: "nop",
			},
		}
		routerVersion, err := rvRepo.Save(routerVersion)
		assert.NoError(t, err)
		assert.Equal(t, models.ID(1), routerVersion.ID)

		// Set current router version
		router.SetCurrRouterVersionID(models.ID(1))
		saved, err := rRepo.Save(router)
		assert.NoError(t, err)
		assert.Equal(t, 1, int(saved.CurrRouterVersionID.Int32))

		// Get router by current version ID
		assert.Equal(t, 1, int(rRepo.CountRoutersByCurrentVersionID(models.ID(1))))
	})
}
