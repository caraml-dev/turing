// +build unit

package cluster

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	"bou.ke/monkey"

	"github.com/stretchr/testify/assert"
	istioclientset "istio.io/client-go/pkg/clientset/versioned/fake"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	knservingv1alpha1 "knative.dev/serving/pkg/apis/serving/v1alpha1"
	knservingclientset "knative.dev/serving/pkg/client/clientset/versioned/fake"
)

type reactor struct {
	verb     string
	resource string
	rFunc    k8stesting.ReactionFunc
}

var reactorVerbs = struct {
	Get    string
	Create string
	Update string
	Delete string
}{
	Get:    "get",
	Create: "create",
	Update: "update",
	Delete: "delete",
}

const (
	knativeGroup    = "serving.knative.dev"
	knativeVersion  = "v1alpha1"
	knativeResource = "services"
)

func TestDeployKnativeService(t *testing.T) {
	testName, testNamespace := "test-name", "test-namespace"
	resourceItem := schema.GroupVersionResource{
		Group:    knativeGroup,
		Version:  knativeVersion,
		Resource: knativeResource,
	}
	testKnSvc := &knservingv1alpha1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: testName,
		},
	}
	svcConf := &KnativeService{
		BaseService: &BaseService{
			Name:      testName,
			Namespace: testNamespace,
		},
	}

	// Define reactor for a successful get
	getSuccess := func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, &knservingv1alpha1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: testName,
			},
			Status: knservingv1alpha1.ServiceStatus{
				Status: duckv1.Status{
					ObservedGeneration: 1,
					Conditions: duckv1.Conditions{
						apis.Condition{
							Type:   apis.ConditionReady,
							Status: corev1.ConditionTrue,
						},
					},
				},
			},
		}, nil
	}

	// Define tests
	cs := knservingclientset.NewSimpleClientset()
	tests := map[string][]reactor{
		"new_service": {
			{
				verb:     reactorVerbs.Get,
				resource: knativeResource,
				rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
					expAction := k8stesting.NewGetAction(resourceItem, testNamespace, testName)
					// Check that the method is called with the expected action
					assert.Equal(t, expAction, action)
					// Return nil object and error to indicate non existent object
					return true, nil, k8serrors.NewNotFound(schema.GroupResource{}, testName)
				},
			},
			{
				verb:     reactorVerbs.Create,
				resource: knativeResource,
				rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
					expAction := k8stesting.NewCreateAction(resourceItem, testNamespace, testKnSvc)
					// Check that the method is called with the expected action
					assert.Equal(t, expAction, action)
					// Prepend a new get reactor for waitKnativeServiceReady to use
					cs.PrependReactor(reactorVerbs.Get, knativeResource, getSuccess)
					// Nil error indicates Create success
					return true, testKnSvc, nil
				},
			},
		},
		"update_service": {
			{
				verb:     reactorVerbs.Get,
				resource: knativeResource,
				rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
					expAction := k8stesting.NewGetAction(resourceItem, testNamespace, testName)
					// Check that the method is called with the expected action
					assert.Equal(t, expAction, action)
					// Return valid object and nil errror to trigger Update
					return true, testKnSvc, nil
				},
			},
			{
				verb:     reactorVerbs.Update,
				resource: knativeResource,
				rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
					expAction := k8stesting.NewUpdateAction(resourceItem, testNamespace, testKnSvc)
					// Check that the method is called with the expected action
					assert.Equal(t, expAction, action)
					// Prepend a new get reactor for waitKnativeServiceReady to use
					cs.PrependReactor(reactorVerbs.Get, knativeResource, getSuccess)
					// Nil error indicates Update success
					return true, testKnSvc, nil
				},
			},
		},
	}

	for name, reactors := range tests {
		t.Run(name, func(t *testing.T) {
			// Patch the functions not being tested
			monkey.PatchInstanceMethod(
				reflect.TypeOf(svcConf),
				"BuildKnativeServiceConfig",
				func(*KnativeService) *knservingv1alpha1.Service {
					return testKnSvc
				})
			monkey.Patch(knServiceSemanticEquals,
				func(*knservingv1alpha1.Service, *knservingv1alpha1.Service) bool {
					// Make method return false always, so that an update will be triggered
					return false
				})
			defer monkey.UnpatchAll()

			// Create test controller
			c := createTestKnController(cs, reactors)
			// Run test
			err := c.DeployKnativeService(context.Background(), svcConf)
			// Validate no error
			assert.NoError(t, err)
		})
	}
}

