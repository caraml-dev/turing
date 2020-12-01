package cluster

import (
	"context"
	"fmt"
	"io"
	"time"

	"k8s.io/client-go/kubernetes"

	apiappsv1 "k8s.io/api/apps/v1"
	apicorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	networkingv1alpha3 "istio.io/client-go/pkg/clientset/versioned/typed/networking/v1alpha3"

	rest "k8s.io/client-go/rest"

	"knative.dev/pkg/kmp"
	knservingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
	knservingclientset "knative.dev/serving/pkg/client/clientset/versioned"
	knservingclient "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1alpha1"

	"github.com/gojek/mlp/pkg/vault"
	"github.com/gojek/turing/api/turing/config"
	"github.com/pkg/errors"

	// Load required auth plugin
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

var (
	ErrNamespaceAlreadyExists = errors.New("namespace already exists")
)

// clusterConfig Model cluster authentication settings
type clusterConfig struct {
	// Kubernetes API server endpoint
	Host string
	// CA Certificate to trust for TLS
	CACert string
	// Client Certificate for authenticating to cluster
	ClientCert string
	// Client Key for authenticating to cluster
	ClientKey string
	// Cluster Name
	ClusterName string
	// GCP project where the cluster resides
	GcpProject string
}

// Controller defines the operations supported by the cluster controller
type Controller interface {
	DeployKnativeService(ctx context.Context, svc *KnativeService) error
	DeleteKnativeService(svcName string, namespace string, timeout time.Duration) error
	GetKnativeServiceURL(svcName string, namespace string) string
	ApplyIstioVirtualService(ctx context.Context, routerEndpoint *VirtualService) error
	DeleteIstioVirtualService(svcName string, namespace string, timeout time.Duration) error
	DeployKubernetesService(ctx context.Context, svc *KubernetesService) error
	DeleteKubernetesService(svcName string, namespace string, timeout time.Duration) error
	CreateNamespace(name string) error
	ApplyConfigMap(namespace string, configMap *ConfigMap) error
	DeleteConfigMap(name, namespace string) error
	CreateSecret(ctx context.Context, secret *Secret) error
	DeleteSecret(secretName string, namespace string) error
	ApplyPersistentVolumeClaim(ctx context.Context, namespace string, pvc *PersistentVolumeClaim) error
	DeletePersistentVolumeClaim(pvcName string, namespace string) error
	ListPods(namespace string, labelSelector string) (*apicorev1.PodList, error)
	ListPodLogs(namespace string, podName string, opts *apicorev1.PodLogOptions) (io.ReadCloser, error)
}

// controller implements the Controller interface
type controller struct {
	knServingClient knservingclient.ServingV1alpha1Interface
	k8sCoreClient   corev1.CoreV1Interface
	k8sAppsClient   appsv1.AppsV1Interface
	istioClient     networkingv1alpha3.NetworkingV1alpha3Interface
}

// newController initializes a new cluster controller with the given cluster config
func newController(clusterCfg clusterConfig) (Controller, error) {
	cfg := &rest.Config{
		Host: clusterCfg.Host,
		TLSClientConfig: rest.TLSClientConfig{
			Insecure: false,
			CAData:   []byte(clusterCfg.CACert),
			CertData: []byte(clusterCfg.ClientCert),
			KeyData:  []byte(clusterCfg.ClientKey),
		},
	}

	// Init the knative serving client
	knsClientSet, err := knservingclientset.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	// Init the k8s clientset
	k8sClientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	istioClientSet, err := networkingv1alpha3.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}

	return &controller{
		knServingClient: knsClientSet.ServingV1alpha1(),
		k8sCoreClient:   k8sClientset.CoreV1(),
		k8sAppsClient:   k8sClientset.AppsV1(),
		istioClient:     istioClientSet,
	}, nil
}

