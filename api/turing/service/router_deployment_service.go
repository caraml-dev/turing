package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/api/client"
	"github.com/gojek/turing/api/turing/cluster"
	"github.com/gojek/turing/api/turing/cluster/labeller"
	"github.com/gojek/turing/api/turing/cluster/servicebuilder"
	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/imagebuilder"
	"github.com/gojek/turing/api/turing/models"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"
)

// DeploymentService handles the deployment of the Turing routers and the related components.
type DeploymentService interface {
	DeployRouterVersion(
		project *mlp.Project,
		environment *merlin.Environment,
		routerVersion *models.RouterVersion,
		routerServiceAccountKey string,
		enricherServiceAccountKey string,
		ensemblerServiceAccountKey string,
		pyfuncEnsembler *models.PyFuncEnsembler,
		experimentConfig json.RawMessage,
		eventsCh *EventChannel,
	) (string, error)
	UndeployRouterVersion(
		project *mlp.Project,
		environment *merlin.Environment,
		routerVersion *models.RouterVersion,
		eventsCh *EventChannel,
	) error
	DeleteRouterEndpoint(project *mlp.Project,
		environment *merlin.Environment,
		routerVersion *models.RouterVersion,
	) error
}
type deploymentService struct {
	// Deployment configs
	deploymentTimeout         time.Duration
	deploymentDeletionTimeout time.Duration
	environmentType           string

	// Router configs
	sentryEnabled  bool
	sentryDSN      string
	routerDefaults *config.RouterDefaults

	// Knative service configs
	knativeServiceConfig *config.KnativeServiceDefaults

	// Ensembler service image builder for real time ensemblers
	ensemblerServiceImageBuilder imagebuilder.ImageBuilder

	clusterControllers map[string]cluster.Controller
	svcBuilder         servicebuilder.ClusterServiceBuilder
}

// uFunc is the function type accepted by the updateKnServices method
type uFunc func(context.Context, *cluster.KnativeService, *sync.WaitGroup, chan<- error, *EventChannel)

// NewDeploymentService initialises a new endpoints service
func NewDeploymentService(
	cfg *config.Config,
	clusterControllers map[string]cluster.Controller,
	ensemblerServiceImageBuilder imagebuilder.ImageBuilder,
) DeploymentService {
	// Create cluster service builder
	sb := servicebuilder.NewClusterServiceBuilder(
		resource.Quantity(cfg.DeployConfig.MaxCPU),
		resource.Quantity(cfg.DeployConfig.MaxMemory),
	)

	return &deploymentService{
		deploymentTimeout:            cfg.DeployConfig.Timeout,
		deploymentDeletionTimeout:    cfg.DeployConfig.DeletionTimeout,
		environmentType:              cfg.DeployConfig.EnvironmentType,
		routerDefaults:               cfg.RouterDefaults,
		knativeServiceConfig:         cfg.KnativeServiceDefaults,
		ensemblerServiceImageBuilder: ensemblerServiceImageBuilder,
		sentryEnabled:                cfg.Sentry.Enabled,
		sentryDSN:                    cfg.Sentry.DSN,
		clusterControllers:           clusterControllers,
		svcBuilder:                   sb,
	}
}

