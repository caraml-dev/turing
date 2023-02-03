package service_test

import (
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/api/client"
	"github.com/gojek/mlp/api/pkg/instrumentation/sentry"

	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/imagebuilder"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/api/turing/service/mock_service"
)

type RouterDeploymentServiceTestSuite struct {
	suite.Suite
	config       *config.Config
	mlpSvc       service.MLPService
	routerSvc    service.RoutersService
	imageBuilder imagebuilder.ImageBuilder
}

func TestRouterDeploymentService(t *testing.T) {
	suite.Run(t, new(RouterDeploymentServiceTestSuite))
}

func (s *RouterDeploymentServiceTestSuite) SetupSuite() {
	s.Suite.T().Log("Setting up RouterDeploymentServiceTestSuite")

	// Set up common dependencies
	testDuration, _ := time.ParseDuration("30s")
	s.config = &config.Config{
		DeployConfig: &config.DeploymentConfig{
			EnvironmentType: "dev",
			Timeout:         testDuration,
			DeletionTimeout: testDuration,
		},
		KnativeServiceDefaults: &config.KnativeServiceDefaults{
			QueueProxyResourcePercentage:    30,
			UserContainerLimitRequestFactor: 1.5,
		},
		RouterDefaults: &config.RouterDefaults{
			FluentdConfig: &config.FluentdConfig{
				Image:                "test-fluentd-image",
				Tag:                  "latest",
				FlushIntervalSeconds: 30,
				WorkerCount:          1,
			},
		},
		Sentry: sentry.Config{
			DSN: "test-dsn",
		},
	}

	// Set up common mock services
	ctrl := gomock.NewController(s.Suite.T())
	mlpSvc := mock_service.NewMockMLPService(ctrl)
	mlpSvc.EXPECT().
		GetEnvironment("test-environment").
		Return(&merlin.Environment{Name: "test-cluster"}, nil).
		AnyTimes()
	s.mlpSvc = mlpSvc
	routerSvc := mock_service.NewMockRoutersService(ctrl)
	routerSvc.EXPECT().
		Save(gomock.Any()).
		DoAndReturn(func(r *models.Router) (*models.Router, error) {
			return r, nil
		}).
		AnyTimes()
	s.routerSvc = routerSvc
}

func (s *RouterDeploymentServiceTestSuite) TestDeployOrRollbackRouter() {
	tests := map[string]struct {
		project       *mlp.Project
		routerVersion *models.RouterVersion
		setup         func(*testing.T) service.RouterDeploymentService
		verify        func(*testing.T, error)
	}{
		"deploy simple version": {
			project: &mlp.Project{ID: 1},
			routerVersion: &models.RouterVersion{
				Model: models.Model{
					ID: models.ID(4),
				},
				Version:  uint(3),
				RouterID: models.ID(2),
				Router: &models.Router{
					Model: models.Model{
						ID: models.ID(2),
					},
					Name:            "test-router",
					EnvironmentName: "test-environment",
					ProjectID:       models.ID(1),
					Status:          models.RouterStatusPending,
				},
				Status: models.RouterVersionStatusPending,
			},
			setup: func(t *testing.T) service.RouterDeploymentService {
				ctrl := gomock.NewController(t)
				eventSvc := mock_service.NewMockEventService(ctrl)
				// TBD: Set up mock cluster controller
				clusterControllers := map[string]cluster.Controller{
					"test-cluster": nil,
				}
				routerVersionSvc := mock_service.NewMockRouterVersionsService(ctrl)
				// Create new router deployment service
				return service.NewDeploymentService(
					s.config, clusterControllers, s.imageBuilder,
					&service.Services{
						EventService:          eventSvc,
						MLPService:            s.mlpSvc,
						RoutersService:        s.routerSvc,
						RouterVersionsService: routerVersionSvc,
					})
			},
			verify: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	}

	for name, tt := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			_ = tt.setup(t)
			// TBD: Uncomment verification
			// err := svc.DeployOrRollbackRouter(tt.project, tt.routerVersion.Router, tt.routerVersion)
			// tt.verify(t, err)
		})
	}
}