func TestDeployKubernetesService(t *testing.T) {
	testName, testNamespace := "test-name", "test-namespace"
	deploymentResourceItem := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
	svcResourceItem := schema.GroupVersionResource{
		Version:  "v1",
		Resource: "services",
	}
	testK8sSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testName,
			Namespace: testNamespace,
		},
	}
	testK8sDeployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      testName,
			Namespace: testNamespace,
		},
	}
	svcConf := &KubernetesService{
		BaseService: &BaseService{
			Name:      testName,
			Namespace: testNamespace,
		},
	}

	replicas := int32(1)
	// Define reactor for a successful get
	getDeploymentSuccess := func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{
				Name:       testName,
				Namespace:  testNamespace,
				Generation: 1,
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
			},
			Status: appsv1.DeploymentStatus{
				ObservedGeneration: 1,
				Replicas:           1,
				ReadyReplicas:      1,
				Conditions: []appsv1.DeploymentCondition{
					{
						Type: appsv1.DeploymentAvailable,
					},
				},
			},
		}, nil
	}
	getSvcSuccess := func(action k8stesting.Action) (bool, runtime.Object, error) {
		return true, testK8sSvc, nil
	}

	// Define tests
	cs := fake.NewSimpleClientset()
	tests := []struct {
		name     string
		reactors []reactor
	}{
		{"new_service",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: deploymentResourceItem.String(),
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(deploymentResourceItem, testNamespace, testName)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Return nil object and error to indicate non existent object
						return true, nil, k8serrors.NewNotFound(schema.GroupResource{}, testName)
					},
				},
				{
					verb:     reactorVerbs.Create,
					resource: "deployments",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewCreateAction(deploymentResourceItem, testNamespace, testK8sDeployment)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Prepend a new get reactor for waitK8sServiceReady to use
						cs.PrependReactor(reactorVerbs.Get, "deployments", getDeploymentSuccess)
						// Nil error indicates Create success
						return true, testK8sDeployment, nil
					},
				},
				{
					verb:     reactorVerbs.Get,
					resource: "services",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(svcResourceItem, testNamespace, testName)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Return nil object and error to indicate non existent object
						return true, nil, k8serrors.NewNotFound(schema.GroupResource{}, testName)
					},
				},
				{
					verb:     reactorVerbs.Create,
					resource: "services",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewCreateAction(svcResourceItem, testNamespace, testK8sSvc)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Prepend a new get reactor for waitKnativeServiceReady to use
						cs.PrependReactor(reactorVerbs.Get, "services", getSvcSuccess)
						// Nil error indicates Create success
						return true, testK8sSvc, nil
					},
				},
			},
		},
		{
			"update_service",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: "deployments",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(deploymentResourceItem, testNamespace, testName)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						return true, testK8sDeployment, nil
					},
				},
				{
					verb:     reactorVerbs.Update,
					resource: "deployments",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewUpdateAction(deploymentResourceItem,
							testNamespace, testK8sDeployment)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Prepend a new get reactor for waitK8sServiceReady to use
						cs.PrependReactor(reactorVerbs.Get, "deployments", getDeploymentSuccess)
						// Nil error indicates Update success
						return true, testK8sDeployment, nil
					},
				},
				{
					verb:     reactorVerbs.Get,
					resource: "services",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(svcResourceItem, testNamespace, testName)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Return nil object and error to indicate non existent object
						return true, testK8sSvc, k8serrors.NewNotFound(schema.GroupResource{}, testName)
					},
				},
				{
					verb:     reactorVerbs.Update,
					resource: "services",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewUpdateAction(svcResourceItem, testNamespace, testK8sSvc)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Nil error indicates Update success
						return true, testK8sSvc, nil
					},
				},
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Patch the functions not being tested
			monkey.PatchInstanceMethod(
				reflect.TypeOf(svcConf),
				"BuildKubernetesServiceConfig",
				func(service *KubernetesService) (*appsv1.Deployment, *corev1.Service) {
					return testK8sDeployment, testK8sSvc
				},
			)
			monkey.Patch(k8sServiceSemanticEquals,
				func(*corev1.Service, *corev1.Service) bool {
					// Make method return false always, so that an update will be triggered
					return false
				})
			monkey.Patch(k8sDeploymentSemanticEquals,
				func(*appsv1.Deployment, *appsv1.Deployment) bool {
					// Make method return false always, so that an update will be triggered
					return false
				})
			defer monkey.UnpatchAll()

			// Create test controller
			c := createTestK8sController(cs, tc.reactors)
			// Run test
			err := c.DeployKubernetesService(context.Background(), svcConf)
			// Validate no error
			assert.NoError(t, err)
		})
	}
}