// InitClusterControllers takes in the the app config and a vault client and uses the credentials
// from vault to initialize one cluster controller per environment and returns a map where the
// key is the env name and the value is the corresponding controller.
func InitClusterControllers(
	cfg *config.Config,
	environmentClusterMap map[string]string,
	vaultClient vault.VaultClient,
) (map[string]Controller, error) {
	// For each supported environment, init controller
	controllers := make(map[string]Controller)
	for envName, clusterName := range environmentClusterMap {
		clusterSecret, err := vaultClient.GetClusterSecret(clusterName)
		if err != nil {
			return nil, errors.Wrapf(err,
				"unable to get cluster secret for cluster: %s", clusterName)
		}

		ctl, err := newController(clusterConfig{
			Host:       clusterSecret.Endpoint,
			CACert:     clusterSecret.CaCert,
			ClientCert: clusterSecret.ClientCert,
			ClientKey:  clusterSecret.ClientKey,

			ClusterName: clusterName,
			GcpProject:  cfg.DeployConfig.GcpProject,
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to initialize cluster controller")
		}

		controllers[envName] = ctl
	}

	return controllers, nil
}

// CreateNamespace creates a namespace. If the namespace already exists, will throw an error.
func (c *controller) CreateNamespace(name string) error {
	_, err := c.k8sCoreClient.Namespaces().Get(name, metav1.GetOptions{})
	if err == nil {
		return ErrNamespaceAlreadyExists
	}
	ns := apicorev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}
	_, err = c.k8sCoreClient.Namespaces().Create(&ns)
	return err
}

// ApplyConfigMap creates a config map in the namespace given the configuration if not exists.
// If the config map already exists, ApplyConfigMap will update the configuration with the given
// data.
func (c *controller) ApplyConfigMap(namespace string, configMap *ConfigMap) error {
	cm := apicorev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      configMap.Name,
			Namespace: namespace,
		},
		Data: map[string]string{
			configMap.FileName: configMap.Data,
		},
	}
	_, err := c.k8sCoreClient.ConfigMaps(namespace).Get(cm.Name, metav1.GetOptions{})
	if err == nil {
		// exists, we update instead
		_, err = c.k8sCoreClient.ConfigMaps(namespace).Update(&cm)
		return err
	}
	_, err = c.k8sCoreClient.ConfigMaps(namespace).Create(&cm)
	return err
}

// DeleteConfigMap deletes a configmap if exists.
func (c *controller) DeleteConfigMap(name, namespace string) error {
	_, err := c.k8sCoreClient.ConfigMaps(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return c.k8sCoreClient.ConfigMaps(namespace).Delete(name, &metav1.DeleteOptions{})
}

// Deploy creates / updates a Kubernetes/Knative service with the given specs
func (c *controller) DeployKnativeService(ctx context.Context, svcConf *KnativeService) error {
	var existingSvc *knservingv1alpha1.Service
	var err error

	// Build the deployment specs
	desiredSvc := svcConf.BuildKnativeServiceConfig()

	// Init knative ServicesGetter
	services := c.knServingClient.Services(svcConf.Namespace)

	// Check if service already exists. If exists, update it. If not, create.
	existingSvc, err = services.Get(svcConf.Name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Create new service
			_, err = services.Create(desiredSvc)
		} else {
			// Unexpected error, return it
			return err
		}
	} else {
		// Check for differences between current and new specs
		if !knServiceSemanticEquals(desiredSvc, existingSvc) {
			_, err = kmp.SafeDiff(
				desiredSvc.Spec.ConfigurationSpec,
				existingSvc.Spec.ConfigurationSpec,
			)
			if err != nil {
				return fmt.Errorf("Failed to diff Knative Service: %v", err)
			}
			// Update the existing service with the new config
			existingSvc.Spec.ConfigurationSpec = desiredSvc.Spec.ConfigurationSpec
			existingSvc.ObjectMeta.Labels = desiredSvc.ObjectMeta.Labels
			_, err = services.Update(existingSvc)
		}
	}

	if err != nil {
		return err
	}

	// Wait until service ready and return any errors
	return c.waitKnativeServiceReady(ctx, svcConf.Name, svcConf.Namespace)
}

