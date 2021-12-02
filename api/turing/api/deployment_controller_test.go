// +build unit

package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/api/client"
	tu "github.com/gojek/turing/api/turing/internal/testutils"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service/mocks"
	"github.com/gojek/turing/api/turing/utils"
	"github.com/gojek/turing/engines/experiment/v2/manager"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestDeployVersionSuccess(t *testing.T) {
	testEnv := "test-env"
	environment := &merlin.Environment{Name: testEnv}
	project := &mlp.Project{Id: 1, Name: "test-project"}
	router := &models.Router{
		Model: models.Model{
			ID: 1,
		},
		EnvironmentName: testEnv,
		Status:          "failed",
	}
	bqLogCfg := &models.LogConfig{
		ResultLoggerType: "bigquery",
		BigQueryConfig: &models.BigQueryConfig{
			ServiceAccountSecret: "svc-acct-secret",
		},
	}
	nopExpCfg := &models.ExperimentEngine{
		Type: "nop",
	}
	expCfg := &manager.TuringExperimentConfig{
		Client: manager.Client{
			ID:      "1",
			Passkey: "xp-passkey",
		},
	}
	expEnabledCfg := &models.ExperimentEngine{
		Type:   "xp",
		Config: expCfg,
	}

	// Define tests
	tests := map[string]struct {
		routerVersion    *models.RouterVersion
		pendingVersion   *models.RouterVersion
		validVersion     *models.RouterVersion
		expCfg           json.RawMessage
		decryptedPasskey string
	}{
		"nop_experiment": {
			routerVersion: &models.RouterVersion{
				Model: models.Model{
					ID: 2,
				},
				LogConfig:        bqLogCfg,
				ExperimentEngine: nopExpCfg,
				Status:           "test-status",
			},
			pendingVersion: &models.RouterVersion{
				Model: models.Model{
					ID: 2,
				},
				LogConfig:        bqLogCfg,
				ExperimentEngine: nopExpCfg,
				Status:           "pending",
			},
			validVersion: &models.RouterVersion{
				Model: models.Model{
					ID: 2,
				},
				LogConfig:        bqLogCfg,
				ExperimentEngine: nopExpCfg,
				Status:           "deployed",
			},
			expCfg: json.RawMessage(nil),
		},
		"experiment_enabled": {
			routerVersion: &models.RouterVersion{
				Model: models.Model{
					ID: 2,
				},
				LogConfig:        bqLogCfg,
				ExperimentEngine: expEnabledCfg,
				Status:           "test-status",
			},
			pendingVersion: &models.RouterVersion{
				Model: models.Model{
					ID: 2,
				},
				LogConfig:        bqLogCfg,
				ExperimentEngine: expEnabledCfg,
				Status:           "pending",
			},
			validVersion: &models.RouterVersion{
				Model: models.Model{
					ID: 2,
				},
				LogConfig:        bqLogCfg,
				ExperimentEngine: expEnabledCfg,
				Status:           "deployed",
			},
			expCfg:           json.RawMessage([]byte(`{"engine": "xp"}`)),
			decryptedPasskey: "xp-passkey-dec",
		},
	}

	// Set up common mock services
	mlps := &mocks.MLPService{}
	mlps.On("GetEnvironment", testEnv).Return(environment, nil)
	mlps.On("GetSecret", models.ID(project.Id), "svc-acct-secret").Return("service-acct", nil)

	rs := &mocks.RoutersService{}
	rs.On("Save", router).Return(nil, nil)
	rs.On("FindByID", uint(1)).Return(router, nil)

	es := &mocks.EventService{}
	es.On("ClearEvents", int(router.ID)).Return(nil)
	es.On("Save", mock.Anything).Return(nil)

	cs := &mocks.CryptoService{}
	cs.On("Decrypt", "xp-passkey").Return("xp-passkey-dec", nil)

	exps := &mocks.ExperimentsService{}
	exps.On("IsStandardExperimentManager", "nop").Return(false)
	exps.On("IsStandardExperimentManager", mock.Anything).Return(true)
	exps.On("GetStandardExperimentConfig", expCfg).Return(*expCfg, nil)
	exps.On("GetExperimentRunnerConfig", "xp", expCfg).Return(json.RawMessage([]byte(`{"engine": "xp"}`)), nil)

	// Run tests and validate
	for name, data := range tests {
		t.Run(name, func(t *testing.T) {
			eventsCh := utils.NewEventChannel()
			defer eventsCh.Close()

			// Set up test-specific mock services
			rvs := &mocks.RouterVersionsService{}
			rvs.On("Save", data.pendingVersion).Return(data.pendingVersion, nil)
			rvs.On("Save", data.validVersion).Return(data.validVersion, nil)

			ds := &mocks.DeploymentService{}
			ds.On("DeployRouterVersion", project, environment, data.pendingVersion, "service-acct",
				"", "", data.expCfg, data.decryptedPasskey, eventsCh).Return("test-url", nil)

			// Create test controller
			ctrl := RouterDeploymentController{
				BaseController{
					AppContext: &AppContext{
						MLPService:            mlps,
						EventService:          es,
						DeploymentService:     ds,
						RoutersService:        rs,
						RouterVersionsService: rvs,
						CryptoService:         cs,
						ExperimentsService:    exps,
					},
				},
			}

			// Run deploy and test that the router version's status is deployed and the endpoint
			// returned by deploy version is expected.
			endpoint, err := ctrl.deployRouterVersion(project, environment, data.routerVersion, eventsCh)
			assert.NoError(t, err)
			assert.Equal(t, models.RouterVersionStatusDeployed, data.routerVersion.Status)
			assert.Equal(t, "test-url", endpoint)
		})
	}
}

