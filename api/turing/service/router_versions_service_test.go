package service_test

import (
	// 	"database/sql"
	// 	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/api/turing/service/mock_service"
	// 	merlin "github.com/gojek/merlin/client"
	// 	mlp "github.com/gojek/mlp/api/client"
	// 	"github.com/google/go-cmp/cmp"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	// 	"gorm.io/gorm"
	// 	"k8s.io/apimachinery/pkg/api/resource"
	// 	"github.com/caraml-dev/turing/api/turing/database"
	"github.com/caraml-dev/turing/api/turing/models"
)

// func TestRouterVersionsServiceIntegration(t *testing.T) {
// 	database.WithTestDatabase(t, func(t *testing.T, db *gorm.DB) {
// 		// Monitoring URL Deps
// 		monitoringURLFormat := "https://www.example.com/{{.ProjectName}}/{{.ClusterName}}/{{.RouterName}}/{{.Version}}"
// 		mlpService := &mocks.MLPService{}
// 		mlpService.On(
// 			"GetEnvironment",
// 			mock.Anything,
// 		).Return(&merlin.Environment{Cluster: "cluster-name"}, nil)
// 		mlpService.On(
// 			"GetProject",
// 			mock.Anything,
// 		).Return(&mlp.Project{Name: "project-name"}, nil)

// 		// create router first
// 		router := &models.Router{
// 			ProjectID:       1,
// 			EnvironmentName: "johto",
// 			Name:            "wooper",
// 			Status:          models.RouterStatusPending,
// 		}
// 		router, err := service.NewRoutersService(db, mlpService, &monitoringURLFormat).Save(router)
// 		assert.NoError(t, err)

// 		// populate database
// 		svc := service.NewRouterVersionsService(db, mlpService, &monitoringURLFormat)
// 		routerVersion := &models.RouterVersion{
// 			RouterID: router.ID,
// 			Status:   models.RouterVersionStatusPending,
// 			Image:    "asia.gcr.io/myimage:1.0.0",
// 			Routes: []*models.Route{
// 				{
// 					ID:       "bulbasaur",
// 					Type:     "PROXY",
// 					Endpoint: "bulbasaur.default.svc.cluster.local:80",
// 					Timeout:  "5s",
// 				},
// 			},
// 			DefaultRouteID: "bulbasaur",
// 			ExperimentEngine: &models.ExperimentEngine{
// 				Type: models.ExperimentEngineTypeNop,
// 			},
// 			ResourceRequest: &models.ResourceRequest{
// 				MinReplica: 0,
// 				MaxReplica: 10,
// 				CPURequest: resource.Quantity{
// 					Format: "100m",
// 				},
// 				MemoryRequest: resource.Quantity{
// 					Format: "1G",
// 				},
// 			},
// 			AutoscalingPolicy: &models.AutoscalingPolicy{
// 				Metric: "concurrency",
// 				Target: "1",
// 			},
// 			Timeout:  "5s",
// 			Protocol: "HTTP_JSON",
// 			LogConfig: &models.LogConfig{
// 				LogLevel:             "DEBUG",
// 				CustomMetricsEnabled: true,
// 				FiberDebugLogEnabled: true,
// 			},
// 			Enricher: &models.Enricher{
// 				Image: "enricher:1.0.0",
// 				ResourceRequest: &models.ResourceRequest{
// 					MinReplica: 0,
// 					MaxReplica: 10,
// 					CPURequest: resource.Quantity{
// 						Format: "100m",
// 					},
// 					MemoryRequest: resource.Quantity{
// 						Format: "1G",
// 					},
// 				},
// 				AutoscalingPolicy: &models.AutoscalingPolicy{
// 					Metric: "concurrency",
// 					Target: "1",
// 				},
// 				Endpoint: "/enrich",
// 				Timeout:  "5s",
// 				Port:     8080,
// 				Env: []*models.EnvVar{
// 					{
// 						Name:  "KEY",
// 						Value: "VALUE",
// 					},
// 				},
// 			},
// 			Ensembler: &models.Ensembler{
// 				Type: "docker",
// 				DockerConfig: &models.EnsemblerDockerConfig{
// 					Image: "ensembler:1.0.0",
// 					ResourceRequest: &models.ResourceRequest{
// 						MinReplica: 0,
// 						MaxReplica: 10,
// 						CPURequest: resource.Quantity{
// 							Format: "100m",
// 						},
// 						MemoryRequest: resource.Quantity{
// 							Format: "1G",
// 						},
// 					},
// 					AutoscalingPolicy: &models.AutoscalingPolicy{
// 						Metric: "concurrency",
// 						Target: "1",
// 					},
// 					Endpoint: "/ensemble",
// 					Timeout:  "5s",
// 					Port:     8080,
// 					Env: []*models.EnvVar{
// 						{
// 							Name:  "KEY",
// 							Value: "VALUE",
// 						},
// 					},
// 				},
// 			},
// 		}
// 		saved, err := svc.Save(routerVersion)
// 		assert.NoError(t, err)
// 		assert.Equal(t, models.ID(1), saved.ID)
// 		assert.NotNil(t, saved.CreatedAt)
// 		assert.NotNil(t, saved.UpdatedAt)
// 		assert.Equal(t, 1, int(saved.EnsemblerID.Int32))
// 		assert.Equal(t, 1, int(saved.EnricherID.Int32))