// Delete removes the Kubernetes/Knative service and all related artifacts
func (c *controller) DeleteKnativeService(
	svcName string,
	namespace string,
	timeout time.Duration,
) error {
	// Init knative ServicesGetter
	services := c.knServingClient.Services(namespace)

	// Get the service
	_, err := services.Get(svcName, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// Delete the service
	gracePeriod := int64(timeout.Seconds())
	delOptions := &metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	}

	return services.Delete(svcName, delOptions)
}

// DeployKubernetesService deploys a kubernetes service and deployment
func (c *controller) DeployKubernetesService(
	ctx context.Context,
	svcConf *KubernetesService,
) error {

	desiredDeployment, desiredSvc := svcConf.BuildKubernetesServiceConfig()

	// Deploy deployment
	deployments := c.k8sAppsClient.Deployments(svcConf.Namespace)
	// Check if deployment already exists. If exists, update it. If not, create.
	var existingDeployment *apiappsv1.Deployment
	var err error
	existingDeployment, err = deployments.Get(svcConf.Name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Create new deployment
			_, err = deployments.Create(desiredDeployment)
		} else {
			// Unexpected error, return it
			return err
		}
	} else {
		// Check for differences between current and new specs
		if !k8sDeploymentSemanticEquals(desiredDeployment, existingDeployment) {
			// Update the existing service with the new config
			existingDeployment.Spec.Template = desiredDeployment.Spec.Template
			existingDeployment.ObjectMeta.Labels = desiredDeployment.ObjectMeta.Labels
			_, err = deployments.Update(existingDeployment)
		}
	}
	if err != nil {
		return err
	}

	// Deploy Service
	services := c.k8sCoreClient.Services(svcConf.Namespace)
	// Check if service already exists. If exists, update it. If not, create.
	var existingSvc *apicorev1.Service
	existingSvc, err = services.Get(svcConf.Name, metav1.GetOptions{})
	if err != nil {
		if k8serrors.IsNotFound(err) {
			// Create new service
			_, err = services.Create(desiredSvc)
		} else {
			// Unexpected error, return it
			return err
		}
	} else {
		// Check for differences between current and new specs
		if !k8sServiceSemanticEquals(desiredSvc, existingSvc) {
			_, err = services.Update(desiredSvc)
		}
	}
	if err != nil {
		return err
	}

	// Wait until deployment ready and return any errors
	return c.waitDeploymentReady(ctx, svcConf.Name, svcConf.Namespace)
}

// DeleteKubernetesService deletes a kubernetes service an deployment
func (c *controller) DeleteKubernetesService(svcName string, namespace string, timeout time.Duration) error {
	gracePeriod := int64(timeout.Seconds())
	delOptions := &metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	}

	deployments := c.k8sAppsClient.Deployments(namespace)
	_, err := deployments.Get(svcName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	err = deployments.Delete(svcName, delOptions)
	if err != nil {
		return err
	}

	services := c.k8sCoreClient.Services(namespace)
	_, err = services.Get(svcName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	return services.Delete(svcName, delOptions)
}

// ApplyIstioVirtualService creates a virtual service if not exists, if exists, updates the
// existing service with the new configuration
func (c *controller) ApplyIstioVirtualService(ctx context.Context, routerEndpoint *VirtualService) error {
	vservices := c.istioClient.VirtualServices(routerEndpoint.Namespace)
	existingVsvc, err := vservices.Get(routerEndpoint.Name, metav1.GetOptions{})
	if err == nil {
		// patch
		existingVsvc.Spec.Http = routerEndpoint.BuildVirtualService().Spec.Http
		_, err := vservices.Update(existingVsvc)
		return err
	}
	_, err = vservices.Create(routerEndpoint.BuildVirtualService())
	return err
}

// DeleteIstioVirtualService deletes an istio virtual service.
func (c *controller) DeleteIstioVirtualService(svcName string, namespace string, timeout time.Duration) error {
	vservices := c.istioClient.VirtualServices(namespace)
	_, err := vservices.Get(svcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to retrieve virtual service %s: %s", svcName, err.Error())
	}
	gracePeriod := int64(timeout.Seconds())
	delOptions := &metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriod,
	}
	return vservices.Delete(svcName, delOptions)
}