// DeployRouterVersion deploys the given router version, returning its external url if successful
func (ds *deploymentService) DeployRouterVersion(
	project *mlp.Project,
	environment *merlin.Environment,
	routerVersion *models.RouterVersion,
	routerServiceAccountKey string,
	enricherServiceAccountKey string,
	ensemblerServiceAccountKey string,
	pyfuncEnsembler *models.PyFuncEnsembler,
	experimentConfig json.RawMessage,
	eventsCh *EventChannel,
) (string, error) {
	var endpoint string

	// If pyfunc ensembler is specified as an ensembler service, build/retrieve its image
	if pyfuncEnsembler != nil {
		err := ds.buildEnsemblerServiceImage(pyfuncEnsembler, project, routerVersion, eventsCh)
		if err != nil {
			return endpoint, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), ds.deploymentTimeout)
	defer cancel()

	// Get the cluster controller
	controller, err := ds.getClusterControllerByEnvironment(environment.Name)
	if err != nil {
		return endpoint, err
	}

	// Create namespace if not exists
	eventsCh.Write(models.NewInfoEvent(
		models.EventStageDeployingDependencies, "preparing namespace for project %s", project.Name))
	err = controller.CreateNamespace(ctx, project.Name)
	if err != nil && err != cluster.ErrNamespaceAlreadyExists {
		return endpoint, err
	}

	// Create secret
	eventsCh.Write(models.NewInfoEvent(models.EventStageDeployingDependencies, "deploying secret"))
	secret := ds.svcBuilder.NewSecret(
		routerVersion,
		project,
		routerServiceAccountKey,
		enricherServiceAccountKey,
		ensemblerServiceAccountKey,
	)
	err = createSecret(ctx, controller, secret)
	if err != nil {
		eventsCh.Write(models.NewErrorEvent(
			models.EventStageDeployingDependencies, "failed to create secret: %s", err.Error()))
		return endpoint, err
	}
	secretName := secret.Name

	// Construct service objects for each of the components and deploy
	services, err := ds.createServices(
		routerVersion, project, ds.environmentType, secretName, experimentConfig,
		ds.routerDefaults, ds.sentryEnabled, ds.sentryDSN,
		ds.knativeServiceConfig.TargetConcurrency, ds.knativeServiceConfig.QueueProxyResourcePercentage,
		ds.knativeServiceConfig.UserContainerLimitRequestFactor,
	)
	if err != nil {
		return endpoint, err
	}

	// Deploy fluentd if enabled
	if routerVersion.LogConfig.ResultLoggerType == models.BigQueryLogger {
		fluentdService := ds.svcBuilder.NewFluentdService(routerVersion, project,
			secretName, ds.routerDefaults.FluentdConfig)
		// Create pvc
		err = createPVC(ctx, controller, project.Name, fluentdService.PersistentVolumeClaim)
		if err != nil {
			eventsCh.Write(models.NewErrorEvent(
				models.EventStageDeployingDependencies, "failed to deploy fluentd service: %s", err.Error()))
			return endpoint, err
		}
		// Deploy fluentd
		err = deployK8sService(ctx, controller, fluentdService)
		if err != nil {
			eventsCh.Write(models.NewErrorEvent(
				models.EventStageDeployingDependencies, "failed to deploy fluentd service: %s", err.Error()))
			return endpoint, err
		}
		eventsCh.Write(models.NewInfoEvent(
			models.EventStageDeployingDependencies, "successfully deployed fluentd service"))
	}

	// Deploy experiment engine plugins server
	if routerVersion.ExperimentEngine.PluginConfig != nil {
		pluginsServerService := ds.svcBuilder.NewPluginsServerService(routerVersion, project)
		err = deployK8sService(ctx, controller, pluginsServerService)
		if err != nil {
			eventsCh.Write(models.NewErrorEvent(
				models.EventStageDeployingDependencies, "failed to deploy plugins server: %s", err.Error()))
			return endpoint, err
		}

		eventsCh.Write(models.NewInfoEvent(
			models.EventStageDeployingDependencies, "successfully deployed plugins server"))
	}

	err = deployKnServices(ctx, controller, services, eventsCh)
	if err != nil {
		return endpoint, err
	}

	// Get the router's external endpoint
	routerSvcName := ds.svcBuilder.GetRouterServiceName(routerVersion)
	endpoint = controller.GetKnativeServiceURL(ctx, routerSvcName, project.Name)

	// Deploy or update the virtual service
	eventsCh.Write(models.NewInfoEvent(models.EventStageUpdatingEndpoint, "updating router endpoint"))
	routerEndpoint, err := ds.svcBuilder.NewRouterEndpoint(routerVersion, project, ds.environmentType, endpoint)
	if err != nil {
		eventsCh.Write(models.NewErrorEvent(
			models.EventStageUpdatingEndpoint, "failed to update router endpoint: %s", err.Error()))
		return endpoint, err
	}
	err = controller.ApplyIstioVirtualService(ctx, routerEndpoint)
	if err == nil {
		eventsCh.Write(models.NewInfoEvent(
			models.EventStageUpdatingEndpoint, "successfully updated router endpoint to downstream %s", endpoint))
	} else {
		eventsCh.Write(models.NewErrorEvent(
			models.EventStageUpdatingEndpoint, "failed to update router endpoint: %s", err.Error()))
	}
	return "http://" + routerEndpoint.Endpoint, err
}

