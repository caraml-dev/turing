//go:build e2e
// +build e2e

package e2e

import (
	"context"
	"fmt"
	"path/filepath"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"

	"github.com/gojek/mlp/api/pkg/vault"
	"github.com/pkg/errors"
	networkingv1alpha3 "istio.io/client-go/pkg/clientset/versioned/typed/networking/v1alpha3"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	knservingclientset "knative.dev/serving/pkg/client/clientset/versioned"
	knservingclient "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"
)

// TestClusterClients holds the Clients for K8s / KNative resource groups
type TestClusterClients struct {
	KnServingClient       knservingclient.ServingV1Interface
	K8sCoreClient         corev1.CoreV1Interface
	K8sAppsClient         appsv1.AppsV1Interface
	IstioNetworkingClient networkingv1alpha3.NetworkingV1alpha3Interface
}

// newClusterClients is a creator for the TestClusterClients
func newClusterClients(cfg *testConfig) (*TestClusterClients, error) {
	var clusterCfg *rest.Config

	if cfg.KubeconfigUseLocal {
		p := cfg.KubeconfigFilePath
		if cfg.KubeconfigFilePath == "" {
			p = filepath.Join(homedir.HomeDir(), ".kube", "config")
		}
		// Authenticate to Kube with local kubeconfig in $HOME/.kube/config
		cfg, err := clientcmd.BuildConfigFromFlags("", p)
		if err != nil {
			return nil, errors.Wrap(err, "unable to initialize Kube config from: "+p)
		}
		clusterCfg = cfg
	} else {
		// Authenticate to Kube with cluster credentials in Vault
		vaultConfig := &vault.Config{
			Address: cfg.VaultAddress,
			Token:   cfg.VaultToken,
		}
		vaultClient, err := vault.NewVaultClient(vaultConfig)
		if err != nil {
			return nil, errors.Wrap(err, "unable to initialize vault")
		}
		clusterSecret, err := vaultClient.GetClusterSecret(cfg.ClusterName)
		if err != nil {
			return nil, errors.Wrapf(err,
				"unable to get cluster secret for cluster: %s", cfg.ClusterName)
		}
		clusterCfg = &rest.Config{
			Host: clusterSecret.Endpoint,
			TLSClientConfig: rest.TLSClientConfig{
				Insecure: false,
				CAData:   []byte(clusterSecret.CaCert),
				CertData: []byte(clusterSecret.ClientCert),
				KeyData:  []byte(clusterSecret.ClientKey),
			},
		}
	}

	knsClientSet, err := knservingclientset.NewForConfig(clusterCfg)
	if err != nil {
		return nil, err
	}
	k8sClientset, err := kubernetes.NewForConfig(clusterCfg)
	if err != nil {
		return nil, err
	}
	istioClientset, err := networkingv1alpha3.NewForConfig(clusterCfg)
	if err != nil {
		return nil, err
	}
	return &TestClusterClients{
		KnServingClient:       knsClientSet.ServingV1(),
		K8sCoreClient:         k8sClientset.CoreV1(),
		K8sAppsClient:         k8sClientset.AppsV1(),
		IstioNetworkingClient: istioClientset,
	}, nil
}

func isConfigMapExists(
	clusterClients *TestClusterClients,
	projectName string,
	name string,
) bool {
	_, err := clusterClients.K8sCoreClient.
		ConfigMaps(projectName).
		Get(context.Background(), name, metav1.GetOptions{})
	return err == nil
}

func isPersistentVolumeClaimExists(clusterClients *TestClusterClients,
	projectName string,
	name string,
) bool {
	_, err := clusterClients.K8sCoreClient.
		PersistentVolumeClaims(projectName).
		Get(context.Background(), name, metav1.GetOptions{})
	return err == nil
}

func isDeploymentExists(
	clusterClients *TestClusterClients,
	projectName string,
	name string,
) bool {
	_, err := clusterClients.K8sAppsClient.
		Deployments(projectName).
		Get(context.Background(), name, metav1.GetOptions{})
	return err == nil
}

