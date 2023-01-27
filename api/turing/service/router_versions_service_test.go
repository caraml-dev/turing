package service_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/api/turing/service/mock_service"
)

type RouterVersionsServiceTestSuite struct {
	suite.Suite
}

func TestRouterVersionsService(t *testing.T) {
	suite.Run(t, new(RouterVersionsServiceTestSuite))
}

func (s *RouterVersionsServiceTestSuite) TestListByRouterID() {
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
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusDeployed, 1, 1, 1, 1)
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
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusDeployed, 1, 1, 1, 1)
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
					createTestRouterVersion("test-router", models.RouterVersionStatusDeployed, 1, 1, 1, 1),
					"http://monitoring",
				)
				assert.Equal(t, []*models.RouterVersion{rv}, versions)
			},
		},
		"non-empty list": {
			routerID: 1,
			setup: func(t *testing.T) service.RouterVersionsService {
				versions := []*models.RouterVersion{
					createTestRouterVersion("test-router-1", models.RouterVersionStatusDeployed, 1, 1, 3, 4),
					createTestRouterVersion("test-router-2", models.RouterVersionStatusDeployed, 1, 1, 5, 6),
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
						createTestRouterVersion("test-router-1", models.RouterVersionStatusDeployed, 1, 1, 3, 4),
						"http://monitoring/4",
					),
					routerVersionWithMonitoringURL(
						createTestRouterVersion("test-router-2", models.RouterVersionStatusDeployed, 1, 1, 5, 6),
						"http://monitoring/6",
					),
				}
				assert.Equal(t, expected, versions)
			},
		},
	}

	for name, tt := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			svc := tt.setup(t)
			versions, err := svc.ListByRouterID(tt.routerID)
			tt.verify(t, versions, err)
		})
	}
}

func (s *RouterVersionsServiceTestSuite) TestListByRouterIDAndStatus() {
	tests := map[string]struct {
		routerID models.ID
		status   models.RouterVersionStatus
		setup    func(*testing.T) service.RouterVersionsService
		verify   func(*testing.T, []*models.RouterVersion, error)
	}{
		"repository error": {
			setup: func(t *testing.T) service.RouterVersionsService {
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					ListByStatus(gomock.Any(), gomock.Any()).
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
			status:   models.RouterVersionStatusPending,
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 1)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					ListByStatus(gomock.Eq(models.ID(1)), gomock.Eq(models.RouterVersionStatusPending)).
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
			status:   models.RouterVersionStatusPending,
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 1)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					ListByStatus(gomock.Eq(models.ID(1)), gomock.Eq(models.RouterVersionStatusPending)).
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
					createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 1),
					"http://monitoring",
				)
				assert.Equal(t, []*models.RouterVersion{rv}, versions)
			},
		},
		"non-empty list": {
			routerID: 1,
			status:   models.RouterVersionStatusPending,
			setup: func(t *testing.T) service.RouterVersionsService {
				versions := []*models.RouterVersion{
					createTestRouterVersion("test-router-1", models.RouterVersionStatusPending, 1, 1, 3, 4),
					createTestRouterVersion("test-router-2", models.RouterVersionStatusPending, 1, 1, 5, 6),
				}
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					ListByStatus(gomock.Eq(models.ID(1)), gomock.Eq(models.RouterVersionStatusPending)).
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
						createTestRouterVersion("test-router-1", models.RouterVersionStatusPending, 1, 1, 3, 4),
						"http://monitoring/4",
					),
					routerVersionWithMonitoringURL(
						createTestRouterVersion("test-router-2", models.RouterVersionStatusPending, 1, 1, 5, 6),
						"http://monitoring/6",
					),
				}
				assert.Equal(t, expected, versions)
			},
		},
	}

	for name, tt := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			svc := tt.setup(t)
			versions, err := svc.ListByRouterIDAndStatus(tt.routerID, tt.status)
			tt.verify(t, versions, err)
		})
	}
}