// UndeployRouterVersion removes the deployed router, if exists. Else, an error is returned.
func (ds *deploymentService) UndeployRouterVersion(
	project *mlp.Project,
	environment *merlin.Environment,
	routerVersion *models.RouterVersion,
	eventsCh *EventChannel,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), ds.deploymentTimeout)
	defer cancel()
	// Get the cluster controller
	controller, err := ds.getClusterControllerByEnvironment(environment.Name)
	if err != nil {
		return err
	}

	// Delete secret
	eventsCh.Write(models.NewInfoEvent(models.EventStageDeletingDependencies, "deleting secrets"))
	secret := ds.svcBuilder.NewSecret(routerVersion, project, "", "", "")
	err = deleteSecret(controller, secret)
	if err != nil {
		return err
	}

	// Construct service objects for each of the components to be deleted
	services, err := ds.createServices(
		routerVersion, project, ds.environmentType, "", nil,
		ds.routerDefaults, ds.sentryEnabled, ds.sentryDSN,
		ds.knativeServiceConfig.TargetConcurrency, ds.knativeServiceConfig.QueueProxyResourcePercentage,
		ds.knativeServiceConfig.UserContainerLimitRequestFactor,
	)
	if err != nil {
		return err
	}

	var errs []string
	// Delete fluentd if required
	if routerVersion.LogConfig.ResultLoggerType == models.BigQueryLogger {
		fluentdService := ds.svcBuilder.NewFluentdService(routerVersion,
			project, "", ds.routerDefaults.FluentdConfig)
		err = deleteK8sService(controller, fluentdService, ds.deploymentTimeout)
		if err != nil {
			errs = append(errs, err.Error())
		}
		err = deletePVC(controller, project.Name, fluentdService.PersistentVolumeClaim)
		if err != nil {
			errs = append(errs, err.Error())
		}
		if len(errs) == 0 {
			eventsCh.Write(models.NewInfoEvent(
				models.EventStageDeletingDependencies, "successfully deleted fluentd"))
		} else {
			eventsCh.Write(models.NewErrorEvent(
				models.EventStageDeletingDependencies, "failed to delete fluentd: %s", strings.Join(errs, ". ")))
		}
	}

	// Delete experiment engine plugins server
	if routerVersion.ExperimentEngine.PluginConfig != nil {
		pluginsServerSvc := ds.svcBuilder.NewPluginsServerService(routerVersion, project)
		err = deleteK8sService(controller, pluginsServerSvc, ds.deploymentTimeout)
		if err != nil {
			eventsCh.Write(
				models.NewErrorEvent(
					models.EventStageDeletingDependencies,
					"failed to delete plugins server: %s", strings.Join(errs, ". ")))
		} else {
			eventsCh.Write(models.NewInfoEvent(
				models.EventStageDeletingDependencies, "successfully deleted plugins server"))
		}
	}

	// Delete all components
	err = deleteKnServices(ctx, controller, services, ds.deploymentDeletionTimeout, eventsCh)
	if err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) != 0 {
		return errors.New(strings.Join(errs, ". "))
	}

	return nil
}