// CreateSecret creates a secret. If the secret already exists, the existing secret will be updated.
func (c *controller) CreateSecret(ctx context.Context, secret *Secret) error {
	secrets := c.k8sCoreClient.Secrets(secret.Namespace)
	_, err := secrets.Get(secret.Name, metav1.GetOptions{})
	if err == nil {
		_, err = secrets.Update(secret.BuildSecret())
		return err
	}
	_, err = secrets.Create(secret.BuildSecret())
	return err
}

// DeleteSecret deletes a secret
func (c *controller) DeleteSecret(secretName string, namespace string) error {
	secrets := c.k8sCoreClient.Secrets(namespace)
	_, err := secrets.Get(secretName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to get secret with name %s: %s", secretName, err.Error())
	}
	return secrets.Delete(secretName, &metav1.DeleteOptions{})
}

// GetKnativeServiceURL returns the URL at which the given service is accessible, if found.
// Else, an empty string is returned.
func (c *controller) GetKnativeServiceURL(svcName string, namespace string) string {
	var url string

	// Init knative ServicesGetter
	services := c.knServingClient.Services(namespace)

	// Get the service
	svc, err := services.Get(svcName, metav1.GetOptions{})

	// If the service is ready, get its url
	if err == nil && svc.Status.IsReady() {
		url = svc.Status.URL.String()
	}

	return url
}

// ApplyPersistentVolumeClaim creates a PVC in the given namespace.
// If the PVC already exists, it will update the existing PVC.
func (c *controller) ApplyPersistentVolumeClaim(
	ctx context.Context,
	namespace string,
	pvcCfg *PersistentVolumeClaim,
) error {
	pvcs := c.k8sCoreClient.PersistentVolumeClaims(namespace)
	existingPVC, err := pvcs.Get(pvcCfg.Name, metav1.GetOptions{})
	pvc := pvcCfg.BuildPersistentVolumeClaim()

	// If not exists, create
	if err != nil {
		_, err := pvcs.Create(pvc)
		return err
	}
	// If exists, update
	existingPVC.Spec.Resources = pvc.Spec.Resources
	_, err = pvcs.Update(existingPVC)
	return err
}

// ApplyPersistentVolumeClaim deletes the PVC in the given namespace.
func (c *controller) DeletePersistentVolumeClaim(pvcName string, namespace string) error {
	pvcs := c.k8sCoreClient.PersistentVolumeClaims(namespace)
	_, err := pvcs.Get(pvcName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to get pvc with name %s: %s", pvcName, err.Error())
	}
	return pvcs.Delete(pvcName, &metav1.DeleteOptions{})
}

func (c *controller) ListPods(namespace string, labelSelector string) (*apicorev1.PodList, error) {
	return c.k8sCoreClient.Pods(namespace).List(metav1.ListOptions{LabelSelector: labelSelector})
}

func (c *controller) ListPodLogs(
	namespace string,
	podName string,
	opts *apicorev1.PodLogOptions,
) (io.ReadCloser, error) {
	return c.k8sCoreClient.Pods(namespace).GetLogs(podName, opts).Stream()
}

