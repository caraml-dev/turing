// +build unit

package service

import (
	"context"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/patrickmn/go-cache"
	"github.com/stretchr/testify/assert"
	"golang.org/x/oauth2/google"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/api/client"
	tu "github.com/gojek/turing/api/turing/internal/testutils"
)

type mockCryptoService struct{}

func (*mockCryptoService) Encrypt(text string) (string, error) {
	return text, nil
}

func (*mockCryptoService) Decrypt(text string) (string, error) {
	return text, nil
}

// testSetupEnvForGoogleCredentials creates a temporary file containing dummy service account JSON
// then set the environment variable GOOGLE_APPLICATION_CREDENTIALS to point to the the file.
//
// This is useful for tests that assume Google Cloud Client libraries can automatically find
// the service account credentials in any environment.
//
// At the end of the test, the returned function can be called to perform cleanup.
func testSetupEnvForGoogleCredentials(t *testing.T) (reset func()) {
	serviceAccountKey := []byte(`{
  "type": "service_account",
  "project_id": "foo",
  "private_key_id": "bar",
  "private_key": "baz",
  "client_email": "foo@example.com",
  "client_id": "bar_client_id",
  "auth_uri": "https://oauth2.googleapis.com/auth",
  "token_uri": "https://oauth2.googleapis.com/token"
}`)

	file, err := ioutil.TempFile("", "dummy-service-account")
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(file.Name(), serviceAccountKey, 0644)
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

	// Create test Google client
	gc, err := google.DefaultClient(context.Background(), "https://www.googleapis.com/auth/userinfo.email")
	tu.FailOnError(t, err)
	// Create test projects and environments
	projects := []mlp.Project{{Id: 1}}
	environments := []merlin.Environment{{Name: "dev"}}

	// Patch new Merlin and MLP Client methods
	defer monkey.UnpatchAll()
	monkey.Patch(newMerlinClient,
		func(googleClient *http.Client, basePath string) *merlinClient {
			assert.Equal(t, gc, googleClient)
			assert.Equal(t, "merlin-base-path", basePath)
			// Create test client
			merlinClient := &merlinClient{
				api: &merlin.APIClient{
					EnvironmentApi: &merlin.EnvironmentApiService{},
					SecretApi:      &merlin.SecretApiService{},
				},
			}
			// Patch Get Environments
			monkey.PatchInstanceMethod(reflect.TypeOf(merlinClient.api.EnvironmentApi), "EnvironmentsGet",
				func(svc *merlin.EnvironmentApiService,
					ctx context.Context,
					localVarOptionals *merlin.EnvironmentApiEnvironmentsGetOpts,
				) ([]merlin.Environment, *http.Response, error) {
					return environments, nil, nil
				})
			return merlinClient
		},
	)
	monkey.Patch(newMLPClient,
		func(googleClient *http.Client, basePath string, encryptionKey string) *mlpClient {
			assert.Equal(t, gc, googleClient)
			assert.Equal(t, "mlp-base-path", basePath)
			assert.Equal(t, "mlp-enc-key", encryptionKey)
			// Create test client
			mlpClient := &mlpClient{
				CryptoService: &mockCryptoService{},
				api: &mlp.APIClient{
					ProjectApi: &mlp.ProjectApiService{},
				},
			}
			// Patch Get Projects
			monkey.PatchInstanceMethod(reflect.TypeOf(mlpClient.api.ProjectApi), "ProjectsGet",
				func(svc *mlp.ProjectApiService, ctx context.Context, localVarOptionals *mlp.ProjectApiProjectsGetOpts,
				) ([]mlp.Project, *http.Response, error) {
					return projects, nil, nil
				})
			return mlpClient
		},
	)

	svc, err := NewMLPService("mlp-base-path", "mlp-enc-key", "merlin-base-path")
	assert.NoError(t, err)
	assert.NotNil(t, svc)
	// Test side effects
	proj, err := svc.GetProject(1)
	tu.FailOnNil(t, proj)
	assert.Equal(t, projects[0], *proj)
	assert.NoError(t, err)
	env, err := svc.GetEnvironment("dev")
	tu.FailOnNil(t, env)
	assert.Equal(t, environments[0], *env)
	assert.NoError(t, err)
}

func TestNewMerlinClient(t *testing.T) {
	reset := testSetupEnvForGoogleCredentials(t)
	defer reset()

	// Create test Google client and Merlin Client
	gc, err := google.DefaultClient(context.Background(), "https://www.googleapis.com/auth/userinfo.email")
	tu.FailOnError(t, err)
	mc := &merlin.APIClient{}
	// Create expected Merlin config
	expectedCfg := merlin.NewConfiguration()
	expectedCfg.BasePath = "base-path"
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
	tu.FailOnError(t, err)
	// Create expected MLP config
	cfg := mlp.NewConfiguration()
	cfg.BasePath = "base-path"
	cfg.HTTPClient = gc

	// Test
	resultClient := newMLPClient(gc, "base-path", "enc-key")
	tu.FailOnNil(t, resultClient)
	assert.Equal(t, NewCryptoService("enc-key"), resultClient.CryptoService)
	assert.Equal(t, mlp.NewAPIClient(cfg), resultClient.api)
}

func TestMLPServiceGetProject(t *testing.T) {
	defer monkey.UnpatchAll()
	projects := []mlp.Project{
		{
			Id: 1,
		},
	}

	svc := newTestMLPService()
	monkey.PatchInstanceMethod(reflect.TypeOf(svc.mlpClient.api.ProjectApi), "ProjectsGet",
		func(svc *mlp.ProjectApiService, ctx context.Context, localVarOptionals *mlp.ProjectApiProjectsGetOpts,
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
	environments := []merlin.Environment{
		{
			Id:   1,
			Name: "env",
		},
	}

	svc := newTestMLPService()
	monkey.PatchInstanceMethod(reflect.TypeOf(svc.merlinClient.api.EnvironmentApi), "EnvironmentsGet",
		func(svc *merlin.EnvironmentApiService,
			ctx context.Context,
			localVarOptionals *merlin.EnvironmentApiEnvironmentsGetOpts,
		) ([]merlin.Environment, *http.Response, error) {
			return environments, nil, nil
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
			Id:   1,
			Name: "key",
			Data: "asd",
		},
	}

	svc := newTestMLPService()
	monkey.PatchInstanceMethod(
		reflect.TypeOf(svc.mlpClient.api.SecretApi),
		"ProjectsProjectIdSecretsGet",
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
				EnvironmentApi: &merlin.EnvironmentApiService{},
				SecretApi:      &merlin.SecretApiService{},
			},
		},
		mlpClient: &mlpClient{
			CryptoService: &mockCryptoService{},
			api: &mlp.APIClient{
				ProjectApi: &mlp.ProjectApiService{},
			},
		},
		cache: cache.New(time.Second*2, time.Second*2),
	}
	return svc
}