func (ds *deploymentService) DeleteRouterEndpoint(
	project *mlp.Project,
	environment *merlin.Environment,
	routerVersion *models.RouterVersion,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), ds.deploymentTimeout)
	defer cancel()
	// Get the cluster controller
	controller, err := ds.getClusterControllerByEnvironment(environment.Name)
	if err != nil {
		return err
	}

	routerEndpointName := fmt.Sprintf("%s-turing-router", routerVersion.Router.Name)
	return controller.DeleteIstioVirtualService(ctx, routerEndpointName, project.Name, ds.deploymentDeletionTimeout)
}

func (ds *deploymentService) getClusterControllerByEnvironment(
	environment string,
) (cluster.Controller, error) {
	controller, ok := ds.clusterControllers[environment]
	if !ok {
		return nil, errors.New("Deployment environment not supported")
	}
	return controller, nil
}

func (ds *deploymentService) createServices(
	routerVersion *models.RouterVersion,
	project *mlp.Project,
	envType string,
	secretName string,
	experimentConfig json.RawMessage,
	routerDefaults *config.RouterDefaults,
	sentryEnabled bool,
	sentryDSN string,
	knativeTargetConcurrency int,
	knativeQueueProxyResourcePercentage int,
	userContainerLimitRequestFactor float64,
) ([]*cluster.KnativeService, error) {
	services := []*cluster.KnativeService{}

	// Enricher
	if routerVersion.Enricher != nil {
		enricherSvc, err := ds.svcBuilder.NewEnricherService(
			routerVersion,
			project,
			envType,
			secretName,
			knativeTargetConcurrency,
			knativeQueueProxyResourcePercentage,
			userContainerLimitRequestFactor,
		)
		if err != nil {
			return services, err
		}
		services = append(services, enricherSvc)
	}

	// Ensembler
	if routerVersion.HasDockerConfig() {
		ensemblerSvc, err := ds.svcBuilder.NewEnsemblerService(
			routerVersion,
			project,
			envType,
			secretName,
			knativeTargetConcurrency,
			knativeQueueProxyResourcePercentage,
			userContainerLimitRequestFactor,
		)
		if err != nil {
			return services, err
		}
		services = append(services, ensemblerSvc)
	}

	// Router
	routerService, err := ds.svcBuilder.NewRouterService(
		routerVersion,
		project,
		envType,
		secretName,
		experimentConfig,
		routerDefaults,
		sentryEnabled,
		sentryDSN,
		knativeTargetConcurrency,
		knativeQueueProxyResourcePercentage,
		userContainerLimitRequestFactor,
	)
	if err != nil {
		return services, err
	}

	services = append(services, routerService)

	return services, nil
}

// buildEnsemblerServiceImage builds the pyfunc ensembler as a service specified in a Docker image
func (ds *deploymentService) buildEnsemblerServiceImage(
	ensembler *models.PyFuncEnsembler,
	project *mlp.Project,
	routerVersion *models.RouterVersion,
	eventsCh *EventChannel,
) error {
	// Build image corresponding to the retrieved ensembler
	request := imagebuilder.BuildImageRequest{
		ProjectName:  project.Name,
		ResourceName: ensembler.Name,
		ResourceID:   *routerVersion.Ensembler.PyfuncConfig.EnsemblerID,
		VersionID:    ensembler.RunID,
		ArtifactURI:  ensembler.ArtifactURI,
		BuildLabels: labeller.BuildLabels(
			labeller.KubernetesLabelsRequest{
				Stream: project.Stream,
				Team:   project.Team,
				App:    ensembler.Name,
			},
		),
		EnsemblerFolder: EnsemblerFolder,
	}
	eventsCh.Write(
		models.NewInfoEvent(
			models.EventStageDeployingDependencies,
			"building/retrieving pyfunc ensembler with project_id: %d and ensembler_id: %d",
			*routerVersion.Ensembler.PyfuncConfig.ProjectID,
			*routerVersion.Ensembler.PyfuncConfig.EnsemblerID,
		),
	)
	imageRef, imageBuildErr := ds.ensemblerServiceImageBuilder.BuildImage(request)
	if imageBuildErr != nil {
		return imageBuildErr
	}

	eventsCh.Write(
		models.NewInfoEvent(
			models.EventStageDeployingDependencies,
			"pyfunc ensembler with project_id: %d and ensembler_id: %d built/retrieved successfully",
			*routerVersion.Ensembler.PyfuncConfig.ProjectID,
			*routerVersion.Ensembler.PyfuncConfig.EnsemblerID,
		),
	)
	// Create a new docker config for the ensembler with the newly generated image
	routerVersion.Ensembler.DockerConfig = &models.EnsemblerDockerConfig{
		Image:           imageRef,
		ResourceRequest: routerVersion.Ensembler.PyfuncConfig.ResourceRequest,
		Timeout:         routerVersion.Ensembler.PyfuncConfig.Timeout,
		Endpoint:        PyFuncEnsemblerServiceEndpoint,
		Port:            PyFuncEnsemblerServicePort,
		Env:             routerVersion.Ensembler.PyfuncConfig.Env,
		ServiceAccount:  "",
	}

	return nil
}

