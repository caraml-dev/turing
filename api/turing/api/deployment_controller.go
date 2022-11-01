package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/api/client"

	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
	"github.com/caraml-dev/turing/engines/experiment/manager"
)

// RouterDeploymentController handles the deployment of routers
type RouterDeploymentController struct {
	BaseController
}

// deployOrRollbackRouter takes in the project, router and the version to be deployed
func (c RouterDeploymentController) deployOrRollbackRouter(
	project *mlp.Project,
	router *models.Router,
	routerVersion *models.RouterVersion,
) error {
	// Get the router environment
	environment, err := c.MLPService.GetEnvironment(router.EnvironmentName)
	if err != nil {
		return err
	}

	// Prepare router for deploy - update status to pending
	err = c.updateRouterStatus(router, true)
	if err != nil {
		return err
	}

	eventsCh := service.NewEventChannel()
	defer eventsCh.Close()
	_ = c.EventService.ClearEvents(int(router.ID))
	// Write events asynchronously
	go c.writeDeploymentEvents(eventsCh, router, routerVersion.Version)

	eventsCh.Write(models.NewInfoEvent(
		models.EventStageDeployingDependencies,
		"starting deployment for router %s version %d", router.Name, routerVersion.Version))

	// Deploy the given router version
	endpoint, err := c.deployRouterVersion(project, environment, routerVersion, eventsCh)

	// Start accumulating non-critical errors
	errorStrings := make([]string, 0)

	if err != nil {
		eventsCh.Write(models.NewErrorEvent(models.EventStageDeploymentFailed,
			"failed to deploy router %s version %d: %s",
			routerVersion.Router.Name, routerVersion.Version, err.Error()))
		// If the router was deleted in the midst of a pending deployment, an error will be thrown
		// by the thread due to its inability to update the absent router record in the db.
		// Check if the error in deployment was due to the router being deleted, if so, return.
		router, getRouterErr := c.RoutersService.FindByID(router.ID)
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
		err = c.DeploymentService.UndeployRouterVersion(project, environment, routerVersion, eventsCh, true)
		if err != nil {
			errorStrings = append(errorStrings, err.Error())
			eventsCh.Write(models.NewErrorEvent(
				models.EventStageRollback,
				"failed to undeploy router %s version %d: %s",
				router.Name, routerVersion.Version, err.Error()),
			)
		}
		// Update router references
		err = c.updateRouterReferences(router, routerVersion, endpoint)
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
		currVersion, err := c.RouterVersionsService.FindByID(router.CurrRouterVersion.ID)
		if err != nil {
			errorStrings = append(errorStrings, err.Error())
		} else {
			err = c.DeploymentService.UndeployRouterVersion(project, environment, currVersion, eventsCh, false)
			if err != nil {
				errorStrings = append(errorStrings, err.Error())
			}
			eventsCh.Write(models.NewInfoEvent(models.EventStageUndeployingPreviousVersion,
				"successfully undeployed previously deployed version %d",
				router.CurrRouterVersion.Version))
		}
	}

	// Finally, update router references, status and endpoint
	err = c.updateRouterReferences(router, routerVersion, endpoint)
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

func (c RouterDeploymentController) writeDeploymentEvents(
	eventsCh *service.EventChannel, router *models.Router, version uint) {
	for {
		event, done := eventsCh.Read()
		if done {
			return
		}
		event.SetRouter(router)
		event.SetVersion(version)
		_ = c.EventService.Save(event)
	}
}

// deployRouterVersion attempts to deploy the given router version. Updates to the router
// version attributes are handled by this method, but the update of router attributes
// (current version reference, status, endpoint, etc.) are not in the scope of this method.
// This method returns the new router endpoint (if successful) and any error.
func (c RouterDeploymentController) deployRouterVersion(
	project *mlp.Project,
	environment *merlin.Environment,
	routerVersion *models.RouterVersion,
	eventsCh *service.EventChannel,
) (string, error) {
	var routerServiceAccountKey, enricherServiceAccountKey, ensemblerServiceAccountKey string
	var experimentConfig json.RawMessage
	var err error

	if routerVersion.LogConfig.ResultLoggerType == models.BigQueryLogger {
		routerServiceAccountKey, err = c.MLPService.GetSecret(
			models.ID(project.Id),
			routerVersion.LogConfig.BigQueryConfig.ServiceAccountSecret,
		)
		if err != nil {
			return "", c.updateRouterVersionStatusToFailed(err, routerVersion)
		}
	}

	if routerVersion.Enricher != nil && routerVersion.Enricher.ServiceAccount != "" {
		enricherServiceAccountKey, err = c.MLPService.GetSecret(
			models.ID(project.Id),
			routerVersion.Enricher.ServiceAccount,
		)
		if err != nil {
			return "", c.updateRouterVersionStatusToFailed(err, routerVersion)
		}
	}

	if routerVersion.Ensembler != nil && routerVersion.Ensembler.Type == models.EnsemblerDockerType {
		if routerVersion.Ensembler.DockerConfig.ServiceAccount != "" {
			ensemblerServiceAccountKey, err = c.MLPService.GetSecret(
				models.ID(project.Id),
				routerVersion.Ensembler.DockerConfig.ServiceAccount,
			)
			if err != nil {
				return "", c.updateRouterVersionStatusToFailed(err, routerVersion)
			}
		}
	}

	expSvc := c.BaseController.AppContext.ExperimentsService
	if routerVersion.ExperimentEngine.Type != models.ExperimentEngineTypeNop {
		experimentConfig = routerVersion.ExperimentEngine.Config
		isClientSelectionEnabled, err := expSvc.IsClientSelectionEnabled(routerVersion.ExperimentEngine.Type)
		if err != nil {
			return "", c.updateRouterVersionStatusToFailed(err, routerVersion)
		}
		if isClientSelectionEnabled {
			// Convert the config to the standard type
			standardExperimentConfig, err := manager.ParseStandardExperimentConfig(experimentConfig)
			if err != nil {
				return "", c.updateRouterVersionStatusToFailed(err, routerVersion)
			}
			// If passkey has been set, decrypt it
			if standardExperimentConfig.Client.Passkey != "" {
				standardExperimentConfig.Client.Passkey, err =
					c.CryptoService.Decrypt(standardExperimentConfig.Client.Passkey)
				if err != nil {
					return "", c.updateRouterVersionStatusToFailed(err, routerVersion)
				}

				experimentConfig, err = json.Marshal(standardExperimentConfig)
				if err != nil {
					return "", c.updateRouterVersionStatusToFailed(err, routerVersion)
				}
			}
		}

		// Get the deployable Router Config for the experiment
		experimentConfig, err = c.ExperimentsService.GetExperimentRunnerConfig(
			routerVersion.ExperimentEngine.Type,
			experimentConfig,
		)
		if err != nil {
			return "", c.updateRouterVersionStatusToFailed(err, routerVersion)
		}
	}

	// Prepare to deploy router version - set version status to pending deployment
	if routerVersion.Status != models.RouterVersionStatusPending {
		routerVersion.Status = models.RouterVersionStatusPending
		_, err := c.RouterVersionsService.Save(routerVersion)
		if err != nil {
			return "", err
		}
	}

	// Retrieve pyfunc ensembler if pyfunc ensembler is specified
	var pyfuncEnsembler *models.PyFuncEnsembler
	if routerVersion.Ensembler != nil && routerVersion.Ensembler.Type == models.EnsemblerPyFuncType {
		ensembler, err := c.EnsemblersService.FindByID(
			*routerVersion.Ensembler.PyfuncConfig.EnsemblerID,
			service.EnsemblersFindByIDOptions{
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
	endpoint, err := c.DeploymentService.DeployRouterVersion(
		project,
		environment,
		routerVersion,
		routerServiceAccountKey,
		enricherServiceAccountKey,
		ensemblerServiceAccountKey,
		pyfuncEnsembler,
		experimentConfig,
		eventsCh,
	)

	if err != nil {
		err = c.updateRouterVersionStatusToFailed(err, routerVersion)
		eventsCh.Write(models.NewErrorEvent(models.EventStageDeploymentFailed,
			"failed to deploy router %s version %d: %s",
			routerVersion.Router.Name, routerVersion.Version, err.Error()))
		return "", err
	}

	// Deploy succeeded - update version's status to deployed and return endpoint
	routerVersion.Status = models.RouterVersionStatusDeployed
	_, err = c.RouterVersionsService.Save(routerVersion)
	return endpoint, err
}

func (c RouterDeploymentController) undeployRouter(
	project *mlp.Project,
	router *models.Router,
) error {
	// Get the router environment
	environment, err := c.MLPService.GetEnvironment(router.EnvironmentName)
	if err != nil {
		return err
	}

	// Start accumulating non-critical errors
	var errorStrings []string

	// Write events asynchronously
	eventsCh := service.NewEventChannel()
	defer eventsCh.Close()
	var version uint
	if router.CurrRouterVersion != nil {
		version = router.CurrRouterVersion.Version
	}
	go c.writeDeploymentEvents(eventsCh, router, version)

	eventsCh.Write(models.NewInfoEvent(models.EventStageDeletingDependencies,
		"undeploying router %s", router.Name))

	// Clean up deployed as well as pending versions from the cluster
	routerVersions, err := c.RouterVersionsService.ListRouterVersions(router.ID)
	if err != nil {
		return err
	}

	for _, routerVersion := range routerVersions {
		if routerVersion.Status == models.RouterVersionStatusPending ||
			routerVersion.Status == models.RouterVersionStatusDeployed {
			// Remove cluster resources
			if routerVersion.Status == models.RouterVersionStatusPending {
				err = c.DeploymentService.UndeployRouterVersion(project, environment, routerVersion, eventsCh, true)
			} else if routerVersion.Status == models.RouterVersionStatusDeployed {
				err = c.DeploymentService.UndeployRouterVersion(project, environment, routerVersion, eventsCh, false)
			}

			if err != nil {
				errorStrings = append(errorStrings, err.Error())
			}

			// Update the version's status
			versionID := routerVersion.ID
			routerVersion.Status = models.RouterVersionStatusUndeployed
			routerVersion, err = c.RouterVersionsService.Save(routerVersion)
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
	err = c.DeploymentService.DeleteRouterEndpoint(project, environment,
		&models.RouterVersion{Router: router})
	if err != nil {
		errorStrings = append(errorStrings, err.Error())
	}

	// Update router's endpoint and status
	router.Endpoint = ""
	err = c.updateRouterStatus(router, false)
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

// updateRouterReferences updates the current version ref, the endpoint
// of the router and its status
func (c RouterDeploymentController) updateRouterReferences(
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
			_, err := c.RouterVersionsService.Save(router.CurrRouterVersion)
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
	return c.updateRouterStatus(router, false)
}

// updateRouterStatus sets the status of the router, based on beginDeployment flag
// and its current version reference.
func (c RouterDeploymentController) updateRouterStatus(
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
	_, err := c.RoutersService.Save(router)
	return err
}

// Updates the given router version status to failed, and persists the result. Returns the
// original error concatenated with any errors that arose as a result of saving the router version.
func (c RouterDeploymentController) updateRouterVersionStatusToFailed(
	err error, routerVersion *models.RouterVersion) error {
	errorsStrings := []string{err.Error()}
	routerVersion.Status = models.RouterVersionStatusFailed
	routerVersion.Error = err.Error()
	_, err = c.RouterVersionsService.Save(routerVersion)
	if err != nil {
		errorsStrings = append(errorsStrings, err.Error())
	}
	return errors.New(strings.Join(errorsStrings, ". "))
}