// 		// find by id
// 		found, err := svc.FindByID(1)
// 		assert.NoError(t, err)
// 		assert.Equal(t, uint(1), found.Version)
// 		routerVersion.Version = found.Version
// 		routerVersion.CreatedAt = found.CreatedAt
// 		routerVersion.UpdatedAt = found.UpdatedAt
// 		router.MonitoringURL = ""
// 		routerVersion.Router = router
// 		routerVersion.Router.CreatedAt = found.Router.CreatedAt
// 		routerVersion.Router.UpdatedAt = found.Router.UpdatedAt
// 		routerVersion.Enricher.CreatedAt = found.Enricher.CreatedAt
// 		routerVersion.Enricher.UpdatedAt = found.Enricher.UpdatedAt
// 		routerVersion.Ensembler.CreatedAt = found.Ensembler.CreatedAt
// 		routerVersion.Ensembler.UpdatedAt = found.Ensembler.UpdatedAt
// 		routerVersion.MonitoringURL = "https://www.example.com/project-name/cluster-name/wooper/1"
// 		if !cmp.Equal(routerVersion, found) {
// 			routerVersionJson, _ := json.Marshal(routerVersion)
// 			foundJson, _ := json.Marshal(found)
// 			t.Errorf("Not equal: \n expected: %s\n actual: %s\n", routerVersionJson, foundJson)
// 		}

// 		// find by version
// 		found, err = svc.FindByRouterIDAndVersion(router.ID, 1)
// 		assert.NoError(t, err)
// 		if !cmp.Equal(routerVersion, found) {
// 			routerVersionJson, _ := json.Marshal(routerVersion)
// 			foundJson, _ := json.Marshal(found)
// 			t.Errorf("Not equal: \n expected: %s\n actual: %s\n", routerVersionJson, foundJson)
// 		}

// 		// list
// 		list, err := svc.ListRouterVersions(router.ID)
// 		assert.NoError(t, err)
// 		assert.ElementsMatch(t, []*models.RouterVersion{found}, list)

// 		// update
// 		found.Status = models.RouterVersionStatusDeployed
// 		saved, err = svc.Save(found)
// 		assert.NoError(t, err)
// 		assert.Equal(t, models.ID(1), saved.ID)
// 		assert.Equal(t, models.RouterVersionStatusDeployed, saved.Status)
// 		found, err = svc.FindByRouterIDAndVersion(router.ID, 1)
// 		assert.NoError(t, err)
// 		assert.Equal(t, models.ID(1), found.ID)

// 		// list with versions
// 		list, err = svc.ListRouterVersionsWithStatus(router.ID, models.RouterVersionStatusDeployed)
// 		assert.NoError(t, err)
// 		assert.Equal(t, 1, len(list))
// 		assert.Equal(t, models.RouterVersionStatusDeployed, list[0].Status)

// 		// delete with upstream reference
// 		router.SetCurrRouterVersion(routerVersion)
// 		_, err = service.NewRoutersService(db, mlpService, &monitoringURLFormat).Save(router)
// 		assert.NoError(t, err)
// 		err = svc.Delete(routerVersion)
// 		assert.Error(t, err, "unable to delete router version - there exists a router that is currently using this version")

// 		// clear reference and delete
// 		router.CurrRouterVersionID = sql.NullInt32{}
// 		router.CurrRouterVersion = nil
// 		_, err = service.NewRoutersService(db, mlpService, &monitoringURLFormat).Save(router)
// 		assert.NoError(t, err)
// 		err = svc.Delete(routerVersion)
// 		found, err = svc.FindByID(1)
// 		assert.Error(t, err)
// 		assert.Nil(t, found)

// 		var count int64 = -1
// 		db.Model(&models.Ensembler{}).Count(&count)
// 		assert.Equal(t, int64(0), count)
// 		// reset count
// 		count = -1
// 		db.Model(&models.Enricher{}).Count(&count)
// 		assert.Equal(t, int64(0), count)