func TestRollbackVersionSuccess(t *testing.T) {
	testEnv := "test-env"
	environment := &merlin.Environment{Name: testEnv}
	project := &mlp.Project{Name: "test-project"}

	router := &models.Router{
		Model: models.Model{
			ID: 100,
		},
		Name:            "router",
		EnvironmentName: testEnv,
		Endpoint:        "current-endpoint",
		Status:          "deployed",
	}
	currVer := &models.RouterVersion{
		Model: models.Model{
			ID: 200,
		},
		Router: router,
		LogConfig: &models.LogConfig{
			ResultLoggerType: "bigquery",
			BigQueryConfig: &models.BigQueryConfig{
				ServiceAccountSecret: "svc-acct-secret",
			},
		},
		ExperimentEngine: &models.ExperimentEngine{
			Type: "nop",
		},
		Status: "deployed",
	}
	newVer := &models.RouterVersion{
		Model: models.Model{
			ID: 300,
		},
		Router: router,
		LogConfig: &models.LogConfig{
			ResultLoggerType: "bigquery",
			BigQueryConfig: &models.BigQueryConfig{
				ServiceAccountSecret: "svc-acct-secret",
			},
		},
		ExperimentEngine: &models.ExperimentEngine{
			Type: "nop",
		},
		Status: "pending",
	}
	newVerFailed := &models.RouterVersion{
		Model: models.Model{
			ID: 300,
		},
		Router: router,
		LogConfig: &models.LogConfig{
			ResultLoggerType: "bigquery",
			BigQueryConfig: &models.BigQueryConfig{
				ServiceAccountSecret: "svc-acct-secret",
			},
		},
		ExperimentEngine: &models.ExperimentEngine{
			Type: "nop",
		},
		Status: "failed",
		Error:  "error",
	}
	router.CurrRouterVersion = currVer
	testSvcAcct := "service-acct"

	// Set up mock services
	mlps := &mocks.MLPService{}
	mlps.On("GetEnvironment", testEnv).Return(environment, nil)
	mlps.On("GetSecret", models.ID(project.Id), "svc-acct-secret").Return(testSvcAcct, nil)

	rs := &mocks.RoutersService{}
	rs.On("Save", mock.Anything).Return(nil, nil)
	rs.On("FindByID", models.ID(100)).Return(router, nil)

	rvs := &mocks.RouterVersionsService{}
	rvs.On("FindByID", models.ID(200)).Return(currVer, nil)
	rvs.On("Save", currVer).Return(currVer, nil)
	rvs.On("Save", newVerFailed).Return(newVerFailed, nil)

	ds := &mocks.DeploymentService{}
	ds.On("DeployRouterVersion", project, environment, newVer, testSvcAcct,
		"", "", json.RawMessage(nil), "", mock.Anything).Return("", errors.New("error"))
	ds.On("UndeployRouterVersion", project, environment, newVer, mock.Anything).
		Return(nil)

	es := &mocks.EventService{}
	es.On("ClearEvents", int(router.ID)).Return(nil)
	es.On("Save", mock.Anything).Return(nil)

	exps := &mocks.ExperimentsService{}
	exps.On("IsStandardExperimentManager", "nop").Return(false)

	// Create test controller
	ctrl := RouterDeploymentController{
		BaseController{
			AppContext: &AppContext{
				MLPService:            mlps,
				DeploymentService:     ds,
				RoutersService:        rs,
				RouterVersionsService: rvs,
				EventService:          es,
				ExperimentsService:    exps,
			},
		},
	}

	// Run test method
	err := ctrl.deployOrRollbackRouter(project, router, newVer)
	assert.Error(t, err)
	// Assert that the call to undeploy failed version happened and the current ver ref
	// is correct, and the endpoint value remains unchanged. Also test that the statuses -
	// the new vers's deployment status should be failed and the router and the current
	// ver should be deployed.
	ds.AssertCalled(t, "UndeployRouterVersion", project, environment, newVer, mock.Anything)
	assert.Equal(t, models.ID(200), router.CurrRouterVersion.ID)
	assert.Equal(t, "current-endpoint", router.Endpoint)
	assert.Equal(t, models.RouterVersionStatusDeployed, router.CurrRouterVersion.Status)
	assert.Equal(t, models.RouterStatusDeployed, router.Status)
	assert.Equal(t, models.RouterVersionStatusFailed, newVer.Status)
}

