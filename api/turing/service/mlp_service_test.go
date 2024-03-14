package service

import (
	"context"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	//nolint:all
	"bou.ke/monkey"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/oauth2/google"

	merlin "github.com/caraml-dev/merlin/client"
	mlp "github.com/caraml-dev/mlp/api/client"
)

// testSetupEnvForGoogleCredentials creates a temporary file containing dummy user account JSON
// then set the environment variable GOOGLE_APPLICATION_CREDENTIALS to point to the file.
//
// This is useful for tests that assume Google Cloud Client libraries can automatically find
// the user account credentials in any environment.
//
// At the end of the test, the returned function can be called to perform cleanup.
func testSetupEnvForGoogleCredentials(t *testing.T) (reset func()) {
	userAccountKey := []byte(`{
		"client_id": "dummyclientid.apps.googleusercontent.com",
		"client_secret": "dummy-secret",
		"quota_project_id": "test-project",
		"refresh_token": "dummy-token",
		"type": "authorized_user"
	}`)

	file, err := os.CreateTemp("", "dummy-user-account")
	if err != nil {
		t.Fatal(err)
	}

	err = os.WriteFile(file.Name(), userAccountKey, 0644)
	if err != nil {
		t.Fatal(err)
	}

	err = os.Setenv("GOOGLE_APPLICATION_CREDENTIALS", file.Name())
	if err != nil {
		t.Fatal(err)
	}

	return func() {
		err := os.Unsetenv("GOOGLE_APPLICATION_CREDENTIALS")
		if err != nil {
			t.Log("Cleanup failed", err)
		}
		err = os.Remove(file.Name())
		if err != nil {
			t.Log("Cleanup failed", err)
		}
	}
}

func TestNewMLPService(t *testing.T) {
	reset := testSetupEnvForGoogleCredentials(t)
	defer reset()

	// Create test projects and environments
	projects := []mlp.Project{{ID: 1}}
	environments := []merlin.Environment{{Name: "dev"}}

	// Patch new Merlin and MLP Client methods
	defer monkey.UnpatchAll()
	monkey.Patch(newMerlinClient,
		func(googleClient *http.Client, basePath string) *merlinClient {
			assert.Equal(t, "merlin-base-path", basePath)
			// Create test client
			merlinClient := &merlinClient{
				api: &merlin.APIClient{
					EnvironmentAPI: &merlin.EnvironmentAPIService{},
					SecretAPI:      &merlin.SecretAPIService{},
				},
			}
			// Patch Get Environments
			monkey.PatchInstanceMethod(reflect.TypeOf(merlinClient.api.EnvironmentAPI), "EnvironmentsGet",
				func(svc *merlin.EnvironmentAPIService,
					ctx context.Context,
				) merlin.ApiEnvironmentsGetRequest {
					apiRequest := merlin.ApiEnvironmentsGetRequest{}
					monkey.PatchInstanceMethod(reflect.TypeOf(apiRequest), "Execute",
						func(_ merlin.ApiEnvironmentsGetRequest) ([]merlin.Environment, *http.Response, error) {
							return environments, nil, nil
						})
					return apiRequest
				})
			return merlinClient
		},
	)
	monkey.Patch(newMLPClient,
		func(googleClient *http.Client, basePath string) *mlpClient {
			assert.Equal(t, "mlp-base-path", basePath)
			// Create test client
			mlpClient := &mlpClient{
				api: &mlp.APIClient{
					ProjectApi: &mlp.ProjectApiService{},
				},
			}
			// Patch Get Projects
			monkey.PatchInstanceMethod(reflect.TypeOf(mlpClient.api.ProjectApi), "V1ProjectsGet",
				func(svc *mlp.ProjectApiService, ctx context.Context, localVarOptionals *mlp.ProjectApiV1ProjectsGetOpts,
				) ([]mlp.Project, *http.Response, error) {
					return projects, nil, nil
				})
			return mlpClient
		},
	)

	svc, err := NewMLPService("mlp-base-path", "merlin-base-path")
	assert.NoError(t, err)
	assert.NotNil(t, svc)

	// Test side effects
	proj, err := svc.GetProject(1)
	require.NotNil(t, proj)
	assert.Equal(t, projects[0], *proj)
	assert.NoError(t, err)

	env, err := svc.GetEnvironment("dev")
	require.NotNil(t, env)
	assert.Equal(t, environments[0], *env)
	assert.NoError(t, err)
}