func (s *RouterVersionsServiceTestSuite) TestFindByID() {
	tests := map[string]struct {
		id     models.ID
		setup  func(*testing.T) service.RouterVersionsService
		verify func(*testing.T, *models.RouterVersion, error)
	}{
		"repository error": {
			setup: func(t *testing.T) service.RouterVersionsService {
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					FindByID(gomock.Any()).
					Return(nil, errors.New("test DB error"))
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{})
			},
			verify: func(t *testing.T, version *models.RouterVersion, err error) {
				assert.ErrorContains(t, err, "test DB error")
			},
		},
		"monitoring URL error": {
			id: 1,
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 1)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					FindByID(gomock.Eq(models.ID(1))).
					Return(rv, nil)
				monitoringSvc := mock_service.NewMockRouterMonitoringService(ctrl)
				monitoringSvc.EXPECT().
					GenerateMonitoringURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("test Monitoring error"))
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{
					RouterMonitoringService: monitoringSvc,
				})
			},
			verify: func(t *testing.T, version *models.RouterVersion, err error) {
				assert.ErrorContains(t, err, "test Monitoring error")
			},
		},
		"success": {
			id: 1,
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					FindByID(gomock.Eq(models.ID(1))).
					Return(rv, nil)
				monitoringSvc := mock_service.NewMockRouterMonitoringService(ctrl)
				monitoringSvc.EXPECT().
					GenerateMonitoringURL(rv.Router.ProjectID, rv.Router.EnvironmentName, rv.Router.Name, &rv.Version).
					Return("http://monitoring", nil)
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{
					RouterMonitoringService: monitoringSvc,
				})
			},
			verify: func(t *testing.T, version *models.RouterVersion, err error) {
				rv := routerVersionWithMonitoringURL(
					createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2),
					"http://monitoring",
				)
				assert.Equal(t, rv, version)
			},
		},
	}

	for name, tt := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			svc := tt.setup(t)
			version, err := svc.FindByID(tt.id)
			tt.verify(t, version, err)
		})
	}
}

func (s *RouterVersionsServiceTestSuite) TestFindByRouterIDAndVersion() {
	tests := map[string]struct {
		routerID models.ID
		version  uint
		setup    func(*testing.T) service.RouterVersionsService
		verify   func(*testing.T, *models.RouterVersion, error)
	}{
		"repository error": {
			setup: func(t *testing.T) service.RouterVersionsService {
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					FindByRouterIDAndVersion(gomock.Any(), gomock.Any()).
					Return(nil, errors.New("test DB error"))
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{})
			},
			verify: func(t *testing.T, version *models.RouterVersion, err error) {
				assert.ErrorContains(t, err, "test DB error")
			},
		},
		"monitoring URL error": {
			routerID: 1,
			version:  2,
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 1)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					FindByRouterIDAndVersion(gomock.Eq(models.ID(1)), gomock.Eq(uint(2))).
					Return(rv, nil)
				monitoringSvc := mock_service.NewMockRouterMonitoringService(ctrl)
				monitoringSvc.EXPECT().
					GenerateMonitoringURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("test Monitoring error"))
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{
					RouterMonitoringService: monitoringSvc,
				})
			},
			verify: func(t *testing.T, version *models.RouterVersion, err error) {
				assert.ErrorContains(t, err, "test Monitoring error")
			},
		},
		"success": {
			routerID: 1,
			version:  2,
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					FindByRouterIDAndVersion(gomock.Eq(models.ID(1)), gomock.Eq(uint(2))).
					Return(rv, nil)
				monitoringSvc := mock_service.NewMockRouterMonitoringService(ctrl)
				monitoringSvc.EXPECT().
					GenerateMonitoringURL(rv.Router.ProjectID, rv.Router.EnvironmentName, rv.Router.Name, &rv.Version).
					Return("http://monitoring", nil)
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{
					RouterMonitoringService: monitoringSvc,
				})
			},
			verify: func(t *testing.T, version *models.RouterVersion, err error) {
				rv := routerVersionWithMonitoringURL(
					createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2),
					"http://monitoring",
				)
				assert.Equal(t, rv, version)
			},
		},
	}

	for name, tt := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			svc := tt.setup(t)
			version, err := svc.FindByRouterIDAndVersion(tt.routerID, tt.version)
			tt.verify(t, version, err)
		})
	}
}