func TestDeleteKnativeService(t *testing.T) {
	testName, testNamespace := "test-name", "test-namespace"
	resourceItem := schema.GroupVersionResource{
		Group:    knativeGroup,
		Version:  knativeVersion,
		Resource: knativeResource,
	}

	// Define reactors. The reactors also validate that they are called
	// with the expected parameters.
	reactors := []reactor{
		{
			verb:     reactorVerbs.Get,
			resource: knativeResource,
			rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
				expAction := k8stesting.NewGetAction(resourceItem, testNamespace, testName)
				// Check that the method is called with the expected action
				assert.Equal(t, expAction, action)
				// Nil error indicates Get success to the DeleteKnativeService() method
				return true, nil, nil
			},
		},
		{
			verb:     reactorVerbs.Delete,
			resource: knativeResource,
			rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
				expAction := k8stesting.NewDeleteAction(resourceItem, testNamespace, testName)
				// Check that the method is called with the expected action
				assert.Equal(t, expAction, action)
				// Nil error indicates Delete success
				return true, nil, nil
			},
		},
	}

	// Create test controller
	c := createTestKnController(knservingclientset.NewSimpleClientset(), reactors)
	// Run tests
	err := c.DeleteKnativeService(testName, testNamespace, time.Second*5)
	// Validate no error
	assert.NoError(t, err)
}

func TestDeleteK8sService(t *testing.T) {
	testName, testNamespace := "test-name", "test-namespace"
	deploymentResourceItem := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
	svcResourceItem := schema.GroupVersionResource{
		Version:  "v1",
		Resource: "services",
	}

	// Define reactors. The reactors also validate that they are called
	// with the expected parameters.
	reactors := []reactor{
		{
			verb:     reactorVerbs.Get,
			resource: "deployments",
			rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
				expAction := k8stesting.NewGetAction(deploymentResourceItem, testNamespace, testName)
				assert.Equal(t, expAction, action)
				return true, nil, nil
			},
		},
		{
			verb:     reactorVerbs.Delete,
			resource: "deployments",
			rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
				expAction := k8stesting.NewDeleteAction(deploymentResourceItem, testNamespace, testName)
				assert.Equal(t, expAction, action)
				return true, nil, nil
			},
		},
		{
			verb:     reactorVerbs.Get,
			resource: "services",
			rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
				expAction := k8stesting.NewGetAction(svcResourceItem, testNamespace, testName)
				assert.Equal(t, expAction, action)
				return true, nil, nil
			},
		},
		{
			verb:     reactorVerbs.Delete,
			resource: "services",
			rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
				expAction := k8stesting.NewDeleteAction(svcResourceItem, testNamespace, testName)
				assert.Equal(t, expAction, action)
				return true, nil, nil
			},
		},
	}

	// Create test controller
	c := createTestK8sController(fake.NewSimpleClientset(), reactors)
	// Run tests
	err := c.DeleteKubernetesService(testName, testNamespace, time.Second*5)
	// Validate no error
	assert.NoError(t, err)
}

func TestCreateNamespace(t *testing.T) {
	namespace := "test-ns"
	nsResource := schema.GroupVersionResource{
		Version:  "v1",
		Resource: "namespaces",
	}
	tests := []struct {
		name          string
		reactors      []reactor
		expectedError error
	}{
		{
			"namespace_exists",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: "namespaces",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expectedAction := k8stesting.NewGetAction(nsResource, "", namespace)
						assert.Equal(t, expectedAction, action)
						return true, nil, nil
					},
				},
			},
			errors.New("namespace already exists"),
		},
		{
			"create_namespace",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: "namespaces",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expectedAction := k8stesting.NewGetAction(nsResource, "", namespace)
						assert.Equal(t, expectedAction, action)
						return false, nil, k8serrors.NewNotFound(schema.GroupResource{}, namespace)
					},
				},
				{
					verb:     reactorVerbs.Create,
					resource: "namespaces",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expectedAction := k8stesting.NewCreateAction(nsResource, "", &corev1.Namespace{
							ObjectMeta: metav1.ObjectMeta{
								Name: namespace,
							},
						})
						assert.Equal(t, expectedAction, action)
						return false, nil, nil
					},
				},
			},
			nil,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cs := fake.NewSimpleClientset()
			for _, reactor := range tc.reactors {
				cs.PrependReactor(reactor.verb, reactor.resource, reactor.rFunc)
			}
			c := &controller{k8sCoreClient: cs.CoreV1()}
			err := c.CreateNamespace(namespace)
			if tc.expectedError == nil {
				// Validate no error
				assert.NoError(t, err)
			} else {
				assert.Error(t, err, tc.expectedError)
			}
		})
	}
}