// deployK8sService deploys a kubernetes service.
func deployK8sService(ctx context.Context, controller cluster.Controller, service *cluster.KubernetesService) error {
	select {
	case <-ctx.Done():
		return errors.New("timeout deploying service")
	default:
		return controller.DeployKubernetesService(ctx, service)
	}
}

// deleteK8sService deletes a kubernetes service.
func deleteK8sService(
	controller cluster.Controller,
	service *cluster.KubernetesService,
	timeout time.Duration,
) error {
	return controller.DeleteKubernetesService(context.Background(), service.Name, service.Namespace, timeout)
}

// createSecret creates a secret.
func createSecret(
	ctx context.Context,
	controller cluster.Controller,
	secret *cluster.Secret,
) error {
	select {
	case <-ctx.Done():
		return errors.New("timeout deploying service")
	default:
		return controller.CreateSecret(ctx, secret)
	}
}

// deleteSecret deletes a secret.
func deleteSecret(controller cluster.Controller, secret *cluster.Secret) error {
	return controller.DeleteSecret(context.Background(), secret.Name, secret.Namespace)
}

func createPVC(
	ctx context.Context,
	controller cluster.Controller,
	namespace string,
	pvc *cluster.PersistentVolumeClaim,
) error {
	select {
	case <-ctx.Done():
		return errors.New("timeout deploying service")
	default:
		return controller.ApplyPersistentVolumeClaim(ctx, namespace, pvc)
	}
}

func deletePVC(
	controller cluster.Controller,
	namespace string,
	pvc *cluster.PersistentVolumeClaim,
) error {
	return controller.DeletePersistentVolumeClaim(context.Background(), pvc.Name, namespace)
}

// deployKnServices deploys all services simulateneously and waits for all of them to
// be ready. This includes the enricher (if enabled), ensembler (if enabled) and router.
// Note: The enricher and ensembler have no health checks, so they will be ready immediately.
func deployKnServices(
	ctx context.Context,
	controller cluster.Controller,
	services []*cluster.KnativeService,
	eventsCh *EventChannel,
) error {
	// Define deploy function
	deployFunc := func(ctx context.Context,
		svc *cluster.KnativeService,
		wg *sync.WaitGroup,
		errCh chan<- error,
		eventsCh *EventChannel,
	) {
		defer wg.Done()
		eventsCh.Write(models.NewInfoEvent(
			models.EventStageDeployingServices, "deploying service %s", svc.Name))
		if svc.ConfigMap != nil {
			err := controller.ApplyConfigMap(context.Background(), svc.Namespace, svc.ConfigMap)
			if err != nil {
				err = errors.Wrapf(err, "Failed to apply config map %s", svc.ConfigMap.Name)
				eventsCh.Write(models.NewErrorEvent(
					models.EventStageDeployingServices, "failed to deploy service %s: %s", svc.Name, err.Error()))
				errCh <- err
				return
			}
		}

		err := controller.DeployKnativeService(ctx, svc)
		if err != nil {
			err = errors.Wrapf(err, "Failed to deploy %s", svc.Name)
			eventsCh.Write(models.NewErrorEvent(
				models.EventStageDeployingServices, "failed to deploy service %s: %s", svc.Name, err.Error()))
		} else {
			eventsCh.Write(models.NewInfoEvent(
				models.EventStageDeployingServices, "successfully deployed %s", svc.Name))
		}
		errCh <- err
	}

	select {
	case <-ctx.Done():
		return errors.New("timeout deploying service")
	default:
		return updateKnServices(ctx, services, deployFunc, eventsCh)
	}
}

