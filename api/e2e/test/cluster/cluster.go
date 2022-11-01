package cluster

import (
	"context"
	"fmt"
	"path/filepath"
	"reflect"

	"github.com/gojek/mlp/api/pkg/vault"
	"github.com/pkg/errors"
	"istio.io/client-go/pkg/apis/networking/v1alpha3"
	networkingv1alpha3 "istio.io/client-go/pkg/clientset/versioned/typed/networking/v1alpha3"
	appsV1 "k8s.io/api/apps/v1"
	coreV1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
	"k8s.io/client-go/kubernetes"
	appsv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
	knservingclientset "knative.dev/serving/pkg/client/clientset/versioned"
	knservingclient "knative.dev/serving/pkg/client/clientset/versioned/typed/serving/v1"

	"github.com/caraml-dev/turing/api/e2e/test/config"
)

var deleteTimeoutSeconds = int64(20)

var clients = &struct {
	KnServingClient       knservingclient.ServingV1Interface
	K8sCoreClient         corev1.CoreV1Interface
	K8sAppsClient         appsv1.AppsV1Interface
	IstioNetworkingClient networkingv1alpha3.NetworkingV1alpha3Interface
}{}

type TuringRouterResources struct {
	KnativeServices []knservingv1.Service
	K8sServices     []coreV1.Service
	IstioServices   []v1alpha3.VirtualService
	K8sDeployments  []appsV1.Deployment
	ConfigMaps      []coreV1.ConfigMap
	Secrets         []coreV1.Secret
	PVCs            []coreV1.PersistentVolumeClaim
}

type DeleteInterface interface {
	Delete(ctx context.Context, name string, opts metav1.DeleteOptions) error
}

func GetModelClusterCredentials(clusterName string, vaultCfg config.VaultConfig) (*vault.ClusterSecret, error) {
	vaultConfig := &vault.Config{
		Address: vaultCfg.Address,
		Token:   vaultCfg.Token,
	}
	vaultClient, err := vault.NewVaultClient(vaultConfig)
	if err != nil {
		return nil, errors.Wrap(err, "unable to initialize vault")
	}

	// Get cluster secret
	clusterSecret, err := vaultClient.GetClusterSecret(clusterName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to get cluster secret for cluster: %s", clusterName)
	}

	return clusterSecret, nil
}