// 		// create router again without ensembler and enricher
// 		routerVersion.EnricherID = sql.NullInt32{}
// 		routerVersion.Enricher = nil
// 		routerVersion.EnsemblerID = sql.NullInt32{}
// 		routerVersion.Ensembler = nil
// 		routerVersion, err = svc.Save(routerVersion)
// 		assert.NoError(t, err)
// 		assert.Nil(t, routerVersion.Enricher)
// 		assert.Nil(t, routerVersion.Ensembler)
// 	})
// }

func TestListRouterVersions(t *testing.T) {
	tests := map[string]struct {
		routerID models.ID
		setup    func(*testing.T) service.RouterVersionsService
		verify   func(*testing.T, []*models.RouterVersion, error)
	}{
		"repository error": {
			setup: func(t *testing.T) service.RouterVersionsService {
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					List(gomock.Any()).
					Return(nil, errors.New("test DB error"))
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{})
			},
			verify: func(t *testing.T, versions []*models.RouterVersion, err error) {
				assert.ErrorContains(t, err, "test DB error")
			},
		},
		"monitoring URL error": {
			routerID: 1,
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", 1, 1, 1, 1)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					List(gomock.Eq(models.ID(1))).
					Return([]*models.RouterVersion{rv}, nil)
				monitoringSvc := mock_service.NewMockRouterMonitoringService(ctrl)
				monitoringSvc.EXPECT().
					GenerateMonitoringURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("test Monitoring error"))
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{
					RouterMonitoringService: monitoringSvc,
				})
			},
			verify: func(t *testing.T, versions []*models.RouterVersion, err error) {
				assert.ErrorContains(t, err, "test Monitoring error")
			},
		},
		"empty list": {
			routerID: 1,
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", 1, 1, 1, 1)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					List(gomock.Eq(models.ID(1))).
					Return([]*models.RouterVersion{rv}, nil)
				monitoringSvc := mock_service.NewMockRouterMonitoringService(ctrl)
				monitoringSvc.EXPECT().
					GenerateMonitoringURL(rv.Router.ProjectID, rv.Router.EnvironmentName, rv.Router.Name, &rv.Version).
					Return("http://monitoring", nil)
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{
					RouterMonitoringService: monitoringSvc,
				})
			},
			verify: func(t *testing.T, versions []*models.RouterVersion, err error) {
				rv := routerVersionWithMonitoringURL(
					createTestRouterVersion("test-router", 1, 1, 1, 1),
					"http://monitoring",
				)
				assert.Equal(t, []*models.RouterVersion{rv}, versions)
			},
		},
		"non-empty list": {
			routerID: 1,
			setup: func(t *testing.T) service.RouterVersionsService {
				versions := []*models.RouterVersion{
					createTestRouterVersion("test-router-1", 1, 1, 3, 4),
					createTestRouterVersion("test-router-2", 1, 1, 5, 6),
				}
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					List(gomock.Eq(models.ID(1))).
					Return(versions, nil)
				monitoringSvc := mock_service.NewMockRouterMonitoringService(ctrl)
				for _, rv := range versions {
					monitoringSvc.EXPECT().
						GenerateMonitoringURL(rv.Router.ProjectID, rv.Router.EnvironmentName, rv.Router.Name, &rv.Version).
						Return(fmt.Sprintf("http://monitoring/%d", rv.Version), nil)
				}
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{
					RouterMonitoringService: monitoringSvc,
				})
			},
			verify: func(t *testing.T, versions []*models.RouterVersion, err error) {
				expected := []*models.RouterVersion{
					routerVersionWithMonitoringURL(
						createTestRouterVersion("test-router-1", 1, 1, 3, 4),
						"http://monitoring/4",
					),
					routerVersionWithMonitoringURL(
						createTestRouterVersion("test-router-2", 1, 1, 5, 6),
						"http://monitoring/6",
					),
				}
				assert.Equal(t, expected, versions)
			},
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			svc := tt.setup(t)
			versions, err := svc.ListRouterVersions(tt.routerID)
			tt.verify(t, versions, err)
		})
	}
}

func createTestRouterVersion(name string, projectID, routerID, versionID, version int) *models.RouterVersion {
	return &models.RouterVersion{
		Model: models.Model{
			ID: models.ID(versionID),
		},
		Version:  uint(version),
		RouterID: models.ID(routerID),
		Router: &models.Router{
			Model: models.Model{
				ID: models.ID(routerID),
			},
			Name:            name,
			EnvironmentName: "test-environment",
			ProjectID:       models.ID(projectID),
		},
	}
}

func routerVersionWithMonitoringURL(rv *models.RouterVersion, url string) *models.RouterVersion {
	rv2 := *rv
	rv2.MonitoringURL = url
	return &rv2
}