// deleteKnServices simultaneously issues a delete call to all services and waits
// until deletion timeout, for a response
func deleteKnServices(
	ctx context.Context,
	controller cluster.Controller,
	services []*cluster.KnativeService,
	timeout time.Duration,
	eventsCh *EventChannel,
) error {
	// Define delete function
	deleteFunc := func(_ context.Context,
		svc *cluster.KnativeService,
		wg *sync.WaitGroup,
		errCh chan<- error,
		eventsCh *EventChannel,
	) {
		defer wg.Done()
		eventsCh.Write(models.NewInfoEvent(
			models.EventStageUndeployingServices, "deleting service %s", svc.Name))
		if svc.ConfigMap != nil {
			err := controller.DeleteConfigMap(ctx, svc.ConfigMap.Name, svc.Namespace)
			if err != nil {
				err = errors.Wrapf(err, "Failed to delete config map %s", svc.ConfigMap.Name)
				eventsCh.Write(models.NewErrorEvent(
					models.EventStageUndeployingServices, "failed to delete service %s: %s", svc.Name, err.Error()))
				errCh <- err
			}
		}
		err := controller.DeleteKnativeService(ctx, svc.Name, svc.Namespace, timeout)
		if err != nil {
			err = errors.Wrapf(err, "Error when deleting %s", svc.Name)
			eventsCh.Write(models.NewErrorEvent(
				models.EventStageUndeployingServices, "failed to delete service %s: %s", svc.Name, err.Error()))
		} else {
			eventsCh.Write(models.NewInfoEvent(
				models.EventStageUndeployingServices, "successfully deleted %s", svc.Name))
		}
		errCh <- err
	}

	return updateKnServices(ctx, services, deleteFunc, eventsCh)
}

// updateKnServices is a helper method for deployment / deletion of serivices that runs the
// given update function on the given services simultaneously and waits for a response,
// within the supplied timeout.
func updateKnServices(ctx context.Context, services []*cluster.KnativeService,
	updateFunc uFunc, eventsCh *EventChannel) error {

	// Init wait group to wait for all goroutines to return
	var wg sync.WaitGroup
	wg.Add(len(services))

	// Create error channel with a buffer of the number of services being deployed
	errCh := make(chan error, len(services))

	// Wait for all goroutines to complete before closing the error channel
	go func() {
		defer close(errCh)
		wg.Wait()
	}()

	// Run update function on all services concurrently
	for _, svc := range services {
		go updateFunc(ctx, svc, &wg, errCh, eventsCh)
	}

	// Wait for as many responses as the number of components or timeout,
	// return immediately on error.
	componentCount := 0
	for componentCount < len(services) {
		select {
		case err := <-errCh:
			componentCount++
			if err != nil {
				return err
			}
		case <-ctx.Done():
			return errors.New("timeout waiting for response")
		}
	}

	return nil
}

const (
	// PyFuncEnsemblerServiceEndpoint URL path for the endpoint, e.g "/"
	PyFuncEnsemblerServiceEndpoint string = "/ensemble"
	// PyFuncEnsemblerServicePort Port number the container listens to for requests
	PyFuncEnsemblerServicePort int = 8083
)
