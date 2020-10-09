package api

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/gojek/turing/api/turing/utils"

	merlin "github.com/gojek/merlin/client"
	mlp "github.com/gojek/mlp/client"
	"github.com/gojek/turing/api/turing/models"
)

// routerDeploymentController handles the deployment of routers
type routerDeploymentController struct {
	*baseController
}

// deployOrRollbackRouter takes in the project, router and the version to be deployed
func (c *routerDeploymentController) deployOrRollbackRouter(
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

	eventsCh := utils.NewEventChannel()
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
		err = c.DeploymentService.UndeployRouterVersion(project, environment, routerVersion, eventsCh)
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
			err = c.DeploymentService.UndeployRouterVersion(project, environment, currVersion, eventsCh)
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

func (c *routerDeploymentController) writeDeploymentEvents(
	eventsCh *utils.EventChannel, router *models.Router, version uint) {
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
func (c *routerDeploymentController) deployRouterVersion(
	project *mlp.Project,
	environment *merlin.Environment,
	routerVersion *models.RouterVersion,
	eventsCh *utils.EventChannel,
) (string, error) {
	var routerServiceAccountKey, enricherServiceAccountKey, ensemblerServiceAccountKey, experimentPasskey string
	var experimentConfig json.RawMessage
	var err error

	if routerVersion.LogConfig.ResultLoggerType == models.BigQueryLogger {
		routerServiceAccountKey, err = c.MLPService.GetSecret(
			int(project.Id),
			routerVersion.LogConfig.BigQueryConfig.ServiceAccountSecret,
		)
		if err != nil {
			return "", c.updateRouterVersionStatusToFailed(err, routerVersion)
		}
	}

	if routerVersion.Enricher != nil && routerVersion.Enricher.ServiceAccount != "" {
		enricherServiceAccountKey, err = c.MLPService.GetSecret(
			int(project.Id),
			routerVersion.Enricher.ServiceAccount,
		)
		if err != nil {
			return "", c.updateRouterVersionStatusToFailed(err, routerVersion)
		}
	}

	if routerVersion.Ensembler != nil && routerVersion.Ensembler.Type == models.EnsemblerDockerType {
		if routerVersion.Ensembler.DockerConfig.ServiceAccount != "" {
			ensemblerServiceAccountKey, err = c.MLPService.GetSecret(
				int(project.Id),
				routerVersion.Ensembler.DockerConfig.ServiceAccount,
			)
			if err != nil {
				return "", c.updateRouterVersionStatusToFailed(err, routerVersion)
			}
		}
	}

	if routerVersion.ExperimentEngine.Type != models.ExperimentEngineTypeNop {
		// If passkey has been set, decrypt it
		if routerVersion.ExperimentEngine.Config.Client.Passkey != "" {
			experimentPasskey, err = c.CryptoService.Decrypt(routerVersion.ExperimentEngine.Config.Client.Passkey)
			if err != nil {
				return "", c.updateRouterVersionStatusToFailed(err, routerVersion)
			}
		}
		// Get the deployable Router Config for the experiment
		experimentConfig, err = c.ExperimentsService.GetExperimentRunnerConfig(
			string(routerVersion.ExperimentEngine.Type),
			routerVersion.ExperimentEngine.Config,
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

	// Deploy the router version
	endpoint, err := c.DeploymentService.DeployRouterVersion(
		project,
		environment,
		routerVersion,
		routerServiceAccountKey,
		enricherServiceAccountKey,
		ensemblerServiceAccountKey,
		experimentConfig,
		experimentPasskey,
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

func (c *routerDeploymentController) undeployRouter(
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
	eventsCh := utils.NewEventChannel()
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
			err = c.DeploymentService.UndeployRouterVersion(project, environment, routerVersion, eventsCh)
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
func (c *routerDeploymentController) updateRouterReferences(
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
func (c *routerDeploymentController) updateRouterStatus(
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
func (c *routerDeploymentController) updateRouterVersionStatusToFailed(
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
