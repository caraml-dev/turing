package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/api/client"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/api/resource"

	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/cluster/labeller"
	"github.com/caraml-dev/turing/api/turing/cluster/servicebuilder"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/imagebuilder"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/engines/experiment/manager"
	routerConfig "github.com/caraml-dev/turing/engines/router/missionctl/config"
)

// RouterDeploymentService handles the deployment of the Turing routers and the related components.
type RouterDeploymentService interface {
	DeployOrRollbackRouter(project *mlp.Project, router *models.Router, routerVersion *models.RouterVersion) error
	UndeployRouter(project *mlp.Project, router *models.Router) error
}
type routerDeploymentService struct {
	services *Services

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

const (
	// PyFuncEnsemblerServiceEndpoint URL path for the endpoint, e.g "/"
	PyFuncEnsemblerServiceEndpoint string = "/ensemble"
	// PyFuncEnsemblerServicePort Port number the container listens to for requests
	PyFuncEnsemblerServicePort int = 8083
)

// NewDeploymentService initialises a new endpoints service
func NewDeploymentService(
	cfg *config.Config,
	clusterControllers map[string]cluster.Controller,
	ensemblerServiceImageBuilder imagebuilder.ImageBuilder,
	services *Services,
) RouterDeploymentService {
	// Create cluster service builder
	sb := servicebuilder.NewClusterServiceBuilder(
		resource.Quantity(cfg.DeployConfig.MaxCPU),
		resource.Quantity(cfg.DeployConfig.MaxMemory),
		cfg.DeployConfig.MaxAllowedReplica,
	)

	return &routerDeploymentService{
		services:                     services,
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

// DeployOrRollbackRouter takes in the project, router and the version to be deployed
func (svc routerDeploymentService) DeployOrRollbackRouter(
	project *mlp.Project,
	router *models.Router,
	routerVersion *models.RouterVersion,
) error {
	// Get the router environment
	environment, err := svc.services.MLPService.GetEnvironment(router.EnvironmentName)
	if err != nil {
		return err
	}

	// Prepare router for deploy - update status to pending
	err = svc.updateRouterStatus(router, true)
	if err != nil {
		return err
	}

	eventsCh := NewEventChannel()
	defer eventsCh.Close()
	_ = svc.services.EventService.ClearEvents(int(router.ID))
	// Write events asynchronously
	go svc.writeDeploymentEvents(eventsCh, router, routerVersion.Version)

	eventsCh.Write(models.NewInfoEvent(
		models.EventStageDeployingDependencies,
		"starting deployment for router %s version %d", router.Name, routerVersion.Version))

	// Deploy the given router version
	endpoint, err := svc.deployRouterVersion(project, environment, routerVersion, eventsCh)

	// Start accumulating non-critical errors
	errorStrings := make([]string, 0)

	if err != nil {
		eventsCh.Write(models.NewErrorEvent(models.EventStageDeploymentFailed,
			"failed to deploy router %s version %d: %s",
			routerVersion.Router.Name, routerVersion.Version, err.Error()))
		// If the router was deleted in the midst of a pending deployment, an error will be thrown
		// by the thread due to its inability to update the absent router record in the db.
		// Check if the error in deployment was due to the router being deleted, if so, return.
		router, getRouterErr := svc.services.RoutersService.FindByID(router.ID)
		if getRouterErr != nil {
			return getRouterErr
		}
		if router == nil {
			return nil
		}

		eventsCh.Write(models.NewInfoEvent(models.EventStageRollback,
			"rolling back router %s version %d", router.Name, routerVersion.Version))
		// Save the error from the failed deployment
		errorStrings = append(errorStrings, err.Error())
		// Remove cluster resources from the failed deployment attempt
		err = svc.undeployRouterVersion(project, environment, routerVersion, eventsCh, true)
		if err != nil {
			errorStrings = append(errorStrings, err.Error())
			eventsCh.Write(models.NewErrorEvent(
				models.EventStageRollback,
				"failed to undeploy router %s version %d: %s",
				router.Name, routerVersion.Version, err.Error()),
			)
		}
		// Update router references
		err = svc.updateRouterReferences(router, routerVersion, endpoint)
		if err != nil {
			errorStrings = append(errorStrings, err.Error())
			eventsCh.Write(models.NewErrorEvent(
				models.EventStageRollback,
				"failed to update reference for router %s version %d: %s",
				router.Name, routerVersion.Version, err.Error()))
		}

		err = errors.New(strings.Join(errorStrings, ". "))
		return err
	}

	// Deployment successful - undeploy the current router version from the cluster
	if router.CurrRouterVersion != nil &&
		router.CurrRouterVersion.Status == models.RouterVersionStatusDeployed {
		currVersion, err := svc.services.RouterVersionsService.FindByID(router.CurrRouterVersion.ID)
		if err != nil {
			errorStrings = append(errorStrings, err.Error())
		} else {
			err = svc.undeployRouterVersion(project, environment, currVersion, eventsCh, false)
			if err != nil {
				errorStrings = append(errorStrings, err.Error())
			}
			eventsCh.Write(models.NewInfoEvent(models.EventStageUndeployingPreviousVersion,
				"successfully undeployed previously deployed version %d",
				router.CurrRouterVersion.Version))
		}
	}

	// Finally, update router references, status and endpoint
	err = svc.updateRouterReferences(router, routerVersion, endpoint)
	if err != nil {
		errorStrings = append(errorStrings, err.Error())
	}
	if len(errorStrings) > 0 {
		err = errors.New(strings.Join(errorStrings, ". "))
		eventsCh.Write(models.NewErrorEvent(models.EventStageUpdatingEndpoint,
			"failed to deploy router %s version %d: %s",
			routerVersion.Router.Name, routerVersion.Version, err.Error()))
		return err
	}
	eventsCh.Write(models.NewInfoEvent(models.EventStageDeploymentSuccess,
		"successfully deployed router %s version %d",
		router.Name, routerVersion.Version))
	return nil
}

func (svc routerDeploymentService) UndeployRouter(
	project *mlp.Project,
	router *models.Router,
) error {
	// Get the router environment
	environment, err := svc.services.MLPService.GetEnvironment(router.EnvironmentName)
	if err != nil {
		return err
	}

	// Start accumulating non-critical errors
	var errorStrings []string

	// Write events asynchronously
	eventsCh := NewEventChannel()
	defer eventsCh.Close()
	var version uint
	if router.CurrRouterVersion != nil {
		version = router.CurrRouterVersion.Version
	}
	go svc.writeDeploymentEvents(eventsCh, router, version)

	eventsCh.Write(models.NewInfoEvent(models.EventStageDeletingDependencies,
		"undeploying router %s", router.Name))

	// Clean up deployed as well as pending versions from the cluster
	routerVersions, err := svc.services.RouterVersionsService.ListRouterVersions(router.ID)
	if err != nil {
		return err
	}

	for _, routerVersion := range routerVersions {
		if routerVersion.Status == models.RouterVersionStatusPending ||
			routerVersion.Status == models.RouterVersionStatusDeployed {
			// Remove cluster resources
			if routerVersion.Status == models.RouterVersionStatusPending {
				err = svc.undeployRouterVersion(project, environment, routerVersion, eventsCh, true)
			} else if routerVersion.Status == models.RouterVersionStatusDeployed {
				err = svc.undeployRouterVersion(project, environment, routerVersion, eventsCh, false)
			}

			if err != nil {
				errorStrings = append(errorStrings, err.Error())
			}

			// Update the version's status
			versionID := routerVersion.ID
			routerVersion.Status = models.RouterVersionStatusUndeployed
			routerVersion, err = svc.services.RouterVersionsService.UpdateRouterVersion(routerVersion)
			if err != nil {
				errorStrings = append(errorStrings, err.Error())
				if router.CurrRouterVersion != nil &&
					router.CurrRouterVersion.ID == versionID {
					// There was an error manipulating the current version,
					// we cannot proceed updating the router properly
					errorStrings = append(errorStrings,
						"Error updating router's current configuration")
					err = errors.New(strings.Join(errorStrings, ". "))
					eventsCh.Write(models.NewErrorEvent(models.EventStageUndeploymentFailed,
						"failed to undeploy router %s: %s",
						routerVersion.Router.Name, err.Error()))
					return err
				}
			} else if router.CurrRouterVersion != nil &&
				router.CurrRouterVersion.ID == routerVersion.ID {
				// If current version, update the router's reference for subsequent updates
				router.CurrRouterVersion = routerVersion
			}
		}
	}
	// Clean up virtual service
	err = svc.deleteRouterEndpoint(project, environment,
		&models.RouterVersion{Router: router})
	if err != nil {
		errorStrings = append(errorStrings, err.Error())
	}

	// Update router's endpoint and status
	router.Endpoint = ""
	err = svc.updateRouterStatus(router, false)
	if err != nil {
		errorStrings = append(errorStrings, err.Error())
	}

	if len(errorStrings) > 0 {
		err = errors.New(strings.Join(errorStrings, ". "))
		eventsCh.Write(models.NewErrorEvent(models.EventStageUndeploymentFailed,
			"failed to undeploy router %s: %s",
			router.Name, err.Error()))
		return err
	}
	eventsCh.Write(models.NewInfoEvent(models.EventStageUndeploymentSuccess,
		"router %s undeployed", router.Name))
	return nil
}

func (svc routerDeploymentService) writeDeploymentEvents(
	eventsCh *EventChannel, router *models.Router, version uint) {
	for {
		event, done := eventsCh.Read()
		if done {
			return
		}
		event.SetRouter(router)
		event.SetVersion(version)
		_ = svc.services.EventService.Save(event)
	}
}

// deployRouterVersion attempts to deploy the given router version. Updates to the router
// version attributes are handled by this method, but the update of router attributes
// (current version reference, status, endpoint, etc.) are not in the scope of this method.
// This method returns the new router endpoint (if successful) and any error.
func (svc routerDeploymentService) deployRouterVersion(
	project *mlp.Project,
	environment *merlin.Environment,
	routerVersion *models.RouterVersion,
	eventsCh *EventChannel,
) (string, error) {
	var routerServiceAccountKey, enricherServiceAccountKey, ensemblerServiceAccountKey,
		expEngineServiceAccountKey string
	var experimentConfig json.RawMessage
	var err error

	if routerVersion.LogConfig.ResultLoggerType == models.BigQueryLogger {
		routerServiceAccountKey, err = svc.services.MLPService.GetSecret(
			models.ID(project.ID),
			routerVersion.LogConfig.BigQueryConfig.ServiceAccountSecret,
		)
		if err != nil {
			return "", svc.updateRouterVersionStatusToFailed(err, routerVersion)
		}
	}

	if routerVersion.Enricher != nil && routerVersion.Enricher.ServiceAccount != "" {
		enricherServiceAccountKey, err = svc.services.MLPService.GetSecret(
			models.ID(project.ID),
			routerVersion.Enricher.ServiceAccount,
		)
		if err != nil {
			return "", svc.updateRouterVersionStatusToFailed(err, routerVersion)
		}
	}

	if routerVersion.Ensembler != nil && routerVersion.Ensembler.Type == models.EnsemblerDockerType {
		if routerVersion.Ensembler.DockerConfig.ServiceAccount != "" {
			ensemblerServiceAccountKey, err = svc.services.MLPService.GetSecret(
				models.ID(project.ID),
				routerVersion.Ensembler.DockerConfig.ServiceAccount,
			)
			if err != nil {
				return "", svc.updateRouterVersionStatusToFailed(err, routerVersion)
			}
		}
	}

	if routerVersion.ExperimentEngine.Type != models.ExperimentEngineTypeNop {
		experimentConfig, err = svc.getExperimentConfig(routerVersion)
		if err != nil {
			return "", svc.updateRouterVersionStatusToFailed(err, routerVersion)
		}

		if routerVersion.ExperimentEngine.ServiceAccountKeyFilePath != nil {
			serviceAccountKey, err := svc.getLocalSecret(*routerVersion.ExperimentEngine.
				ServiceAccountKeyFilePath)
			if err != nil {
				return "", svc.updateRouterVersionStatusToFailed(err, routerVersion)
			}
			expEngineServiceAccountKey = *serviceAccountKey
		}
	}

	// Prepare to deploy router version - set version status to pending deployment
	if routerVersion.Status != models.RouterVersionStatusPending {
		routerVersion.Status = models.RouterVersionStatusPending
		_, err := svc.services.RouterVersionsService.UpdateRouterVersion(routerVersion)
		if err != nil {
			return "", err
		}
	}

	// Retrieve pyfunc ensembler if pyfunc ensembler is specified
	var pyfuncEnsembler *models.PyFuncEnsembler
	if routerVersion.Ensembler != nil && routerVersion.Ensembler.Type == models.EnsemblerPyFuncType {
		ensembler, err := svc.services.EnsemblersService.FindByID(
			*routerVersion.Ensembler.PyfuncConfig.EnsemblerID,
			EnsemblersFindByIDOptions{
				ProjectID: routerVersion.Ensembler.PyfuncConfig.ProjectID,
			})
		if err != nil {
			return "", fmt.Errorf("failed to find specified ensembler: %w", err)
		}

		castedEnsembler, ok := ensembler.(*models.PyFuncEnsembler)
		if !ok {
			return "", fmt.Errorf("failed to cast ensembler: %w", err)
		}
		pyfuncEnsembler = castedEnsembler
	}

	// Deploy the router version
	endpoint, err := svc.deployRouterVersionResources(
		project,
		environment,
		routerVersion,
		routerServiceAccountKey,
		enricherServiceAccountKey,
		ensemblerServiceAccountKey,
		expEngineServiceAccountKey,
		pyfuncEnsembler,
		experimentConfig,
		eventsCh,
	)

	if err != nil {
		err = svc.updateRouterVersionStatusToFailed(err, routerVersion)
		eventsCh.Write(models.NewErrorEvent(models.EventStageDeploymentFailed,
			"failed to deploy router %s version %d: %s",
			routerVersion.Router.Name, routerVersion.Version, err.Error()))
		return "", err
	}

	// Deploy succeeded - update version's status to deployed and return endpoint
	routerVersion.Status = models.RouterVersionStatusDeployed
	_, err = svc.services.RouterVersionsService.UpdateRouterVersion(routerVersion)
	return endpoint, err
}

// updateRouterReferences updates the current version ref, the endpoint
// of the router and its status
func (svc routerDeploymentService) updateRouterReferences(
	router *models.Router,
	routerVersion *models.RouterVersion,
	endpoint string,
) error {
	// Update router version reference if:
	// a) No current version
	// b) Current version and new version have both failed deployment
	// c) New version is successfully deployed
	if router.CurrRouterVersion == nil ||
		(router.CurrRouterVersion.Status == models.RouterVersionStatusFailed &&
			routerVersion.Status == models.RouterVersionStatusFailed) ||
		routerVersion.Status == models.RouterVersionStatusDeployed {
		// If current version exists and it is valid, change its status to undeployed
		if router.CurrRouterVersion != nil &&
			router.CurrRouterVersion.Status == models.RouterVersionStatusDeployed &&
			// If the current version has been re-deployed, don't set its status to undeployed
			router.CurrRouterVersion.ID != routerVersion.ID {
			router.CurrRouterVersion.Status = models.RouterVersionStatusUndeployed
			_, err := svc.services.RouterVersionsService.UpdateRouterVersion(router.CurrRouterVersion)
			if err != nil {
				return err
			}
		}

		// Update router's current version
		router.SetCurrRouterVersion(routerVersion)

		// Update router endpoint
		router.Endpoint = endpoint
	}

	// Update router status and save
	return svc.updateRouterStatus(router, false)
}

// updateRouterStatus sets the status of the router, based on beginDeployment flag
// and its current version reference.
func (svc routerDeploymentService) updateRouterStatus(
	router *models.Router,
	beginDeployment bool,
) error {
	// Update status
	if beginDeployment {
		// If beginDeployment is set, update status to Pending
		router.Status = models.RouterStatusPending
	} else if router.CurrRouterVersion == nil {
		// No current deployment, set status to failed
		router.Status = models.RouterStatusFailed
	} else {
		// Copy the current router version's status (this could be deployed / undeployed / failed)
		router.Status = models.RouterStatus(router.CurrRouterVersion.Status)
	}

	// Save the router
	_, err := svc.services.RoutersService.Save(router)
	return err
}

// Updates the given router version status to failed, and persists the result. Returns the
// original error concatenated with any errors that arose as a result of saving the router version.
func (svc routerDeploymentService) updateRouterVersionStatusToFailed(
	err error, routerVersion *models.RouterVersion) error {
	errorsStrings := []string{err.Error()}
	routerVersion.Status = models.RouterVersionStatusFailed
	routerVersion.Error = err.Error()
	_, err = svc.services.RouterVersionsService.UpdateRouterVersion(routerVersion)
	if err != nil {
		errorsStrings = append(errorsStrings, err.Error())
	}
	return errors.New(strings.Join(errorsStrings, ". "))
}

func (svc *routerDeploymentService) deployRouterVersionResources(
	project *mlp.Project,
	environment *merlin.Environment,
	routerVersion *models.RouterVersion,
	routerServiceAccountKey string,
	enricherServiceAccountKey string,
	ensemblerServiceAccountKey string,
	expEngineServiceAccountKey string,
	pyfuncEnsembler *models.PyFuncEnsembler,
	experimentConfig json.RawMessage,
	eventsCh *EventChannel,
) (string, error) {
	var endpoint string

	// If pyfunc ensembler is specified as an ensembler service, build/retrieve its image
	if pyfuncEnsembler != nil {
		err := svc.buildEnsemblerServiceImage(pyfuncEnsembler, project, routerVersion, eventsCh)
		if err != nil {
			return endpoint, err
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), svc.deploymentTimeout)
	defer cancel()

	// Get the cluster controller
	controller, err := svc.getClusterControllerByEnvironment(environment.Name)
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
	secret := svc.svcBuilder.NewSecret(
		routerVersion,
		project,
		routerServiceAccountKey,
		enricherServiceAccountKey,
		ensemblerServiceAccountKey,
		expEngineServiceAccountKey,
	)
	err = createSecret(ctx, controller, secret)
	if err != nil {
		eventsCh.Write(models.NewErrorEvent(
			models.EventStageDeployingDependencies, "failed to create secret: %s", err.Error()))
		return endpoint, err
	}
	secretName := secret.Name

	// Construct service objects for each of the components and deploy
	services, err := svc.createServices(
		routerVersion, project, svc.environmentType, secretName, experimentConfig,
		svc.routerDefaults, svc.sentryEnabled, svc.sentryDSN,
		svc.knativeServiceConfig.QueueProxyResourcePercentage,
		svc.knativeServiceConfig.UserContainerLimitRequestFactor,
	)
	if err != nil {
		return endpoint, err
	}

	// Deploy fluentd if enabled
	if routerVersion.LogConfig.ResultLoggerType == models.BigQueryLogger {
		fluentdService := svc.svcBuilder.NewFluentdService(routerVersion, project,
			secretName, svc.routerDefaults.FluentdConfig)
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

	err = createKnServices(ctx, controller, services, eventsCh)
	if err != nil {
		return endpoint, err
	}

	// Get the router's external endpoint
	routerSvcName := svc.svcBuilder.GetRouterServiceName(routerVersion)
	endpoint = controller.GetKnativeServiceURL(ctx, routerSvcName, project.Name)

	// Deploy or update the virtual service
	eventsCh.Write(models.NewInfoEvent(models.EventStageUpdatingEndpoint, "updating router endpoint"))
	routerEndpoint, err := svc.svcBuilder.NewRouterEndpoint(routerVersion, project, svc.environmentType, endpoint)
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

	// only base endpoint is returned, models/router.go will unmarshal with /v1/predict for http routers
	if routerVersion.Protocol == routerConfig.UPI {
		return routerEndpoint.Endpoint + ":80", err
	}
	return fmt.Sprintf("http://%s", routerEndpoint.Endpoint), err
}

// undeployRouterVersion removes the deployed router, if exists. Else, an error is returned.
func (svc *routerDeploymentService) undeployRouterVersion(
	project *mlp.Project,
	environment *merlin.Environment,
	routerVersion *models.RouterVersion,
	eventsCh *EventChannel,
	isCleanUp bool,
) error {
	ctx, cancel := context.WithTimeout(context.Background(), svc.deploymentDeletionTimeout)
	defer cancel()
	// Get the cluster controller
	controller, err := svc.getClusterControllerByEnvironment(environment.Name)
	if err != nil {
		return err
	}

	// Delete secret
	eventsCh.Write(models.NewInfoEvent(models.EventStageDeletingDependencies, "deleting secrets"))
	secret := svc.svcBuilder.NewSecret(routerVersion, project, "", "", "", "")
	err = deleteSecret(controller, secret, isCleanUp)
	if err != nil {
		return err
	}

	// Construct service objects for each of the components to be deleted
	services, err := svc.createServices(
		routerVersion, project, svc.environmentType, "", nil,
		svc.routerDefaults, svc.sentryEnabled, svc.sentryDSN,
		svc.knativeServiceConfig.QueueProxyResourcePercentage,
		svc.knativeServiceConfig.UserContainerLimitRequestFactor,
	)
	if err != nil {
		return err
	}

	var errs []string
	// Delete fluentd if required
	if routerVersion.LogConfig.ResultLoggerType == models.BigQueryLogger {
		fluentdService := svc.svcBuilder.NewFluentdService(routerVersion,
			project, "", svc.routerDefaults.FluentdConfig)
		err = deleteK8sService(controller, fluentdService, isCleanUp)
		if err != nil {
			errs = append(errs, err.Error())
		}
		err = deletePVC(controller, project.Name, fluentdService.PersistentVolumeClaim, isCleanUp)
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

	// TODO: Delete this block once all existing routers no longer use the old plugins server service to deploy plugins
	// Delete experiment engine plugins server
	if routerVersion.ExperimentEngine.PluginConfig != nil {
		pluginsServerSvc := servicebuilder.NewPluginsServerService(routerVersion, project)
		err = deleteK8sService(controller, pluginsServerSvc, isCleanUp)
		if err == nil {
			eventsCh.Write(models.NewInfoEvent(
				models.EventStageDeletingDependencies, "successfully deleted plugins server"))
		}
	}

	// Delete all components
	err = deleteKnServices(ctx, controller, services, eventsCh, isCleanUp)
	if err != nil {
		errs = append(errs, err.Error())
	}
	if len(errs) != 0 {
		return errors.New(strings.Join(errs, ". "))
	}

	return nil
}

func (svc *routerDeploymentService) deleteRouterEndpoint(
	project *mlp.Project,
	environment *merlin.Environment,
	routerVersion *models.RouterVersion,
) error {
	// Get the cluster controller
	controller, err := svc.getClusterControllerByEnvironment(environment.Name)
	if err != nil {
		return err
	}

	routerEndpointName := fmt.Sprintf("%s-turing-router", routerVersion.Router.Name)
	return controller.DeleteIstioVirtualService(context.Background(), routerEndpointName, project.Name)
}

func (svc *routerDeploymentService) getClusterControllerByEnvironment(
	environment string,
) (cluster.Controller, error) {
	controller, ok := svc.clusterControllers[environment]
	if !ok {
		return nil, errors.New("Deployment environment not supported")
	}
	return controller, nil
}

func (svc *routerDeploymentService) createServices(
	routerVersion *models.RouterVersion,
	project *mlp.Project,
	envType string,
	secretName string,
	experimentConfig json.RawMessage,
	routerDefaults *config.RouterDefaults,
	sentryEnabled bool,
	sentryDSN string,
	knativeQueueProxyResourcePercentage int,
	userContainerLimitRequestFactor float64,
) ([]*cluster.KnativeService, error) {
	services := []*cluster.KnativeService{}

	// Enricher
	if routerVersion.Enricher != nil {
		enricherSvc, err := svc.svcBuilder.NewEnricherService(
			routerVersion,
			project,
			envType,
			secretName,
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
		ensemblerSvc, err := svc.svcBuilder.NewEnsemblerService(
			routerVersion,
			project,
			envType,
			secretName,
			knativeQueueProxyResourcePercentage,
			userContainerLimitRequestFactor,
		)
		if err != nil {
			return services, err
		}
		services = append(services, ensemblerSvc)
	}

	// Router
	routerService, err := svc.svcBuilder.NewRouterService(
		routerVersion,
		project,
		envType,
		secretName,
		experimentConfig,
		routerDefaults,
		sentryEnabled,
		sentryDSN,
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
func (svc *routerDeploymentService) buildEnsemblerServiceImage(
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
		BaseImageRefTag: ensembler.PythonVersion,
	}
	eventsCh.Write(
		models.NewInfoEvent(
			models.EventStageDeployingDependencies,
			"building/retrieving pyfunc ensembler with project_id: %d and ensembler_id: %d",
			*routerVersion.Ensembler.PyfuncConfig.ProjectID,
			*routerVersion.Ensembler.PyfuncConfig.EnsemblerID,
		),
	)
	imageRef, imageBuildErr := svc.ensemblerServiceImageBuilder.BuildImage(request)
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
		Image:             imageRef,
		ResourceRequest:   routerVersion.Ensembler.PyfuncConfig.ResourceRequest,
		AutoscalingPolicy: routerVersion.Ensembler.PyfuncConfig.AutoscalingPolicy,
		Timeout:           routerVersion.Ensembler.PyfuncConfig.Timeout,
		Endpoint:          PyFuncEnsemblerServiceEndpoint,
		Port:              PyFuncEnsemblerServicePort,
		Env:               routerVersion.Ensembler.PyfuncConfig.Env,
		ServiceAccount:    "",
	}

	return nil
}

func (svc routerDeploymentService) getExperimentConfig(routerVersion *models.RouterVersion) (json.RawMessage, error) {
	var experimentConfig json.RawMessage
	experimentConfig = routerVersion.ExperimentEngine.Config
	isClientSelectionEnabled, err := svc.services.ExperimentsService.IsClientSelectionEnabled(
		routerVersion.ExperimentEngine.Type,
	)
	if err != nil {
		return nil, svc.updateRouterVersionStatusToFailed(err, routerVersion)
	}
	if isClientSelectionEnabled {
		// Convert the config to the standard type
		standardExperimentConfig, err := manager.ParseStandardExperimentConfig(experimentConfig)
		if err != nil {
			return nil, svc.updateRouterVersionStatusToFailed(err, routerVersion)
		}
		// If passkey has been set, decrypt it
		if standardExperimentConfig.Client.Passkey != "" {
			standardExperimentConfig.Client.Passkey, err =
				svc.services.CryptoService.Decrypt(standardExperimentConfig.Client.Passkey)
			if err != nil {
				return nil, svc.updateRouterVersionStatusToFailed(err, routerVersion)
			}

			experimentConfig, err = json.Marshal(standardExperimentConfig)
			if err != nil {
				return nil, svc.updateRouterVersionStatusToFailed(err, routerVersion)
			}
		}
	}

	// Get the deployable Router Config for the experiment
	experimentConfig, err = svc.services.ExperimentsService.GetExperimentRunnerConfig(
		routerVersion.ExperimentEngine.Type,
		experimentConfig,
	)
	if err != nil {
		return nil, svc.updateRouterVersionStatusToFailed(err, routerVersion)
	}

	return experimentConfig, nil
}

// K8s Secret /////////////////////////////////////////////////////////////////

func (svc *routerDeploymentService) getLocalSecret(
	serviceAccountKeyFilePath string,
) (*string, error) {
	byteValue, err := os.ReadFile(serviceAccountKeyFilePath)
	if err != nil {
		return nil, err
	}

	serviceAccountKey := string(byteValue)
	return &serviceAccountKey, nil
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
func deleteSecret(controller cluster.Controller, secret *cluster.Secret, isCleanUp bool) error {
	return controller.DeleteSecret(context.Background(), secret.Name, secret.Namespace, isCleanUp)
}

// PVC ////////////////////////////////////////////////////////////////////////

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
	isCleanUp bool,
) error {
	return controller.DeletePersistentVolumeClaim(context.Background(), pvc.Name, namespace, isCleanUp)
}

// Knative Services ///////////////////////////////////////////////////////////

// createKnServices deploys all services simulateneously and waits for all of them to
// be ready. This includes the enricher (if enabled), ensembler (if enabled) and router.
// Note: The enricher and ensembler have no health checks, so they will be ready immediately.
func createKnServices(
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
			err := controller.ApplyConfigMap(ctx, svc.Namespace, svc.ConfigMap)
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

// updateKnServices is a helper method for deployment / deletion of services that runs the
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

// deleteKnServices simultaneously issues a delete call to all services and waits
// until deletion timeout, for a response
func deleteKnServices(
	ctx context.Context,
	controller cluster.Controller,
	services []*cluster.KnativeService,
	eventsCh *EventChannel,
	isCleanUp bool,
) error {
	// Define delete function
	deleteFunc := func(_ context.Context,
		svc *cluster.KnativeService,
		wg *sync.WaitGroup,
		errCh chan<- error,
		eventsCh *EventChannel,
	) {
		defer wg.Done()
		var err error
		eventsCh.Write(models.NewInfoEvent(
			models.EventStageUndeployingServices, "deleting service %s", svc.Name))
		if svc.ConfigMap != nil {
			err = controller.DeleteConfigMap(context.Background(), svc.ConfigMap.Name, svc.Namespace, isCleanUp)
			if err != nil {
				err = errors.Wrapf(err, "Failed to delete config map %s", svc.ConfigMap.Name)
				eventsCh.Write(models.NewErrorEvent(
					models.EventStageUndeployingServices, "failed to delete service %s: %s", svc.Name, err.Error()))
				errCh <- err
			}
		}
		err = controller.DeleteKnativeService(context.Background(), svc.Name, svc.Namespace, isCleanUp)
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

// K8s Services ///////////////////////////////////////////////////////////////

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
	isCleanUp bool,
) error {
	err := controller.DeleteKubernetesDeployment(context.Background(), service.Name, service.Namespace, isCleanUp)
	if err != nil {
		return err
	}
	return controller.DeleteKubernetesService(context.Background(), service.Name, service.Namespace, isCleanUp)
}
