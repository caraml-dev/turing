package batchcontroller

import (
	"context"
	"errors"
	"fmt"

	apisparkv1beta2 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	"github.com/gojek/turing/api/turing/cluster"
	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/log"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service"
	apicorev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"
)

const (
	ensemblingJobApplicationPath = "local:///home/spark/batch-ensembler/main.py"
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
}

type ensemblingController struct {
	clusterController cluster.Controller
	mlpService        service.MLPService
	sparkInfraConfig  *config.SparkInfraConfig
}

// NewBatchEnsemblingController creates a new batch ensembling controller
func NewBatchEnsemblingController(
	clusterController cluster.Controller,
	mlpService service.MLPService,
	sparkInfraConfig *config.SparkInfraConfig,
) EnsemblingController {
	return &ensemblingController{
		clusterController: clusterController,
		mlpService:        mlpService,
		sparkInfraConfig:  sparkInfraConfig,
	}
}

func (c *ensemblingController) Create(request *CreateEnsemblingJobRequest) error {
	var err error
	defer func() {
		if err != nil {
			c.cleanup(request.EnsemblingJob.Name, request.Namespace)
		}
	}()

	err = c.clusterController.CreateNamespace(request.Namespace)
	if errors.Is(err, cluster.ErrNamespaceAlreadyExists) {
		// This error is ok to ignore because we just need the namespace.
		err = nil
	}

	if err != nil {
		return fmt.Errorf("failed creating namespace %s: %v", request.Namespace, err)
	}

	serviceAccount, err := c.createDriverAuthorization(request.Namespace)
	if err != nil {
		return fmt.Errorf(
			"failed creating spark driver authorization in namespace %s: %v",
			request.Namespace,
			err,
		)
	}

	secretString, err := c.mlpService.GetSecret(
		request.EnsemblingJob.ProjectID,
		request.EnsemblingJob.InfraConfig.ServiceAccountName,
	)
	if err != nil {
		return fmt.Errorf(
			"service account %s is not found within %s project: %s",
			request.EnsemblingJob.InfraConfig.ServiceAccountName,
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

	err = c.createJobConfigMap(request.EnsemblingJob, request.Namespace)
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
			fmt.Sprintf("%s%s", cluster.JobConfigMount, cluster.JobConfigFileName),
		},
		DriverCPURequest:      infraConfig.Resources.DriverCPURequest,
		DriverMemoryRequest:   infraConfig.Resources.DriverMemoryRequest,
		ExecutorCPURequest:    infraConfig.Resources.ExecutorCPURequest,
		ExecutorMemoryRequest: infraConfig.Resources.ExecutorMemoryRequest,
		ExecutorReplica:       infraConfig.Resources.ExecutorReplica,
		ServiceAccountName:    serviceAccount.Name,
		SparkInfraConfig:      c.sparkInfraConfig,
	}
	return c.clusterController.CreateSparkApplication(jobRequest.Namespace, request)
}

func (c *ensemblingController) createJobConfigMap(ensemblingJob *models.EnsemblingJob, namespace string) error {
	jobConfigJSON, err := ensemblingJob.JobConfig.MarshalJSON()
	if err != nil {
		return err
	}

	jobConfigYAML, err := yaml.JSONToYAML(jobConfigJSON)
	if err != nil {
		return err
	}

	cm := &cluster.ConfigMap{
		Name:     ensemblingJob.Name,
		FileName: cluster.JobConfigFileName,
		Data:     string(jobConfigYAML),
	}
	err = c.clusterController.ApplyConfigMap(namespace, cm)
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
	}
	// I'm not sure why we need to pass in a context here but not other kubernetes cluster functions.
	// Leaving a context.TODO() until we figure out what to do with this.
	err := c.clusterController.CreateSecret(context.TODO(), secret)
	if err != nil {
		return err
	}

	return nil
}

func (c *ensemblingController) cleanup(jobName string, namespace string) {
	err := c.clusterController.DeleteSecret(jobName, namespace)
	if err != nil {
		log.Warnf("failed deleting secret %s in namespace %s: %v", jobName, namespace, err)
	}

	err = c.clusterController.DeleteConfigMap(jobName, namespace)
	if err != nil {
		log.Warnf("failed deleting job spec %s in namespace %s: %v", jobName, namespace, err)
	}
}

func (c *ensemblingController) createDriverAuthorization(namespace string) (*apicorev1.ServiceAccount, error) {
	serviceAccountName, roleName, roleBindingName := createAuthorizationResourceNames(namespace)

	sa, err := c.clusterController.CreateServiceAccount(namespace, serviceAccountName)
	if err != nil {
		return nil, err
	}

	r, err := c.clusterController.CreateRole(namespace, roleName, cluster.DefaultSparkDriverRoleRules)
	if err != nil {
		return nil, err
	}

	_, err = c.clusterController.CreateRoleBinding(namespace, roleBindingName, sa.Name, r.Name)
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