func (s *RouterVersionsServiceTestSuite) TestFindLatestVersionByRouterID() {
	tests := map[string]struct {
		routerID models.ID
		setup    func(*testing.T) service.RouterVersionsService
		verify   func(*testing.T, *models.RouterVersion, error)
	}{
		"repository error": {
			setup: func(t *testing.T) service.RouterVersionsService {
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					FindLatestVersion(gomock.Any()).
					Return(nil, errors.New("test DB error"))
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{})
			},
			verify: func(t *testing.T, version *models.RouterVersion, err error) {
				assert.ErrorContains(t, err, "test DB error")
			},
		},
		"monitoring URL error": {
			routerID: 1,
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 1)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					FindLatestVersion(gomock.Eq(models.ID(1))).
					Return(rv, nil)
				monitoringSvc := mock_service.NewMockRouterMonitoringService(ctrl)
				monitoringSvc.EXPECT().
					GenerateMonitoringURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("test Monitoring error"))
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{
					RouterMonitoringService: monitoringSvc,
				})
			},
			verify: func(t *testing.T, version *models.RouterVersion, err error) {
				assert.ErrorContains(t, err, "test Monitoring error")
			},
		},
		"success": {
			routerID: 1,
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					FindLatestVersion(gomock.Eq(models.ID(1))).
					Return(rv, nil)
				monitoringSvc := mock_service.NewMockRouterMonitoringService(ctrl)
				monitoringSvc.EXPECT().
					GenerateMonitoringURL(rv.Router.ProjectID, rv.Router.EnvironmentName, rv.Router.Name, &rv.Version).
					Return("http://monitoring", nil)
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{
					RouterMonitoringService: monitoringSvc,
				})
			},
			verify: func(t *testing.T, version *models.RouterVersion, err error) {
				rv := routerVersionWithMonitoringURL(
					createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2),
					"http://monitoring",
				)
				assert.Equal(t, rv, version)
			},
		},
	}

	for name, tt := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			svc := tt.setup(t)
			version, err := svc.FindLatestVersionByRouterID(tt.routerID)
			tt.verify(t, version, err)
		})
	}
}

func (s *RouterVersionsServiceTestSuite) TestCreate() {
	tests := map[string]struct {
		routerVersion *models.RouterVersion
		setup         func(*testing.T) service.RouterVersionsService
		verify        func(*testing.T, *models.RouterVersion, error)
	}{
		"repository error": {
			setup: func(t *testing.T) service.RouterVersionsService {
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					Save(gomock.Any()).
					Return(nil, errors.New("test DB error"))
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{})
			},
			verify: func(t *testing.T, version *models.RouterVersion, err error) {
				assert.ErrorContains(t, err, "test DB error")
			},
		},
		"monitoring URL error": {
			routerVersion: createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2),
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					Save(rv).
					Return(rv, nil)
				monitoringSvc := mock_service.NewMockRouterMonitoringService(ctrl)
				monitoringSvc.EXPECT().
					GenerateMonitoringURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("test Monitoring error"))
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{
					RouterMonitoringService: monitoringSvc,
				})
			},
			verify: func(t *testing.T, version *models.RouterVersion, err error) {
				assert.ErrorContains(t, err, "test Monitoring error")
			},
		},
		"success": {
			routerVersion: createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2),
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					Save(rv).
					Return(rv, nil)
				monitoringSvc := mock_service.NewMockRouterMonitoringService(ctrl)
				monitoringSvc.EXPECT().
					GenerateMonitoringURL(rv.Router.ProjectID, rv.Router.EnvironmentName, rv.Router.Name, &rv.Version).
					Return("http://monitoring", nil)
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{
					RouterMonitoringService: monitoringSvc,
				})
			},
			verify: func(t *testing.T, version *models.RouterVersion, err error) {
				rv := routerVersionWithMonitoringURL(
					createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2),
					"http://monitoring",
				)
				assert.Equal(t, rv, version)
			},
		},
	}

	for name, tt := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			svc := tt.setup(t)
			version, err := svc.Create(tt.routerVersion)
			tt.verify(t, version, err)
		})
	}
}