func InitClusterClients(cfg *config.Config) error {
	var clusterCfg *rest.Config

	if cfg.KubeconfigUseLocal {
		p := cfg.KubeconfigFilePath
		if cfg.KubeconfigFilePath == "" {
			p = filepath.Join(homedir.HomeDir(), ".kube", "config")
		}
		// Authenticate to Kube with local kubeconfig in $HOME/.kube/config
		cfg, err := clientcmd.BuildConfigFromFlags("", p)
		if err != nil {
			return errors.Wrap(err, "unable to initialize Kube config from: "+p)
		}
		clusterCfg = cfg
	} else {
		// Authenticate to Kube with cluster credentials in Vault
		vaultConfig := &vault.Config{
			Address: cfg.Vault.Address,
			Token:   cfg.Vault.Token,
		}
		vaultClient, err := vault.NewVaultClient(vaultConfig)
		if err != nil {
			return errors.Wrap(err, "unable to initialize vault")
		}
		clusterSecret, err := vaultClient.GetClusterSecret(cfg.Cluster.Name)
		if err != nil {
			return errors.Wrapf(err,
				"unable to get cluster secret for cluster: %s", cfg.Cluster.Name)
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
		return err
	}
	k8sClientSet, err := kubernetes.NewForConfig(clusterCfg)
	if err != nil {
		return err
	}
	istioClientSet, err := networkingv1alpha3.NewForConfig(clusterCfg)
	if err != nil {
		return err
	}

	clients.KnServingClient = knsClientSet.ServingV1()
	clients.K8sCoreClient = k8sClientSet.CoreV1()
	clients.K8sAppsClient = k8sClientSet.AppsV1()
	clients.IstioNetworkingClient = istioClientSet

	return nil
}

func deleteK8sResource(
	resourceInterface DeleteInterface,
	resources []string,
) {
	ctx := context.Background()
	delOptions := metav1.DeleteOptions{
		GracePeriodSeconds: &deleteTimeoutSeconds,
	}

	for _, name := range resources {
		err := resourceInterface.Delete(ctx, name, delOptions)
		if err != nil && !k8serrors.IsNotFound(err) {
			fmt.Printf("Error deleting resource %s: %v\n", name, err)
		}
	}
}

func ListRouterResources(projectName, routerName string) (TuringRouterResources, error) {
	ctx := context.Background()
	req, _ := labels.NewRequirement(
		"app", selection.Equals, []string{routerName})
	opts := metav1.ListOptions{LabelSelector: labels.NewSelector().Add(*req).String()}

	var resources TuringRouterResources
	ksvc, err := clients.KnServingClient.Services(projectName).List(ctx, opts)
	if err != nil {
		return resources, err
	}

	svc, err := clients.K8sCoreClient.Services(projectName).List(ctx, opts)
	if err != nil {
		return resources, err
	}

	deploy, err := clients.K8sAppsClient.Deployments(projectName).List(ctx, opts)
	if err != nil {
		return resources, err
	}

	secrets, err := clients.K8sCoreClient.Secrets(projectName).List(ctx, opts)
	if err != nil {
		return resources, err
	}

	cMaps, err := clients.K8sCoreClient.ConfigMaps(projectName).List(ctx, opts)
	if err != nil {
		return resources, err
	}

	pvc, err := clients.K8sCoreClient.PersistentVolumeClaims(projectName).List(ctx, opts)
	if err != nil {
		return resources, err
	}

	vSvc, err := clients.IstioNetworkingClient.VirtualServices(projectName).List(ctx, opts)
	if err != nil {
		return resources, err
	}

	return TuringRouterResources{
		KnativeServices: ksvc.Items,
		K8sServices:     svc.Items,
		IstioServices:   vSvc.Items,
		K8sDeployments:  deploy.Items,
		ConfigMaps:      cMaps.Items,
		Secrets:         secrets.Items,
		PVCs:            pvc.Items,
	}, nil
}

func CleanupRouterDeployment(
	projectName, routerName string,
) error {
	resources, err := ListRouterResources(projectName, routerName)
	if err != nil {
		return err
	}

	knSvcInterface := clients.KnServingClient.Services(projectName)
	svcInterface := clients.K8sCoreClient.Services(projectName)
	deploymentsInterface := clients.K8sAppsClient.Deployments(projectName)
	secretsInterface := clients.K8sCoreClient.Secrets(projectName)
	cfgMapsInterface := clients.K8sCoreClient.ConfigMaps(projectName)
	pvcsInterface := clients.K8sCoreClient.PersistentVolumeClaims(projectName)
	istioSvcInterface := clients.IstioNetworkingClient.VirtualServices(projectName)

	extractNames := func(objects interface{}) []string {
		slice := reflect.ValueOf(objects)
		names := make([]string, slice.Len())

		for i := 0; i < slice.Len(); i++ {
			names[i] = slice.Index(i).FieldByName("ObjectMeta").
				Interface().(metav1.ObjectMeta).Name
		}
		return names
	}

	deleteK8sResource(knSvcInterface, extractNames(resources.KnativeServices))
	deleteK8sResource(svcInterface, extractNames(resources.K8sServices))
	deleteK8sResource(deploymentsInterface, extractNames(resources.K8sDeployments))
	deleteK8sResource(secretsInterface, extractNames(resources.Secrets))
	deleteK8sResource(cfgMapsInterface, extractNames(resources.ConfigMaps))
	deleteK8sResource(pvcsInterface, extractNames(resources.PVCs))
	deleteK8sResource(istioSvcInterface, extractNames(resources.IstioServices))

	return nil
}