func TestCreateConfigMap(t *testing.T) {
	namespace := "namespace"
	cmap := ConfigMap{
		Name:     "my-data",
		FileName: "key",
		Data:     "value",
	}
	k8scmap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmap.Name,
			Namespace: namespace,
		},
		Data: map[string]string{
			cmap.FileName: cmap.Data,
		},
	}
	cmResource := schema.GroupVersionResource{
		Version:  "v1",
		Resource: "configmaps",
	}
	tests := []struct {
		name     string
		reactors []reactor
	}{
		{
			"update_configmap",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: "configmaps",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expectedAction := k8stesting.NewGetAction(cmResource, namespace, cmap.Name)
						assert.Equal(t, expectedAction, action)
						return true, nil, nil
					},
				},
				{
					verb:     reactorVerbs.Update,
					resource: "configmaps",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expectedAction := k8stesting.NewUpdateAction(cmResource, namespace, &k8scmap)
						assert.Equal(t, expectedAction, action)
						return true, nil, nil
					},
				},
			},
		},
		{
			"create_configmap",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: "configmaps",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expectedAction := k8stesting.NewGetAction(cmResource, namespace, cmap.Name)
						assert.Equal(t, expectedAction, action)
						return true, nil, k8serrors.NewNotFound(schema.GroupResource{}, namespace)
					},
				},
				{
					verb:     reactorVerbs.Create,
					resource: "configmaps",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expectedAction := k8stesting.NewCreateAction(cmResource, namespace, &k8scmap)
						assert.Equal(t, expectedAction, action)
						return true, nil, nil
					},
				},
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cs := fake.NewSimpleClientset()
			for _, reactor := range tc.reactors {
				cs.PrependReactor(reactor.verb, reactor.resource, reactor.rFunc)
			}
			c := &controller{k8sCoreClient: cs.CoreV1()}
			err := c.ApplyConfigMap(namespace, &cmap)
			assert.NoError(t, err)
		})
	}
}

func TestCreateSecret(t *testing.T) {
	secretResource := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "secrets",
	}
	testNamespace := "namespace"

	secretConf := Secret{
		Name:      "secret",
		Namespace: testNamespace,
		Data: map[string]string{
			"key": "value",
		},
	}
	testSecret := secretConf.BuildSecret()
	cs := fake.NewSimpleClientset()
	tests := []struct {
		name     string
		reactors []reactor
		hasErr   bool
	}{
		{"new_secret",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: secretResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(secretResource, testNamespace, testSecret.Name)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Return nil object and error to indicate non existent object
						return true, nil, k8serrors.NewNotFound(schema.GroupResource{}, testSecret.Name)
					},
				},
				{
					verb:     reactorVerbs.Create,
					resource: secretResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewCreateAction(secretResource, testNamespace, testSecret)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Nil error indicates Create success
						return true, testSecret, nil
					},
				},
			},
			false,
		},
		{
			"secret_exists",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: secretResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(secretResource, testNamespace, testSecret.Name)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Return nil object and error to indicate non existent object
						return true, testSecret, nil
					},
				},
				{
					verb:     reactorVerbs.Update,
					resource: secretResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewUpdateAction(secretResource, testNamespace, testSecret)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Nil error indicates Create success
						return true, testSecret, nil
					},
				},
			},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create test controller
			c := createTestK8sController(cs, tc.reactors)
			// Run test
			err := c.CreateSecret(context.Background(), &secretConf)
			// Validate no error
			assert.Equal(t, tc.hasErr, err != nil)
		})
	}
}