func TestNewMerlinClient(t *testing.T) {
	reset := testSetupEnvForGoogleCredentials(t)
	defer reset()

	// Create test Google client and Merlin Client
	gc, err := google.DefaultClient(context.Background(), "https://www.googleapis.com/auth/userinfo.email")
	require.NoError(t, err)
	mc := &merlin.APIClient{}
	// Create expected Merlin config
	expectedCfg := merlin.NewConfiguration()
	expectedCfg.Servers = merlin.ServerConfigurations{
		{
			URL: "base-path",
		},
	}
	expectedCfg.HTTPClient = gc

	// Monkey patch merlin.NewAPIClient
	defer monkey.UnpatchAll()
	monkey.Patch(merlin.NewAPIClient, func(cfg *merlin.Configuration) *merlin.APIClient {
		assert.Equal(t, expectedCfg, cfg)
		return mc
	})

	// Test
	assert.Equal(t, &merlinClient{api: mc}, newMerlinClient(gc, "base-path"))
}

func TestNewMLPClient(t *testing.T) {
	reset := testSetupEnvForGoogleCredentials(t)
	defer reset()

	// Create test Google client and Merlin Client
	gc, err := google.DefaultClient(context.Background(), "https://www.googleapis.com/auth/userinfo.email")
	require.NoError(t, err)
	// Create expected MLP config
	cfg := mlp.NewConfiguration()
	cfg.BasePath = "base-path"
	cfg.HTTPClient = gc

	// Test
	resultClient := newMLPClient(gc, "base-path")
	require.NotNil(t, resultClient)
	assert.Equal(t, mlp.NewAPIClient(cfg), resultClient.api)
}

func TestMLPServiceGetProject(t *testing.T) {
	defer monkey.UnpatchAll()
	projects := []mlp.Project{
		{
			ID: 1,
		},
	}

	svc := newTestMLPService()
	monkey.PatchInstanceMethod(reflect.TypeOf(svc.mlpClient.api.ProjectApi), "V1ProjectsGet",
		func(svc *mlp.ProjectApiService, ctx context.Context, localVarOptionals *mlp.ProjectApiV1ProjectsGetOpts,
		) ([]mlp.Project, *http.Response, error) {
			return projects, nil, nil
		})

	// getting valid project should refresh cache and return the project
	project, err := svc.GetProject(1)
	assert.NoError(t, err)
	assert.Equal(t, *project, projects[0])

	// getting invalid project should return error
	_, err = svc.GetProject(2)
	assert.Error(t, err)
}

func TestMLPServiceGetEnvironment(t *testing.T) {
	defer monkey.UnpatchAll()
	envID := int32(1)
	environments := []merlin.Environment{
		{
			Id:   &envID,
			Name: "env",
		},
	}

	svc := newTestMLPService()
	monkey.PatchInstanceMethod(reflect.TypeOf(svc.merlinClient.api.EnvironmentAPI), "EnvironmentsGet",
		func(svc *merlin.EnvironmentAPIService,
			ctx context.Context,
		) merlin.ApiEnvironmentsGetRequest {
			apiRequest := merlin.ApiEnvironmentsGetRequest{}
			monkey.PatchInstanceMethod(reflect.TypeOf(apiRequest), "Execute",
				func(_ merlin.ApiEnvironmentsGetRequest) ([]merlin.Environment, *http.Response, error) {
					return environments, nil, nil
				})
			return apiRequest
		})

	// getting valid project should refresh cache and return the project
	env, err := svc.GetEnvironment("env")
	assert.NoError(t, err)
	assert.Equal(t, *env, environments[0])

	// getting invalid project should return error
	_, err = svc.GetEnvironment("notexists")
	assert.Error(t, err)
}

func TestMLPServiceGetSecret(t *testing.T) {
	defer monkey.UnpatchAll()
	secrets := []mlp.Secret{
		{
			ID:   1,
			Name: "key",
			Data: "asd",
		},
	}

	svc := newTestMLPService()
	monkey.PatchInstanceMethod(
		reflect.TypeOf(svc.mlpClient.api.SecretApi),
		"V1ProjectsProjectIdSecretsGet",
		func(svc *mlp.SecretApiService,
			ctx context.Context,
			projectId int32,
		) ([]mlp.Secret, *http.Response, error) {
			if projectId != 1 {
				return []mlp.Secret{}, nil, nil
			}
			return secrets, nil, nil
		})

	// getting valid project should refresh cache and return the project
	secret, err := svc.GetSecret(1, "key")
	assert.NoError(t, err)
	assert.Equal(t, secret, secrets[0].Data)

	// getting invalid secret should return error
	_, err = svc.GetSecret(2, "key")
	assert.Error(t, err)

	_, err = svc.GetSecret(1, "nope")
	assert.Error(t, err)
}

func newTestMLPService() *mlpService {
	svc := &mlpService{
		merlinClient: &merlinClient{
			api: &merlin.APIClient{
				EnvironmentAPI: &merlin.EnvironmentAPIService{},
				SecretAPI:      &merlin.SecretAPIService{},
			},
		},
		mlpClient: &mlpClient{
			api: &mlp.APIClient{
				ProjectApi: &mlp.ProjectApiService{},
			},
		},
		cache: cache.New(time.Second*2, time.Second*2),
	}
	return svc
}