// waitKnativeServiceReady waits for the given knative service to become ready, until the
// default timeout
func (c *controller) waitKnativeServiceReady(
	ctx context.Context,
	svcName string,
	namespace string,
) error {
	// Init ticker to check status every second
	ticker := time.NewTicker(time.Second)

	// Init knative ServicesGetter
	services := c.knServingClient.Services(namespace)

	for {
		select {
		case <-ctx.Done():
			terminationMessage := c.getKnativePodTerminationMessage(svcName, namespace)
			return fmt.Errorf("timeout waiting for service %s to be ready: %s", svcName, terminationMessage)
		case <-ticker.C:
			svc, err := services.Get(svcName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("unable to get service status for %s: %v", svcName, err)
			}

			if svc.Status.IsReady() {
				// Service is completely ready
				return nil
			}
		}
	}
}

// getKnativePodTerminationMessage retrieves the termination message of the user container
// in the pod, which will be returned for logging as a part of the deployment failure error.
func (c *controller) getKnativePodTerminationMessage(svcName string, namespace string) string {
	labelSelector := KnativeServiceLabelKey + "=" + svcName
	podList, err := c.ListPods(namespace, labelSelector)
	if err != nil {
		return err.Error()
	}

	var terminationMessage string
	for _, pod := range podList.Items {
		for _, containerStatus := range pod.Status.ContainerStatuses {
			if containerStatus.Name == KnativeUserContainerName {
				if containerStatus.LastTerminationState.Terminated != nil {
					terminationMessage = containerStatus.LastTerminationState.Terminated.Message
					break
				}
			}

		}
	}
	return terminationMessage
}

// waitDeploymentReady waits for the given k8s deployment to become ready, until the
// default timeout
func (c *controller) waitDeploymentReady(
	ctx context.Context,
	deploymentName string,
	namespace string,
) error {
	// Init ticker to check status every second
	ticker := time.NewTicker(time.Second)

	// Init knative ServicesGetter
	deployments := c.k8sAppsClient.Deployments(namespace)

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for deployment %s to be ready", deploymentName)
		case <-ticker.C:
			deployment, err := deployments.Get(deploymentName, metav1.GetOptions{})
			if err != nil {
				return fmt.Errorf("unable to get deployment status for %s: %v", deploymentName, err)
			}

			if deploymentReady(deployment) {
				// Service is completely ready
				return nil
			}
		}
	}
}

func deploymentReady(deployment *apiappsv1.Deployment) bool {
	if deployment.Generation <= deployment.Status.ObservedGeneration {
		cond := deployment.Status.Conditions[0]
		ready := cond.Type == apiappsv1.DeploymentAvailable
		if deployment.Spec.Replicas != nil {
			// Account for replica surge during updates
			ready = ready &&
				deployment.Status.ReadyReplicas == *deployment.Spec.Replicas &&
				deployment.Status.Replicas == *deployment.Spec.Replicas
		}
		return ready
	}
	return false
}

func knServiceSemanticEquals(desiredService, service *knservingv1alpha1.Service) bool {
	return equality.Semantic.DeepEqual(
		desiredService.Spec.ConfigurationSpec,
		service.Spec.ConfigurationSpec) &&
		equality.Semantic.DeepEqual(desiredService.ObjectMeta.Labels, service.ObjectMeta.Labels)
}

func k8sDeploymentSemanticEquals(desiredDeployment, deployment *apiappsv1.Deployment) bool {
	return equality.Semantic.DeepEqual(desiredDeployment.Spec.Template, deployment.Spec.Template) &&
		equality.Semantic.DeepEqual(desiredDeployment.ObjectMeta.Labels, deployment.ObjectMeta.Labels) &&
		desiredDeployment.Spec.Replicas == deployment.Spec.Replicas
}

func k8sServiceSemanticEquals(desiredService, service *apicorev1.Service) bool {
	return equality.Semantic.DeepEqual(desiredService.Spec.Ports, service.Spec.Ports) &&
		equality.Semantic.DeepEqual(desiredService.ObjectMeta.Labels, service.ObjectMeta.Labels) &&
		equality.Semantic.DeepEqual(desiredService.Spec.Selector, service.Spec.Selector)
}
