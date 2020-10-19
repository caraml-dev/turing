// +build e2e

package e2e

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// These constants are set according to the values used by the app,
// for the test deployment
const (
	deploymentWaitTimeoutSeconds = 200
	deleteTimeoutSeconds         = 20
	fluentdSyncIntervalSeconds   = 10
	// fluentdSyncMaxWaitSeconds is the max expected wait time for log entries to be written to BigQuery from FluentD
	fluentdSyncMaxWaitSeconds = 120
)

// Test configs
type testConfig struct {
	// TestID is a unique identifier for a test run to allow concurrent e2e tests to run independently.
	// The router name will currently contains TestID to ensure concurrent e2e tests will create routers with
	// distinct name. The test runner must ensure the TestID provided is unique across concurrent runs.
	TestID string `envconfig:"TEST_ID" required:"true"`
	// MockserverEndpoint will be used as the router endpoints in the e2e tests.
	// This endpoint is expected to handle POST request and returns a JSON object
	MockserverEndpoint                      string `envconfig:"MOCKSERVER_ENDPOINT" required:"true"`
	KServiceDomain                          string `envconfig:"KSERVICE_DOMAIN" default:"models.id.d.gods.golabs.io"`
	APIBasePath                             string `envconfig:"API_BASE_PATH" required:"true"`
	ClusterName                             string `envconfig:"MODEL_CLUSTER_NAME" required:"true"`
	ProjectID                               int    `envconfig:"PROJECT_ID" required:"true"`
	ProjectName                             string `envconfig:"PROJECT_NAME" required:"true"`
	TestLitmusPasskey                       string `envconfig:"TEST_LITMUS_PASSKEY" required:"true"`
	TestLitmusClientID                      string `envconfig:"TEST_LITMUS_CLIENT_ID" required:"true"`
	TestLitmusCASToken                      string `envconfig:"TEST_LITMUS_CAS_TOKEN" required:"true"`
	TestLitmusExperimentName                string `envconfig:"TEST_LITMUS_EXPERIMENT_NAME" required:"true"`
	TestLitmusExperimentID                  string `envconfig:"TEST_LITMUS_EXPERIMENT_ID" required:"true"`
	TestLitmusExperimentUnitType            string `envconfig:"TEST_LITMUS_EXPERIMENT_UNIT_TYPE" required:"true"`
	TestLitmusExperimentUnitIDForControl    string `envconfig:"TEST_LITMUS_EXPERIMENT_UNIT_ID_FOR_CONTROL" required:"true"`
	TestLitmusExperimentUnitIDForTreatment1 string `envconfig:"TEST_LITMUS_EXPERIMENT_UNIT_ID_FOR_TREATMENT_1" required:"true"`
	TestXpPasskey                           string `envconfig:"TEST_XP_PASSKEY" required:"true"`
	VaultAddress                            string `envconfig:"VAULT_ADDRESS" required:"true"`
	VaultToken                              string `envconfig:"VAULT_TOKEN" required:"true"`
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
		t.Run("UndeployRouter", TestUndeployRouter)
		t.Run("DeployRouterVersion_InvalidConfig", TestDeployRouterInvalidConfig)
		t.Run("DeployRouter", TestDeployValidConfig)
		t.Run("DeleteRouter", TestDeleteRouter)
	})
	t.Run("TestTrafficRules", func(t *testing.T) {
		t.Parallel()
		t.Run("DeployRouter", TestDeployRouterWithTrafficRules)
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

	checkLitmusExperimentPrerequisites(*cfg)

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

// checkLitmusExperimentPrerequisites ensures that the Litmus client and experiment required for e2e test have been
// set up correctly in Litmus engine. These assumptions will be used when testing Turing router with standard ensembler
// in e2e test.
func checkLitmusExperimentPrerequisites(c testConfig) {
	// makeLitmusRequest is a helper function to send HTTP requests to Litmus URL. It will return the text body
	// for succesful request and fail immediately for failed requests.
	sendLitmusRequest := func(url string) string {
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Fatal(err)
		}
		req.Header.Set("x-client-id", c.TestLitmusClientID)
		req.Header.Set("x-passkey", c.TestLitmusPasskey)
		req.Header.Set("x-cas-token", c.TestLitmusCASToken)
		httpClient := http.Client{Timeout: 3 * time.Second}
		resp, err := httpClient.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		if resp.StatusCode != http.StatusOK {
			log.Fatalf("GET %s. Status code want: 200, got: %d", url, resp.StatusCode)
		}
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		return string(data)
	}

	log.Println("Ensure client 'data_science_platform' has been created")
	body := sendLitmusRequest("http://integration-litmus-internal.golabs.io/internal/clients")
	if !strings.Contains(body, `"name":"data_science_platform"`) {
		log.Fatal("client 'data_science_platform' is not yet registered in Litmus")
	}

	log.Println("Ensure experiment 'exp_exp_test_experiment_dsp_1' has been created and the status is 'running'")
	body = sendLitmusRequest("http://integration-litmus-internal.golabs.io/internal/experiment/" + c.TestLitmusExperimentID)
	if !strings.Contains(body, `"status":"running"`) {
		log.Fatal("experiment 'exp_exp_test_experiment_dsp_1' is not yet registered and/or not running in Litmus")
	}

	log.Println("Ensure variant 'control' is returned from 'run experiment' request for " + c.TestLitmusExperimentUnitType + ":" + c.TestLitmusExperimentUnitIDForControl)
	url := fmt.Sprintf("http://integration-litmus-internal.golabs.io/internal/v2/run/experiment/%s?unit-type=%s&unit-id=%s", c.TestLitmusExperimentName, c.TestLitmusExperimentUnitType, c.TestLitmusExperimentUnitIDForControl)
	body = sendLitmusRequest(url)
	if !strings.Contains(body, `"variant":"control"`) {
		log.Fatal("run experiment does not return the expected variant")
	}

	log.Println("Ensure variant 'treatment-1' is returned from 'run experiment' request for " + c.TestLitmusExperimentUnitType + ":" + c.TestLitmusExperimentUnitIDForTreatment1)
	url = fmt.Sprintf("http://integration-litmus-internal.golabs.io/internal/v2/run/experiment/%s?unit-type=%s&unit-id=%s", c.TestLitmusExperimentName, c.TestLitmusExperimentUnitType, c.TestLitmusExperimentUnitIDForTreatment1)
	body = sendLitmusRequest(url)
	if !strings.Contains(body, `"variant":"treatment-1"`) {
		log.Fatal("run experiment does not return the expected variant")
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