func getRouterDownstream(clusterClients *TestClusterClients, projectName string, routerName string) (string, error) {
	vs, err := clusterClients.
		IstioNetworkingClient.
		VirtualServices(projectName).
		Get(context.Background(), routerName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return vs.Spec.Http[0].Route[0].Headers.Request.Set["Host"], nil
}

func deleteExperiments(
	clusterClients *TestClusterClients,
	projectName string,
	exps []struct {
		Name       string
		MaxVersion int
	}) {
	// Create a list of each type of resources
	knServiceNames := []string{}
	serviceNames := []string{}
	deploymentNames := []string{}
	secrets := []string{}
	configmaps := []string{}
	pvcs := []string{}
	istioSvcs := []string{}

	for _, exp := range exps {
		// Resources must be cleared for each version created for the experiment
		for ver := 1; ver <= exp.MaxVersion; ver++ {
			knServiceNames = append(knServiceNames,
				fmt.Sprintf("%s-turing-router-%d", exp.Name, exp.MaxVersion),
				fmt.Sprintf("%s-turing-enricher-%d", exp.Name, exp.MaxVersion),
				fmt.Sprintf("%s-turing-ensembler-%d", exp.Name, exp.MaxVersion),
			)

			serviceNames = append(serviceNames,
				fmt.Sprintf("%s-turing-fluentd-logger-%d", exp.Name, exp.MaxVersion))
			deploymentNames = append(deploymentNames,
				fmt.Sprintf("%s-turing-fluentd-logger-%d", exp.Name, exp.MaxVersion))
			secrets = append(secrets, fmt.Sprintf("%s-turing-secret-%d", exp.Name, exp.MaxVersion))
			configmaps = append(configmaps, fmt.Sprintf("%s-turing-fiber-config-%d", exp.Name, exp.MaxVersion))
			pvcs = append(pvcs, fmt.Sprintf("%s-turing-cache-volume-%d", exp.Name, exp.MaxVersion))
			istioSvcs = append(istioSvcs, fmt.Sprintf("%s-turing-router", exp.Name))
		}
	}

	// Delete K8s Services
	deleteServices(clusterClients, projectName, serviceNames)
	// Delete K8s Deployments
	deleteDeployments(clusterClients, projectName, deploymentNames)
	// Delete KNative Services
	deleteKnServices(clusterClients, projectName, knServiceNames)
	// Delete Secrets
	deleteSecrets(clusterClients, projectName, secrets)
	// Delte Configmaps
	deleteConfigmaps(clusterClients, projectName, configmaps)
	// Delte PVCs
	deletePVCs(clusterClients, projectName, pvcs)
	// Delete Istio virtual services
	deleteIstioVirtualServices(clusterClients, projectName, istioSvcs)
}

func deleteServices(
	clusterClients *TestClusterClients,
	projectName string,
	serviceNames []string,
) {
	gracePeriodSeconds := int64(deleteTimeoutSeconds)
	delOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}

	services := clusterClients.K8sCoreClient.Services(projectName)
	for _, name := range serviceNames {
		err := services.Delete(context.Background(), name, delOptions)
		if err != nil && !k8serrors.IsNotFound(err) {
			fmt.Printf("Error deleting service %s: %v\n", name, err)
		}
	}
}

func deleteDeployments(
	clusterClients *TestClusterClients,
	projectName string,
	deploymentNames []string,
) {
	gracePeriodSeconds := int64(deleteTimeoutSeconds)
	delOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}

	deployments := clusterClients.K8sAppsClient.Deployments(projectName)
	for _, name := range deploymentNames {
		err := deployments.Delete(context.Background(), name, delOptions)
		if err != nil && !k8serrors.IsNotFound(err) {
			fmt.Printf("Error deleting deployment %s: %v\n", name, err)
		}
	}
}

func deleteKnServices(
	clusterClients *TestClusterClients,
	projectName string,
	knServiceNames []string,
) {
	gracePeriodSeconds := int64(deleteTimeoutSeconds)
	delOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}

	services := clusterClients.KnServingClient.Services(projectName)
	for _, name := range knServiceNames {
		err := services.Delete(context.Background(), name, delOptions)
		if err != nil && !k8serrors.IsNotFound(err) {
			fmt.Printf("Error deleting Knative service %s: %v\n", name, err)
		}
	}
}

func deleteSecrets(
	clusterClients *TestClusterClients,
	projectName string,
	secretNames []string,
) {
	gracePeriodSeconds := int64(deleteTimeoutSeconds)
	delOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}

	secrets := clusterClients.K8sCoreClient.Secrets(projectName)
	for _, name := range secretNames {
		err := secrets.Delete(context.Background(), name, delOptions)
		if err != nil && !k8serrors.IsNotFound(err) {
			fmt.Printf("Error deleting secret %s: %v\n", name, err)
		}
	}
}

func deleteConfigmaps(
	clusterClients *TestClusterClients,
	projectName string,
	cfgMapNames []string,
) {
	gracePeriodSeconds := int64(deleteTimeoutSeconds)
	delOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}

	configMaps := clusterClients.K8sCoreClient.ConfigMaps(projectName)
	for _, name := range cfgMapNames {
		err := configMaps.Delete(context.Background(), name, delOptions)
		if err != nil && !k8serrors.IsNotFound(err) {
			fmt.Printf("Error deleting configmap %s: %v\n", name, err)
		}
	}
}

func deletePVCs(
	clusterClients *TestClusterClients,
	projectName string,
	pvcs []string,
) {
	gracePeriodSeconds := int64(deleteTimeoutSeconds)
	delOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}

	clusterPVCs := clusterClients.K8sCoreClient.PersistentVolumeClaims(projectName)
	for _, name := range pvcs {
		err := clusterPVCs.Delete(context.Background(), name, delOptions)
		if err != nil && !k8serrors.IsNotFound(err) {
			fmt.Printf("Error deleting PVC %s: %v\n", name, err)
		}
	}
}

func deleteIstioVirtualServices(
	clusterClients *TestClusterClients,
	projectName string,
	svcs []string,
) {
	gracePeriodSeconds := int64(deleteTimeoutSeconds)
	delOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}

	vservices := clusterClients.IstioNetworkingClient.VirtualServices(projectName)
	for _, name := range svcs {
		err := vservices.Delete(context.Background(), name, delOptions)
		if err != nil && !k8serrors.IsNotFound(err) {
			fmt.Printf("Error deleting Istio Virtual Service %s: %v\n", name, err)
		}
	}
}