func TestDeleteSecret(t *testing.T) {
	secretResource := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "secrets",
	}
	testNamespace := "namespace"

	secretName := "secret"
	cs := fake.NewSimpleClientset()
	tests := []struct {
		name     string
		reactors []reactor
		hasErr   bool
	}{
		{"not_exists",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: secretResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(secretResource, testNamespace, secretName)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Return nil object and error to indicate non existent object
						return true, nil, k8serrors.NewNotFound(schema.GroupResource{}, secretName)
					},
				},
			},
			true,
		},
		{
			"delete_secret",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: secretResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(secretResource, testNamespace, secretName)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						return true, nil, nil
					},
				},
				{
					verb:     reactorVerbs.Delete,
					resource: secretResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewDeleteAction(secretResource, testNamespace, secretName)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Return nil object and error to indicate non existent object
						return true, nil, nil
					},
				},
			},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create test controller
			c := createTestK8sController(cs, tc.reactors)
			// Run test
			err := c.DeleteSecret(secretName, testNamespace)
			// Validate no error
			assert.Equal(t, err != nil, tc.hasErr)
		})
	}
}

func TestCreatePVC(t *testing.T) {
	pvcResource := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "persistentvolumeclaims",
	}
	testNamespace := "namespace"

	cacheVolumeSize := "2Gi"
	volSize, _ := resource.ParseQuantity(cacheVolumeSize) // drop error since this volume size is a constant

	pvcConf := PersistentVolumeClaim{
		Name:        "test-svc-turing-pvc",
		AccessModes: []string{"ReadWriteOnce"},
		Size:        volSize,
	}
	testPvc := pvcConf.BuildPersistentVolumeClaim()
	cs := fake.NewSimpleClientset()
	tests := []struct {
		name     string
		reactors []reactor
		hasErr   bool
	}{
		{"new_pvc",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: pvcResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(pvcResource, testNamespace, testPvc.Name)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Return nil object and error to indicate non existent object
						return true, nil, k8serrors.NewNotFound(schema.GroupResource{}, testPvc.Name)
					},
				},
				{
					verb:     reactorVerbs.Create,
					resource: pvcResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewCreateAction(pvcResource, testNamespace, testPvc)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Nil error indicates Create success
						return true, testPvc, nil
					},
				},
			},
			false,
		},
		{
			"pvc_exists",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: pvcResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(pvcResource, testNamespace, testPvc.Name)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Return nil object and error to indicate non existent object
						return true, testPvc, nil
					},
				},
				{
					verb:     reactorVerbs.Update,
					resource: pvcResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewUpdateAction(pvcResource, testNamespace, testPvc)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Nil error indicates Create success
						return true, testPvc, nil
					},
				},
			},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create test controller
			c := createTestK8sController(cs, tc.reactors)
			// Run test
			err := c.ApplyPersistentVolumeClaim(context.Background(), testNamespace, &pvcConf)
			// Validate no error
			assert.Equal(t, tc.hasErr, err != nil)
		})
	}
}

func TestDeletePVC(t *testing.T) {
	pvcResource := schema.GroupVersionResource{
		Group:    "",
		Version:  "v1",
		Resource: "persistentvolumeclaims",
	}
	testNamespace := "namespace"

	cacheVolumeSize := "2Gi"
	volSize, _ := resource.ParseQuantity(cacheVolumeSize) // drop error since this volume size is a constant

	pvcConf := PersistentVolumeClaim{
		Name:        "test-svc-turing-pvc",
		AccessModes: []string{"ReadWriteOnce"},
		Size:        volSize,
	}
	cs := fake.NewSimpleClientset()
	tests := []struct {
		name     string
		reactors []reactor
		hasErr   bool
	}{
		{"not_exists",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: pvcResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(pvcResource, testNamespace, pvcConf.Name)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Return nil object and error to indicate non existent object
						return true, nil, k8serrors.NewNotFound(schema.GroupResource{}, pvcConf.Name)
					},
				},
			},
			true,
		},
		{
			"delete_pvc",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: pvcResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(pvcResource, testNamespace, pvcConf.Name)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						return true, nil, nil
					},
				},
				{
					verb:     reactorVerbs.Delete,
					resource: pvcResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewDeleteAction(pvcResource, testNamespace, pvcConf.Name)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Return nil object and error to indicate non existent object
						return true, nil, nil
					},
				},
			},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create test controller
			c := createTestK8sController(cs, tc.reactors)
			// Run test
			err := c.DeletePersistentVolumeClaim(pvcConf.Name, testNamespace)
			// Validate no error
			assert.Equal(t, tc.hasErr, err != nil)
		})
	}
}