func (s *RouterVersionsServiceTestSuite) TestUpdate() {
	tests := map[string]struct {
		routerVersion *models.RouterVersion
		setup         func(*testing.T) service.RouterVersionsService
		verify        func(*testing.T, *models.RouterVersion, error)
	}{
		"repository error": {
			setup: func(t *testing.T) service.RouterVersionsService {
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					Save(gomock.Any()).
					Return(nil, errors.New("test DB error"))
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{})
			},
			verify: func(t *testing.T, version *models.RouterVersion, err error) {
				assert.ErrorContains(t, err, "test DB error")
			},
		},
		"monitoring URL error": {
			routerVersion: createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2),
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					Save(rv).
					Return(rv, nil)
				monitoringSvc := mock_service.NewMockRouterMonitoringService(ctrl)
				monitoringSvc.EXPECT().
					GenerateMonitoringURL(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
					Return("", errors.New("test Monitoring error"))
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{
					RouterMonitoringService: monitoringSvc,
				})
			},
			verify: func(t *testing.T, version *models.RouterVersion, err error) {
				assert.ErrorContains(t, err, "test Monitoring error")
			},
		},
		"success": {
			routerVersion: createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2),
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					Save(rv).
					Return(rv, nil)
				monitoringSvc := mock_service.NewMockRouterMonitoringService(ctrl)
				monitoringSvc.EXPECT().
					GenerateMonitoringURL(rv.Router.ProjectID, rv.Router.EnvironmentName, rv.Router.Name, &rv.Version).
					Return("http://monitoring", nil)
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{
					RouterMonitoringService: monitoringSvc,
				})
			},
			verify: func(t *testing.T, version *models.RouterVersion, err error) {
				rv := routerVersionWithMonitoringURL(
					createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 1, 1, 2),
					"http://monitoring",
				)
				assert.Equal(t, rv, version)
			},
		},
	}

	for name, tt := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			svc := tt.setup(t)
			version, err := svc.Update(tt.routerVersion)
			tt.verify(t, version, err)
		})
	}
}

func (s *RouterVersionsServiceTestSuite) TestDelete() {
	tests := map[string]struct {
		routerVersion *models.RouterVersion
		setup         func(*testing.T) service.RouterVersionsService
		verify        func(*testing.T, error)
	}{
		"router version deploying error": {
			routerVersion: createTestRouterVersion("test-router", models.RouterVersionStatusPending, 1, 2, 3, 4),
			setup: func(t *testing.T) service.RouterVersionsService {
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{})
			},
			verify: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "currently deploying")
			},
		},
		"current router version error": {
			routerVersion: createTestRouterVersion("test-router", models.RouterVersionStatusFailed, 1, 2, 3, 4),
			setup: func(t *testing.T) service.RouterVersionsService {
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rRepo.EXPECT().
					CountRoutersByCurrentVersionID(models.ID(3)).
					Return(int64(1))
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{})
			},
			verify: func(t *testing.T, err error) {
				assert.ErrorContains(t, err, "there exists a router that is currently using this version")
			},
		},
		"router version repository error": {
			routerVersion: createTestRouterVersion("test-router", models.RouterVersionStatusFailed, 1, 2, 3, 4),
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusFailed, 1, 2, 3, 4)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rRepo.EXPECT().
					CountRoutersByCurrentVersionID(models.ID(3)).
					Return(int64(0))
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					Delete(rv).
					Return(errors.New("test DB error"))
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{})
			},
			verify: func(t *testing.T, err error) {
				assert.Error(t, err, "test DB error")
			},
		},
		"success": {
			routerVersion: createTestRouterVersion("test-router", models.RouterVersionStatusFailed, 1, 2, 3, 4),
			setup: func(t *testing.T) service.RouterVersionsService {
				rv := createTestRouterVersion("test-router", models.RouterVersionStatusFailed, 1, 2, 3, 4)
				ctrl := gomock.NewController(t)
				rRepo := mock_service.NewMockRoutersRepository(ctrl)
				rRepo.EXPECT().
					CountRoutersByCurrentVersionID(models.ID(3)).
					Return(int64(0))
				rvRepo := mock_service.NewMockRouterVersionsRepository(ctrl)
				rvRepo.EXPECT().
					Delete(rv).
					Return(nil)
				// Create new router versions service
				return service.NewRouterVersionsService(rRepo, rvRepo, &service.Services{})
			},
			verify: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
		},
	}

	for name, tt := range tests {
		s.Suite.T().Run(name, func(t *testing.T) {
			svc := tt.setup(t)
			err := svc.Delete(tt.routerVersion)
			tt.verify(t, err)
		})
	}
}

func createTestRouterVersion(
	routerName string,
	status models.RouterVersionStatus,
	projectID, routerID, versionID, version int,
) *models.RouterVersion {
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
			Name:            routerName,
			EnvironmentName: "test-environment",
			ProjectID:       models.ID(projectID),
		},
		Status: status,
	}
}

func routerVersionWithMonitoringURL(rv *models.RouterVersion, url string) *models.RouterVersion {
	rv2 := *rv
	rv2.MonitoringURL = url
	return &rv2
}
