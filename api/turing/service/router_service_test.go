// +build integration

package service

import (
	"database/sql"
	"testing"

	"github.com/gojek/turing/api/turing/it/database"
	"github.com/gojek/turing/api/turing/models"
	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/assert"
)

func TestRoutersServiceIntegration(t *testing.T) {
	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
		svc := NewRoutersService(db)

		// create routers
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
			{
				ProjectID:       2,
				EnvironmentName: "env",
				Name:            "pizza",
				Status:          models.RouterStatusDeployed,
			},
		}
		for i, router := range routers {
			router, err := svc.Save(router)
			assert.NoError(t, err)
			assert.Equal(t, uint(i+1), router.ID)
			assert.NotNil(t, router.CreatedAt)
			assert.NotNil(t, router.UpdatedAt)
		}

		router := routers[0]
		// update router with routerversions
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
			Timeout:         "5s",
			LogConfig:       &models.LogConfig{},
			Enricher: &models.Enricher{
				Image:           "lalal",
				ResourceRequest: &models.ResourceRequest{},
				Endpoint:        "fsaf",
				Timeout:         "5s",
				Port:            8080,
				Env:             []*models.EnvVar{},
			},
			Ensembler: &models.Ensembler{
				Type: "docker",
				DockerConfig: &models.EnsemblerDockerConfig{
					Image:           "lalalala",
					ResourceRequest: &models.ResourceRequest{},
					Endpoint:        "fsaf",
					Timeout:         "5s",
					Port:            8080,
					Env:             []*models.EnvVar{},
				},
			},
		}
		routerVersion, err := NewRouterVersionsService(db).Save(routerVersion)
		assert.NoError(t, err)
		router.SetCurrRouterVersionID(routerVersion.ID)
		saved, err := svc.Save(router)
		assert.NoError(t, err)
		assert.Equal(t, int32(routerVersion.ID), saved.CurrRouterVersionID.Int32)

		// get router by name
		found, err := svc.FindByProjectAndEnvironmentAndName(router.ProjectID, router.EnvironmentName, router.Name)
		router.CurrRouterVersion = routerVersion
		assert.NoError(t, err)
		assert.Equal(t, found.ID, router.ID)
		assert.NotNil(t, found.CurrRouterVersion)

		// find by id
		found, err = svc.FindByID(routers[2].ID)
		assert.NoError(t, err)
		routers[2].Model = found.Model
		assert.Equal(t, found, routers[2])

		// list routers
		list, err := svc.ListRouters(2, "env")
		assert.NoError(t, err)
		assert.ElementsMatch(t, []*models.Router{found}, list)

		// dissociate curr_version and get
		router.CurrRouterVersionID = sql.NullInt32{Valid: false}
		router.CurrRouterVersion = nil
		router, err = svc.Save(router)
		assert.NoError(t, err)
		assert.Nil(t, router.CurrRouterVersion)

		// delete
		err = svc.Delete(router)
		assert.NoError(t, err)
		found, err = svc.FindByID(router.ID)
		assert.Error(t, err)
		assert.Nil(t, found)
		var count int
		db.Select("router_versions.*").Count(&count)
		assert.Equal(t, count, 0)
		db.Select("ensemblers.*").Count(&count)
		assert.Equal(t, count, 0)
		db.Select("enrichers.*").Count(&count)
		assert.Equal(t, count, 0)
	})
}