func TestUndeployRouterSuccess(t *testing.T) {
	testEnv := "test-env"
	environment := &merlin.Environment{Name: testEnv}
	project := &mlp.Project{Name: "test-project"}

	// Create test router / version
	routerVersion := &models.RouterVersion{
		Model: models.Model{
			ID: 1,
		},
		Status: "deployed",
	}
	pendingRouterVersion := &models.RouterVersion{
		Model: models.Model{
			ID: 1,
		},
		Status: "pending",
	}
	undeployedRouterVersion := &models.RouterVersion{
		Model: models.Model{
			ID: 1,
		},
		Status: "undeployed",
	}
	router := &models.Router{
		Model: models.Model{
			ID: 1,
		},
		EnvironmentName:     testEnv,
		Endpoint:            "current-endpoint",
		CurrRouterVersion:   routerVersion,
		CurrRouterVersionID: sql.NullInt32{Int32: int32(1), Valid: true},
		Status:              "deployed",
	}
	modifiedRouter := &models.Router{
		Model: models.Model{
			ID: 1,
		},
		EnvironmentName:     testEnv,
		CurrRouterVersion:   undeployedRouterVersion,
		CurrRouterVersionID: sql.NullInt32{Int32: int32(1), Valid: true},
		Status:              "undeployed",
	}

	eventsCh := make(chan *models.Event)
	defer close(eventsCh)

	// Set up mock services
	mlps := &mocks.MLPService{}
	mlps.On("GetEnvironment", testEnv).Return(environment, nil)

	rs := &mocks.RoutersService{}
	rs.On("Save", modifiedRouter).Return(modifiedRouter, nil)

	rvs := &mocks.RouterVersionsService{}
	rvs.On("FindByID", models.ID(1)).Return(routerVersion, nil)
	rvs.On("Save", undeployedRouterVersion).Return(undeployedRouterVersion, nil)
	rvs.On("ListRouterVersions", models.ID(1)).
		Return([]*models.RouterVersion{routerVersion, pendingRouterVersion}, nil)

	ds := &mocks.DeploymentService{}
	ds.On("UndeployRouterVersion", project, environment, routerVersion, mock.Anything).Return(nil)
	ds.On("UndeployRouterVersion", project, environment, pendingRouterVersion, mock.Anything).Return(nil)
	ds.On("DeleteRouterEndpoint", project, environment, &models.RouterVersion{Router: router}).Return(nil)

	es := &mocks.EventService{}
	es.On("ClearEvents", int(router.ID)).Return(nil)
	es.On("Save", mock.Anything).Return(nil)

	// Create test controller
	ctrl := RouterDeploymentController{
		BaseController{
			AppContext: &AppContext{
				MLPService:            mlps,
				DeploymentService:     ds,
				RoutersService:        rs,
				RouterVersionsService: rvs,
				EventService:          es,
			},
		},
	}

	// Run test and validate
	err := ctrl.undeployRouter(project, router)
	// Test outcomes - no error, current version status is undeployed, empty endpoint
	assert.NoError(t, err)
	if router.CurrRouterVersion == nil {
		tu.FailOnError(t, fmt.Errorf("Current Version is not expected to be nil"))
	}
	assert.Equal(t, "undeployed", string(router.CurrRouterVersion.Status))
	assert.Equal(t, "", router.Endpoint)
	// Assert calls
	ds.AssertCalled(t, "UndeployRouterVersion", project, environment, routerVersion, mock.Anything)
	ds.AssertCalled(t, "UndeployRouterVersion", project, environment, pendingRouterVersion, mock.Anything)
	ds.AssertCalled(t, "DeleteRouterEndpoint", project, environment, &models.RouterVersion{Router: router})
	rs.AssertCalled(t, "Save", modifiedRouter)
}
