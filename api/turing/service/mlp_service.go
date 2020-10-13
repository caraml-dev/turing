package service

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/patrickmn/go-cache"
	"golang.org/x/oauth2/google"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/client"
)

const (
	mlpCacheExpirySeconds  = 600
	mlpCacheCleanUpSeconds = 900
	mlpQueryTimeoutSeconds = 30
)

// MLPService provides a set of methods to interact with the MLP / Merlin APIs
type MLPService interface {
	// GetEnvironments gets all available environments from Merlin
	GetEnvironments() ([]merlin.Environment, error)
	// GetEnvironment gets the environment matching the provided name.
	GetEnvironment(name string) (*merlin.Environment, error)
	// GetProject gets the project matching the provided id.
	GetProject(id int) (*mlp.Project, error)
	// GetSecret gets a secret by project and name.
	GetSecret(projectID int, name string) (string, error)
}

type mlpService struct {
	merlinClient *merlinClient
	mlpClient    *mlpClient
	cache        *cache.Cache
}

type merlinClient struct {
	api *merlin.APIClient
}

func newMerlinClient(googleClient *http.Client, basePath string) *merlinClient {
	cfg := merlin.NewConfiguration()
	cfg.BasePath = basePath
	cfg.HTTPClient = googleClient

	return &merlinClient{
		api: merlin.NewAPIClient(cfg),
	}
}

type mlpClient struct {
	CryptoService
	api *mlp.APIClient
}

func newMLPClient(googleClient *http.Client, basePath string, encryptionKey string) *mlpClient {
	cfg := mlp.NewConfiguration()
	cfg.BasePath = basePath
	cfg.HTTPClient = googleClient

	return &mlpClient{
		CryptoService: NewCryptoService(encryptionKey),
		api:           mlp.NewAPIClient(cfg),
	}
}

// NewMLPService returns a service that retrieves information that is shared across MLP projects
// from (currently) the Merlin API.
func NewMLPService(
	mlpBasePath string,
	mlpEncryptionkey string,
	merlinBasePath string,
) (MLPService, error) {
	googleClient, err := google.DefaultClient(context.Background(), "https://www.googleapis.com/auth/userinfo.email")
	if err != nil {
		return nil, err
	}

	svc := &mlpService{
		merlinClient: newMerlinClient(googleClient, merlinBasePath),
		mlpClient:    newMLPClient(googleClient, mlpBasePath, mlpEncryptionkey),
		cache:        cache.New(mlpCacheExpirySeconds*time.Second, mlpCacheCleanUpSeconds*time.Second),
	}

	err = svc.refreshEnvironments()
	if err != nil {
		return nil, err
	}
	err = svc.refreshProjects()
	if err != nil {
		return nil, err
	}
	return svc, nil
}

// GetEnvironment gets the environment matching the provided name.
// This method will hit the cache first, and if not found, will call Merlin once to get
// the updated list of projects and refresh the cache, then try to get the value again.
// If still not found, will return a freecache NotFound error.
func (service mlpService) GetEnvironment(name string) (*merlin.Environment, error) {
	env, err := service.getEnvironment(name)
	if err != nil {
		err = service.refreshEnvironments()
		if err != nil {
			return nil, err
		}
		return service.getEnvironment(name)
	}
	return env, nil
}

func (service mlpService) getEnvironment(name string) (*merlin.Environment, error) {
	key := buildEnvironmentKey(name)
	cachedValue, found := service.cache.Get(key)
	if !found {
		return nil, fmt.Errorf("Environment info for %s not found in the cache", name)
	}
	// Cast the data
	environment, ok := cachedValue.(merlin.Environment)
	if !ok {
		return nil, fmt.Errorf("Malformed project info found in the cache for %s", name)
	}
	return &environment, nil
}

// GetProject gets the project matching the provided id.
// This method will hit the cache first, and if not found, will call Merlin once to get
// the updated list of projects and refresh the cache, then try to get the value again.
// If still not found, will return a freecache NotFound error.
func (service mlpService) GetProject(id int) (*mlp.Project, error) {
	project, err := service.getProject(id)
	if err != nil {
		err = service.refreshProjects()
		if err != nil {
			return nil, err
		}
		return service.getProject(id)
	}
	return project, nil
}

func (service mlpService) getProject(id int) (*mlp.Project, error) {
	key := buildProjectKey(int32(id))
	cachedValue, found := service.cache.Get(key)
	if !found {
		return nil, fmt.Errorf("Project info for id %d not found in the cache", id)
	}
	// Cast the data
	project, ok := cachedValue.(mlp.Project)
	if !ok {
		return nil, fmt.Errorf("Malformed project info found in the cache for id %d", id)
	}
	return &project, nil
}

// GetSecret gets a secret attached to a project by name.
func (service mlpService) GetSecret(projectID int, name string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mlpQueryTimeoutSeconds*time.Second)
	defer cancel()

	secrets, resp, err := service.mlpClient.api.SecretApi.ProjectsProjectIdSecretsGet(ctx, int32(projectID))
	if err != nil {
		return "", err
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	for _, secret := range secrets {
		if secret.Name == name {
			return service.mlpClient.Decrypt(secret.Data)
		}
	}
	return "", fmt.Errorf("secret %s not found in project %d", name, projectID)
}

func (service mlpService) refreshProjects() error {
	ctx, cancel := context.WithTimeout(context.Background(), mlpQueryTimeoutSeconds*time.Second)
	defer cancel()

	projects, resp, err := service.mlpClient.api.ProjectApi.ProjectsGet(ctx, nil)
	if err != nil {
		return err
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	for _, project := range projects {
		key := buildProjectKey(project.Id)
		service.cache.Set(key, project, cache.DefaultExpiration)
	}
	return nil
}

// GetEnvironments gets all available environments from Merlin. This method does not access
// the cache so as to always retrieve an updated list.
func (service mlpService) GetEnvironments() ([]merlin.Environment, error) {
	ctx, cancel := context.WithTimeout(context.Background(), mlpQueryTimeoutSeconds*time.Second)
	defer cancel()

	environments, resp, err := service.merlinClient.api.EnvironmentApi.EnvironmentsGet(ctx, nil)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	return environments, err
}

func (service mlpService) refreshEnvironments() error {
	environments, err := service.GetEnvironments()
	if err != nil {
		return err
	}

	for _, environment := range environments {
		key := buildEnvironmentKey(environment.Name)
		service.cache.Set(key, environment, cache.DefaultExpiration)
	}
	return nil
}

func buildProjectKey(id int32) string {
	return fmt.Sprintf("proj:%d", id)
}

func buildEnvironmentKey(name string) string {
	return fmt.Sprintf("env:%s", name)
}
