package batchensembling

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	apisparkv1beta2 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	apicorev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/caraml-dev/turing/api/turing/batch"
	"github.com/caraml-dev/turing/api/turing/cluster"
	"github.com/caraml-dev/turing/api/turing/config"
	"github.com/caraml-dev/turing/api/turing/log"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
)

// SparkApplicationState is the state of the spark application
type SparkApplicationState int

const (
	ensemblingJobApplicationPath = "local:///home/spark/batch-ensembler/main.py"
	// SparkApplicationStateRunning is when the spark application is still running
	SparkApplicationStateRunning = SparkApplicationState(iota)
	// SparkApplicationStateCompleted is when the spark application has completed its run
	SparkApplicationStateCompleted
	// SparkApplicationStateFailed is when the spark application has failed its run
	SparkApplicationStateFailed
	// SparkApplicationStateUnknown is when the spark application state is unknown
	SparkApplicationStateUnknown
)

// CreateEnsemblingJobRequest is a request to run an ensembling job on Kubernetes.
type CreateEnsemblingJobRequest struct {
	EnsemblingJob *models.EnsemblingJob
	Labels        map[string]string
	ImageRef      string
	Namespace     string
}

// EnsemblingController is an interface that exposes the batch ensembling kubernetes controller.
type EnsemblingController interface {
	Create(request *CreateEnsemblingJobRequest) error
	Delete(namespace string, ensemblingJob *models.EnsemblingJob) error
	GetStatus(namespace string, ensemblingJob *models.EnsemblingJob) (SparkApplicationState, error)
}

type ensemblingController struct {
	clusterController cluster.Controller
	mlpService        service.MLPService
	sparkInfraConfig  *config.SparkAppConfig
}

// NewBatchEnsemblingController creates a new batch ensembling controller
func NewBatchEnsemblingController(
	clusterController cluster.Controller,
	mlpService service.MLPService,
	sparkInfraConfig *config.SparkAppConfig,
) EnsemblingController {
	return &ensemblingController{
		clusterController: clusterController,
		mlpService:        mlpService,
		sparkInfraConfig:  sparkInfraConfig,
	}
}

func (c *ensemblingController) Delete(namespace string, ensemblingJob *models.EnsemblingJob) error {
	sa, err := c.clusterController.GetSparkApplication(context.Background(), namespace, ensemblingJob.Name)
	if err != nil {
		c.cleanup(ensemblingJob.Name, namespace)
		// Not found, we do not consider this as an error because its just no further
		// action required on this part.
		return nil
	}
	c.cleanup(ensemblingJob.Name, namespace)
	return c.clusterController.DeleteSparkApplication(context.Background(), namespace, sa.Name)
}

func (c *ensemblingController) GetStatus(
	namespace string,
	ensemblingJob *models.EnsemblingJob,
) (SparkApplicationState, error) {
	sa, err := c.clusterController.GetSparkApplication(context.Background(), namespace, ensemblingJob.Name)
	if err != nil {
		return SparkApplicationStateUnknown, fmt.Errorf("failed to retrieve spark application %v", err)
	}

	state := sa.Status.AppState.State
	if state == apisparkv1beta2.CompletedState {
		return SparkApplicationStateCompleted, nil
	}

	if state == apisparkv1beta2.FailedState {
		return SparkApplicationStateFailed, nil
	}

	if state == apisparkv1beta2.UnknownState {
		return SparkApplicationStateUnknown, nil
	}

	return SparkApplicationStateRunning, nil
}

func (c *ensemblingController) Create(request *CreateEnsemblingJobRequest) error {
	var err error
	defer func() {
		if err != nil {
			c.cleanup(request.EnsemblingJob.Name, request.Namespace)
		}
	}()

	err = c.clusterController.CreateNamespace(context.Background(), request.Namespace)
	if errors.Is(err, cluster.ErrNamespaceAlreadyExists) {
		// This error is ok to ignore because we just need the namespace.
		err = nil
	}

	if err != nil {
		return fmt.Errorf("failed creating namespace %s: %v", request.Namespace, err)
	}

	serviceAccount, err := c.createSparkDriverAuthorization(request.Namespace, request.Labels)
	if err != nil {
		return fmt.Errorf(
			"failed creating spark driver authorization in namespace %s: %v",
			request.Namespace,
			err,
		)
	}

	secretString, err := c.mlpService.GetSecret(
		request.EnsemblingJob.ProjectID,
		request.EnsemblingJob.InfraConfig.GetServiceAccountName(),
	)
	if err != nil {
		return fmt.Errorf(
			"service account %s is not found within %s project: %s",
			request.EnsemblingJob.InfraConfig.GetServiceAccountName(),
			request.Namespace,
			err,
		)
	}

	err = c.createSecret(request, secretString)
	if err != nil {
		return fmt.Errorf(
			"failed creating secret for job %s in namespace %s: %v",
			request.EnsemblingJob.Name,
			request.Namespace,
			err,
		)
	}

	err = c.createJobConfigMap(request.EnsemblingJob, request.Namespace, request.Labels)
	if err != nil {
		return fmt.Errorf(
			"failed creating job specification configmap for job %s in namespace %s: %v",
			request.EnsemblingJob.Name,
			request.Namespace,
			err,
		)
	}

	_, err = c.createSparkApplication(request, serviceAccount)
	if err != nil {
		return fmt.Errorf(
			"failed submitting spark application to spark controller for job %s in namespace %s: %v",
			request.EnsemblingJob.Name,
			request.Namespace,
			err,
		)
	}

	return nil
}