func TestApplyIstioVirtualService(t *testing.T) {
	virtualServiceResource := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "virtualservices",
	}
	testNamespace := "namespace"

	vsConf := VirtualService{
		Name:      "test-svc-turing-router",
		Namespace: testNamespace,
		Endpoint:  "test-svc-turing-router.models.example.com",
	}
	testVs := vsConf.BuildVirtualService()
	cs := istioclientset.NewSimpleClientset()
	tests := []struct {
		name     string
		reactors []reactor
		hasErr   bool
	}{
		{"new_virtual_service",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: virtualServiceResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(virtualServiceResource, testNamespace, testVs.Name)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Return nil object and error to indicate non existent object
						return true, nil, k8serrors.NewNotFound(schema.GroupResource{}, testVs.Name)
					},
				},
				{
					verb:     reactorVerbs.Create,
					resource: virtualServiceResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewCreateAction(virtualServiceResource, testNamespace, testVs)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Nil error indicates Create success
						return true, testVs, nil
					},
				},
			},
			false,
		},
		{
			"virtual_service_exists",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: virtualServiceResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(virtualServiceResource, testNamespace, testVs.Name)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Return nil object and error to indicate non existent object
						return true, testVs, nil
					},
				},
				{
					verb:     reactorVerbs.Update,
					resource: virtualServiceResource.Resource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewUpdateAction(virtualServiceResource, testNamespace, testVs)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Nil error indicates Create success
						return true, testVs, nil
					},
				},
			},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create test controller
			c := createTestIstioController(cs, tc.reactors)
			// Run test
			err := c.ApplyIstioVirtualService(context.Background(), &vsConf)
			// Validate no error
			assert.Equal(t, tc.hasErr, err != nil)
		})
	}
}

func TestDeleteIstioVirtualService(t *testing.T) {
	virtualServiceResource := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1alpha3",
		Resource: "virtualservices",
	}
	testNamespace := "namespace"

	vsConf := VirtualService{
		Name:      "test-svc-turing-router",
		Namespace: testNamespace,
		Endpoint:  "test-svc-turing-router.models.example.com",
	}
	testVs := vsConf.BuildVirtualService()
	cs := istioclientset.NewSimpleClientset()

	reactors := []reactor{
		{
			verb:     reactorVerbs.Get,
			resource: virtualServiceResource.Resource,
			rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
				expAction := k8stesting.NewGetAction(virtualServiceResource, testNamespace, testVs.Name)
				// Check that the method is called with the expected action
				assert.Equal(t, expAction, action)
				// Return nil object and error to indicate non existent object
				return true, testVs, nil
			},
		},
		{
			verb:     reactorVerbs.Delete,
			resource: virtualServiceResource.Resource,
			rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
				expAction := k8stesting.NewDeleteAction(virtualServiceResource, testNamespace, testVs.Name)
				// Check that the method is called with the expected action
				assert.Equal(t, expAction, action)
				// Nil error indicates Create success
				return true, nil, nil
			},
		},
	}

	c := createTestIstioController(cs, reactors)
	// Run test
	err := c.DeleteIstioVirtualService(vsConf.Name, testNamespace, time.Second*5)
	// Validate no error
	assert.NoError(t, err)
}

func createTestKnController(cs *knservingclientset.Clientset, reactors []reactor) *controller {
	// Add reactors
	for _, reactor := range reactors {
		cs.PrependReactor(reactor.verb, reactor.resource, reactor.rFunc)
	}
	// Create clientset
	client := cs.ServingV1alpha1()
	// Return test controller with a fake knative serving client
	return &controller{knServingClient: client}
}

func createTestK8sController(cs *fake.Clientset, reactors []reactor) *controller {
	// Add reactors
	cs.ClearActions()
	for _, reactor := range reactors {
		cs.PrependReactor(reactor.verb, reactor.resource, reactor.rFunc)
	}
	// Return test controller with a fake knative serving client
	return &controller{
		k8sCoreClient: cs.CoreV1(),
		k8sAppsClient: cs.AppsV1(),
	}
}

func createTestIstioController(cs *istioclientset.Clientset, reactors []reactor) *controller {
	// Add reactors
	for _, reactor := range reactors {
		cs.PrependReactor(reactor.verb, reactor.resource, reactor.rFunc)
	}
	// Create clientset
	client := cs.NetworkingV1alpha3()
	// Return test controller with a fake knative serving client
	return &controller{istioClient: client}
}
