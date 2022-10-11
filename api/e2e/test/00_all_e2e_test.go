//go:build e2e

package e2e

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/kelseyhightower/envconfig"
)

// These constants are set according to the values used by the app,
// for the test deployment
const (
	deploymentWaitTimeoutSeconds = 200
	deleteTimeoutSeconds         = 20
)

// Test configs
type testConfig struct {
	// TestID is a unique identifier for a test run to allow concurrent e2e tests to run independently.
	// The router name will currently contains TestID to ensure concurrent e2e tests will create routers with
	// distinct name. The test runner must ensure the TestID provided is unique across concurrent runs.
	TestID string `envconfig:"TEST_ID" required:"true"`
	// MockserverHTTPEndpoint will be used as the router endpoints in the e2e tests.
	// This endpoint is expected to handle POST request and returns a JSON object
	MockserverHTTPEndpoint string `envconfig:"MOCKSERVER_HTTP_ENDPOINT" required:"true"`
	MockserverUPIEndpoint  string `envconfig:"MOCKSERVER_UPI_ENDPOINT" required:"true"`
	KServiceDomain         string `envconfig:"KSERVICE_DOMAIN" default:"127.0.0.1.nip.io"`
	APIBasePath            string `envconfig:"API_BASE_PATH" required:"true"`
	ClusterName            string `envconfig:"MODEL_CLUSTER_NAME" required:"true"`
	DeploymentEnvironment  string `envconfig:"DEPLOYMENT_ENVIRONMENT" default:"dev"`
	ProjectID              int    `envconfig:"PROJECT_ID" required:"true"`
	ProjectName            string `envconfig:"PROJECT_NAME" required:"true"`
	TestEchoImage          string `envconfig:"TEST_ECHO_IMAGE" default:"eexit/mirror-http-server:1.1.3"`
	// KubeconfigUseLocal specifies whether the test helper should use local Kube config to
	// authenticate to the cluster. The Kube config is assumed to be available at $HOME/.kube/config.
	// If false, the helper will use the cluster credentials from the configured Vault environment.
	KubeconfigUseLocal bool   `envconfig:"KUBECONFIG_USE_LOCAL" default:"false"`
	KubeconfigFilePath string `envconfig:"KUBECONFIG_FILE_PATH" default:""`
	VaultAddress       string `envconfig:"VAULT_ADDRESS"`
	VaultToken         string `envconfig:"VAULT_TOKEN"`
}

func fromEnv() (*testConfig, error) {
	var cfg testConfig
	err := envconfig.Process("", &cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Test context
type TestContext struct {
	*testConfig
	clusterClients *TestClusterClients
	httpClient     *http.Client
}

// Global test context, accessible to each test case
var globalTestContext TestContext

// TestEndToEnd executes the test cases sequentially
func TestEndToEnd(t *testing.T) {
	// Run Tests
	t.Run("EndToEnd", func(t *testing.T) {
		t.Parallel()
		t.Run("CreateRouter_KnativeServices", TestCreateRouter)
		t.Run("UpdateRouter_InvalidConfig", TestUpdateRouterInvalidConfig)
		t.Run("CreateRouterVersion", TestCreateRouterVersion)
		t.Run("UndeployRouter", TestUndeployRouter)
		t.Run("DeployRouterVersion_InvalidConfig", TestDeployRouterInvalidConfig)
		t.Run("DeployRouter", TestDeployValidConfig)
		t.Run("DeleteRouter", TestDeleteRouter)
		t.Run("CreateRouter_Upi", TestUPIRouter)
	})
	t.Run("TestTrafficRules", func(t *testing.T) {
		t.Parallel()
		t.Run("DeployRouter", TestDeployRouterWithTrafficRules)
	})
	t.Run("TestStandardEnsembler", func(t *testing.T) {
		t.Parallel()
		t.Run("DeployRouter", TestDeployRouterWithStandardEnsembler)
	})
}

func TestMain(m *testing.M) {
	// Set up
	setUp()
	// Run tests
	code := m.Run()
	// Teardown
	tearDown()
	os.Exit(code)
}

func setUp() {
	fmt.Println("Setting up test context...")
	// Read env vars
	cfg, err := fromEnv()
	if err != nil {
		fmt.Println("Error reading env vars:", err)
		os.Exit(1)
	}

	// Init k8s clients
	clients, err := newClusterClients(cfg)
	if err != nil {
		fmt.Println("Error initialising cluster client:", err)
		os.Exit(1)
	}
	// Update the global test context
	globalTestContext = TestContext{
		testConfig:     cfg,
		clusterClients: clients,
		httpClient:     http.DefaultClient,
	}
}

func tearDown() {
	// Delete all cluster resources that were created for each experiment (and version)
	fmt.Println("Removing all cluster resources created by the tests...")
	deleteExperiments(
		globalTestContext.clusterClients,
		globalTestContext.ProjectName,
		[]struct {
			Name       string
			MaxVersion int
		}{
			{
				Name:       fmt.Sprintf("e2e-experiment-%s", globalTestContext.TestID),
				MaxVersion: 3,
			},
		},
	)
}