func (c *ensemblingController) createSparkApplication(
	jobRequest *CreateEnsemblingJobRequest,
	serviceAccount *apicorev1.ServiceAccount,
) (*apisparkv1beta2.SparkApplication, error) {
	infraConfig := jobRequest.EnsemblingJob.InfraConfig
	request := &cluster.CreateSparkRequest{
		JobName:            jobRequest.EnsemblingJob.Name,
		JobLabels:          jobRequest.Labels,
		JobImageRef:        jobRequest.ImageRef,
		JobApplicationPath: ensemblingJobApplicationPath,
		JobArguments: []string{
			"--job-spec",
			fmt.Sprintf("%s%s", batch.JobConfigMount, batch.JobConfigFileName),
		},
		JobConfigMount:        batch.JobConfigMount,
		DriverCPURequest:      *infraConfig.GetResources().DriverCpuRequest,
		DriverMemoryRequest:   *infraConfig.GetResources().DriverMemoryRequest,
		ExecutorCPURequest:    *infraConfig.GetResources().ExecutorCpuRequest,
		ExecutorMemoryRequest: *infraConfig.GetResources().ExecutorMemoryRequest,
		ExecutorReplica:       *infraConfig.GetResources().ExecutorReplica,
		ServiceAccountName:    serviceAccount.Name,
		SparkInfraConfig:      c.sparkInfraConfig,
		EnvVars:               jobRequest.EnsemblingJob.InfraConfig.Env,
	}
	return c.clusterController.CreateSparkApplication(context.Background(), jobRequest.Namespace, request)
}

func (c *ensemblingController) createJobConfigMap(
	ensemblingJob *models.EnsemblingJob,
	namespace string,
	labels map[string]string,
) error {
	jobConfigJSON, err := json.Marshal(ensemblingJob.JobConfig)
	if err != nil {
		return err
	}

	jobConfigYAML, err := yaml.JSONToYAML(jobConfigJSON)
	if err != nil {
		return err
	}

	cm := &cluster.ConfigMap{
		Name:     ensemblingJob.Name,
		FileName: batch.JobConfigFileName,
		Data:     string(jobConfigYAML),
		Labels:   labels,
	}
	err = c.clusterController.ApplyConfigMap(context.Background(), namespace, cm)
	if err != nil {
		return err
	}

	return nil
}

func (c *ensemblingController) createSecret(request *CreateEnsemblingJobRequest, secretName string) error {
	secret := &cluster.Secret{
		Name:      request.EnsemblingJob.Name,
		Namespace: request.Namespace,
		Data: map[string]string{
			cluster.ServiceAccountFileName: secretName,
		},
		Labels: request.Labels,
	}
	// I'm not sure why we need to pass in a context here but not other kubernetes cluster functions.
	// Leaving a context.Background() until we figure out what to do with this.
	err := c.clusterController.CreateSecret(context.Background(), secret)
	if err != nil {
		return err
	}

	return nil
}

func (c *ensemblingController) cleanup(jobName string, namespace string) {
	err := c.clusterController.DeleteSecret(context.Background(), jobName, namespace, false)
	if err != nil {
		log.Warnf("failed deleting secret %s in namespace %s: %v", jobName, namespace, err)
	}

	err = c.clusterController.DeleteConfigMap(context.Background(), jobName, namespace, false)
	if err != nil {
		log.Warnf("failed deleting job spec %s in namespace %s: %v", jobName, namespace, err)
	}
}

func (c *ensemblingController) createSparkDriverAuthorization(
	namespace string,
	labels map[string]string,
) (*apicorev1.ServiceAccount, error) {
	serviceAccountName, roleName, roleBindingName := createAuthorizationResourceNames(namespace)

	saCfg := &cluster.ServiceAccount{
		Name:      serviceAccountName,
		Namespace: namespace,
		Labels:    labels,
	}
	sa, err := c.clusterController.CreateServiceAccount(context.Background(), namespace, saCfg)
	if err != nil {
		return nil, err
	}

	roleCfg := &cluster.Role{
		Name:        roleName,
		Namespace:   namespace,
		Labels:      labels,
		PolicyRules: cluster.DefaultSparkDriverRoleRules,
	}
	r, err := c.clusterController.CreateRole(context.Background(), namespace, roleCfg)
	if err != nil {
		return nil, err
	}

	roleBindingCfg := &cluster.RoleBinding{
		Name:               roleBindingName,
		Namespace:          namespace,
		Labels:             labels,
		RoleName:           r.Name,
		ServiceAccountName: sa.Name,
	}
	_, err = c.clusterController.CreateRoleBinding(context.Background(), namespace, roleBindingCfg)
	if err != nil {
		return nil, err
	}
	return sa, err
}

func createAuthorizationResourceNames(namespace string) (string, string, string) {
	serviceAccountName := fmt.Sprintf("%s-driver-sa", namespace)
	driverRoleName := fmt.Sprintf("%s-driver-role", namespace)
	driverRoleBindingName := fmt.Sprintf("%s-driver-role-binding", namespace)
	return serviceAccountName, driverRoleName, driverRoleBindingName
}
