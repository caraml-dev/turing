package cluster

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	"k8s.io/apimachinery/pkg/api/resource"

	"bou.ke/monkey"

	sparkv1beta2 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	sparkOpFake "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/assert"
	istioclientset "istio.io/client-go/pkg/clientset/versioned/fake"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/fake"
	k8stesting "k8s.io/client-go/testing"
	"knative.dev/pkg/apis"
	duckv1 "knative.dev/pkg/apis/duck/v1"
	knservingv1 "knative.dev/serving/pkg/apis/serving/v1"
	knservingclientset "knative.dev/serving/pkg/client/clientset/versioned/fake"

	"github.com/caraml-dev/turing/api/turing/batch"
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
	knativeGroup           = "serving.knative.dev"
	knativeVersion         = "v1"
	knativeResource        = "services"
	contextTimeoutDuration = 15 * time.Second
)

func TestDeployKnativeService(t *testing.T) {
	testName, testNamespace := "test-name", "test-namespace"
	resourceItem := schema.GroupVersionResource{
		Group:    knativeGroup,
		Version:  knativeVersion,
		Resource: knativeResource,
	}
	testKnSvc := &knservingv1.Service{
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
		return true, &knservingv1.Service{
			ObjectMeta: metav1.ObjectMeta{
				Name: testName,
			},
			Status: knservingv1.ServiceStatus{
				Status: duckv1.Status{
					ObservedGeneration: 0,
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
				func(*KnativeService) (*knservingv1.Service, error) {
					return testKnSvc, nil
				})
			monkey.Patch(knServiceSemanticEquals,
				func(*knservingv1.Service, *knservingv1.Service) bool {
					// Make method return false always, so that an update will be triggered
					return false
				})
			defer monkey.UnpatchAll()

			// Create test controller
			c := createTestKnController(cs, reactors)

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()

			err := c.DeployKnativeService(ctx, svcConf)

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

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()

			// Run test
			err := c.DeployKubernetesService(ctx, svcConf)
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
	cs := knservingclientset.NewSimpleClientset()

	tests := []struct {
		name           string
		reactors       []reactor
		ignoreNotFound bool
		hasErr         bool
	}{
		{
			"not_exists; ignore knative resource not found",
			[]reactor{
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
			},
			true,
			false,
		},
		{
			"exists; ignore knative resource not found",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: knativeResource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(resourceItem, testNamespace, testName)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						return true, nil, nil
					},
				},
				{
					verb:     reactorVerbs.Delete,
					resource: knativeResource,
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewDeleteAction(resourceItem, testNamespace, testName)
						assert.Equal(t, expAction, action)
						return true, nil, nil
					},
				},
			},
			true,
			false,
		},
		{
			"not_exists; do not ignore knative resource not found",
			[]reactor{
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
			},
			false,
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()
			// Create test controller
			c := createTestKnController(cs, tc.reactors)
			// Run test
			err := c.DeleteKnativeService(ctx, testName, testNamespace, tc.ignoreNotFound)
			// Validate no error
			assert.Equal(t, err != nil, tc.hasErr)
		})
	}
}

func TestDeleteKubernetesDeployment(t *testing.T) {
	testName, testNamespace := "test-name", "test-namespace"
	deploymentResourceItem := schema.GroupVersionResource{
		Group:    "apps",
		Version:  "v1",
		Resource: "deployments",
	}
	cs := fake.NewSimpleClientset()

	tests := []struct {
		name           string
		reactors       []reactor
		ignoreNotFound bool
		hasErr         bool
	}{
		{
			"not_exists; ignore deployment not found",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: "deployments",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(deploymentResourceItem, testNamespace, testName)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Return nil object and error to indicate non existent object
						return true, nil, k8serrors.NewNotFound(schema.GroupResource{}, testName)
					},
				},
			},
			true,
			false,
		},
		{
			"exists; ignore deployment not found",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: "deployments",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(deploymentResourceItem, testNamespace, testName)
						// Check that the method is called with the expected action
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
			},
			true,
			false,
		},
		{
			"not_exists; do not ignore deployment not found",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: "deployments",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(deploymentResourceItem, testNamespace, testName)
						// Check that the method is called with the expected action
						assert.Equal(t, expAction, action)
						// Return nil object and error to indicate non existent object
						return true, nil, k8serrors.NewNotFound(schema.GroupResource{}, testName)
					},
				},
			},
			false,
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()
			// Create test controller
			c := createTestK8sController(cs, tc.reactors)
			// Run test
			err := c.DeleteKubernetesDeployment(ctx, testName, testNamespace, tc.ignoreNotFound)
			// Validate no error
			assert.Equal(t, err != nil, tc.hasErr)
		})
	}
}

func TestDeleteKubernetesService(t *testing.T) {
	testName, testNamespace := "test-name", "test-namespace"
	svcResourceItem := schema.GroupVersionResource{
		Version:  "v1",
		Resource: "services",
	}
	cs := fake.NewSimpleClientset()

	tests := []struct {
		name           string
		reactors       []reactor
		ignoreNotFound bool
		hasErr         bool
	}{
		{
			"not_exists; ignore service not found",
			[]reactor{
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
			},
			true,
			false,
		},
		{
			"exists; ignore service not found",
			[]reactor{
				{
					verb:     reactorVerbs.Get,
					resource: "services",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						expAction := k8stesting.NewGetAction(svcResourceItem, testNamespace, testName)
						// Check that the method is called with the expected action
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
			},
			true,
			false,
		},
		{
			"not_exists; do not ignore service not found",
			[]reactor{
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
			},
			false,
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()
			// Create test controller
			c := createTestK8sController(cs, tc.reactors)
			// Run test
			err := c.DeleteKubernetesService(ctx, testName, testNamespace, tc.ignoreNotFound)
			// Validate no error
			assert.Equal(t, err != nil, tc.hasErr)
		})
	}
}

func TestCreateKanikoJob(t *testing.T) {
	j := Job{
		Name:                    jobName,
		Namespace:               namespace,
		Labels:                  labels,
		Completions:             &jobCompletions,
		BackOffLimit:            &jobBackOffLimit,
		TTLSecondsAfterFinished: &jobTTLSecondAfterComplete,
		RestartPolicy:           corev1.RestartPolicyNever,
		Containers: []Container{
			CreateContainer(),
		},
		SecretVolumes: []SecretVolume{
			CreateSecretVolume(),
		},
	}

	t.Run("success | nominal flow", func(t *testing.T) {
		cs := fake.NewSimpleClientset()
		cs.PrependReactor("create", "jobs", func(action k8stesting.Action) (handled bool, ret runtime.Object, err error) {
			po := action.(k8stesting.CreateAction).GetObject().(*batchv1.Job)
			return true, &batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name: po.Name,
				},
			}, nil
		})
		c := &controller{
			k8sBatchClient: cs.BatchV1(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
		defer cancel()

		job, err := c.CreateJob(ctx, namespace, j)
		assert.Nil(t, err)
		assert.NotNil(t, job)
	})
}

func TestGetJob(t *testing.T) {
	namespace := "test-ns"
	jobName := "bicycle"
	tests := map[string]struct {
		reactors []reactor
		jobNil   bool
		errNil   bool
	}{
		"failure | no such job": {
			reactors: []reactor{
				{
					verb:     reactorVerbs.Get,
					resource: "job",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						return true, nil, k8serrors.NewNotFound(schema.GroupResource{}, jobName)
					},
				},
			},
			jobNil: true,
			errNil: false,
		},
		"success | job exists": {
			reactors: []reactor{
				{
					verb:     reactorVerbs.Get,
					resource: "jobs",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						return true, &batchv1.Job{
							ObjectMeta: metav1.ObjectMeta{
								Name: jobName,
							},
						}, nil
					},
				},
			},
			jobNil: false,
			errNil: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cs := fake.NewSimpleClientset()
			for _, reactor := range tt.reactors {
				cs.PrependReactor(reactor.verb, reactor.resource, reactor.rFunc)
			}
			c := &controller{
				k8sBatchClient: cs.BatchV1(),
			}

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()

			job, err := c.GetJob(ctx, namespace, jobName)
			assert.True(t, (job == nil) == tt.jobNil)
			assert.True(t, (err == nil) == tt.errNil)
		})
	}
}

func TestDeleteJob(t *testing.T) {
	namespace := "test-ns"
	jobName := "bicycle"
	tests := map[string]struct {
		reactors []reactor
		jobNil   bool
		errNil   bool
	}{
		"failure | no such job": {
			reactors: []reactor{
				{
					verb:     reactorVerbs.Delete,
					resource: "job",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						return true, nil, k8serrors.NewNotFound(schema.GroupResource{}, jobName)
					},
				},
			},
			errNil: false,
		},
		"success | job exists": {
			reactors: []reactor{
				{
					verb:     reactorVerbs.Delete,
					resource: "jobs",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						return true, nil, nil
					},
				},
			},
			errNil: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cs := fake.NewSimpleClientset()
			for _, reactor := range tt.reactors {
				cs.PrependReactor(reactor.verb, reactor.resource, reactor.rFunc)
			}
			c := &controller{
				k8sBatchClient: cs.BatchV1(),
			}

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()

			err := c.DeleteJob(ctx, namespace, jobName)
			assert.True(t, (err == nil) == tt.errNil)
		})
	}
}

func TestCreateServiceAccount(t *testing.T) {
	namespace := "test-ns"
	serviceAccountName := "bicycle"
	labels := map[string]string{"key": "val"}
	tests := map[string]struct {
		reactors  []reactor
		errNil    bool
		svcAccNil bool
	}{
		"success | service account exists": {
			reactors: []reactor{
				{
					verb:     reactorVerbs.Get,
					resource: "serviceaccount",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						return true, &corev1.ServiceAccount{
							ObjectMeta: metav1.ObjectMeta{
								Name: serviceAccountName,
							},
						}, nil
					},
				},
			},
			errNil:    true,
			svcAccNil: false,
		},
		"success | service account created": {
			reactors: []reactor{
				{
					verb:     reactorVerbs.Create,
					resource: "serviceaccount",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						return true, &corev1.ServiceAccount{
							ObjectMeta: metav1.ObjectMeta{
								Name: serviceAccountName,
							},
						}, nil
					},
				},
			},
			errNil:    true,
			svcAccNil: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cs := fake.NewSimpleClientset()
			for _, reactor := range tt.reactors {
				cs.PrependReactor(reactor.verb, reactor.resource, reactor.rFunc)
			}
			c := &controller{
				k8sBatchClient: cs.BatchV1(),
				k8sRBACClient:  cs.RbacV1(),
				k8sCoreClient:  cs.CoreV1(),
			}
			saCfg := &ServiceAccount{
				Name:      serviceAccountName,
				Namespace: namespace,
				Labels:    labels,
			}

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()

			svcAcc, err := c.CreateServiceAccount(ctx, namespace, saCfg)
			assert.True(t, (err == nil) == tt.errNil)
			assert.True(t, (svcAcc == nil) == tt.svcAccNil)
		})
	}
}

func TestCreateRole(t *testing.T) {
	namespace := "test-ns"
	roleName := "bicycle"
	labels := map[string]string{"key": "val"}
	tests := map[string]struct {
		reactors []reactor
		errNil   bool
		roleNil  bool
	}{
		"success | role exists": {
			reactors: []reactor{
				{
					verb:     reactorVerbs.Get,
					resource: "role",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						return true, &rbacv1.Role{
							ObjectMeta: metav1.ObjectMeta{
								Name: roleName,
							},
						}, nil
					},
				},
			},
			errNil:  true,
			roleNil: false,
		},
		"success | role created": {
			reactors: []reactor{
				{
					verb:     reactorVerbs.Create,
					resource: "role",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						return true, &rbacv1.Role{
							ObjectMeta: metav1.ObjectMeta{
								Name: roleName,
							},
						}, nil
					},
				},
			},
			errNil:  true,
			roleNil: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cs := fake.NewSimpleClientset()
			for _, reactor := range tt.reactors {
				cs.PrependReactor(reactor.verb, reactor.resource, reactor.rFunc)
			}
			c := &controller{
				k8sBatchClient: cs.BatchV1(),
				k8sRBACClient:  cs.RbacV1(),
				k8sCoreClient:  cs.CoreV1(),
			}
			roleCfg := &Role{
				Name:      roleName,
				Namespace: namespace,
				Labels:    labels,
			}

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()

			role, err := c.CreateRole(ctx, namespace, roleCfg)
			assert.True(t, (err == nil) == tt.errNil)
			assert.True(t, (role == nil) == tt.roleNil)
		})
	}
}

func TestCreateRoleBinding(t *testing.T) {
	namespace := "test-ns"
	roleName := "bicycle"
	roleBindingName := "wd-40"
	serviceAccountName := "bicycle-shop"
	labels := map[string]string{"key": "val"}
	tests := map[string]struct {
		reactors []reactor
		errNil   bool
		roleNil  bool
	}{
		"success | service account exists": {
			reactors: []reactor{
				{
					verb:     reactorVerbs.Get,
					resource: "rolebinding",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						return true, &rbacv1.RoleBinding{
							ObjectMeta: metav1.ObjectMeta{
								Name: roleBindingName,
							},
						}, nil
					},
				},
			},
			errNil:  true,
			roleNil: false,
		},
		"success | service account created": {
			reactors: []reactor{
				{
					verb:     reactorVerbs.Create,
					resource: "rolebinding",
					rFunc: func(action k8stesting.Action) (bool, runtime.Object, error) {
						return true, &rbacv1.RoleBinding{
							ObjectMeta: metav1.ObjectMeta{
								Name: roleBindingName,
							},
						}, nil
					},
				},
			},
			errNil:  true,
			roleNil: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cs := fake.NewSimpleClientset()
			for _, reactor := range tt.reactors {
				cs.PrependReactor(reactor.verb, reactor.resource, reactor.rFunc)
			}
			c := &controller{
				k8sBatchClient: cs.BatchV1(),
				k8sRBACClient:  cs.RbacV1(),
				k8sCoreClient:  cs.CoreV1(),
			}
			roleBindingCfg := &RoleBinding{
				Name:               roleBindingName,
				Namespace:          namespace,
				Labels:             labels,
				RoleName:           roleName,
				ServiceAccountName: serviceAccountName,
			}

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()

			role, err := c.CreateRoleBinding(ctx, namespace, roleBindingCfg)
			assert.True(t, (err == nil) == tt.errNil)
			assert.True(t, (role == nil) == tt.roleNil)
		})
	}
}

func TestCreateSparkApplication(t *testing.T) {
	namespace := "test-ci"
	t.Run("success | nominal", func(t *testing.T) {
		cs := fake.NewSimpleClientset()
		cs.PrependReactor(
			reactorVerbs.Create,
			"sparkapplication",
			func(action k8stesting.Action) (bool, runtime.Object, error) {
				return true, &sparkv1beta2.SparkApplication{
					ObjectMeta: metav1.ObjectMeta{
						Name: "spark",
					},
				}, nil
			},
		)
		sparkClientSet := sparkOpFake.Clientset{}
		c := &controller{
			k8sBatchClient:   cs.BatchV1(),
			k8sRBACClient:    cs.RbacV1(),
			k8sCoreClient:    cs.CoreV1(),
			k8sSparkOperator: sparkClientSet.SparkoperatorV1beta2(),
		}
		req := &CreateSparkRequest{
			JobName:               jobName,
			JobLabels:             jobLabels,
			JobImageRef:           jobImageRef,
			JobApplicationPath:    jobApplicationPath,
			JobArguments:          jobArguments,
			JobConfigMount:        batch.JobConfigMount,
			DriverCPURequest:      cpuValue,
			DriverMemoryRequest:   memoryValue,
			ExecutorCPURequest:    cpuValue,
			ExecutorMemoryRequest: memoryValue,
			ExecutorReplica:       executorReplica,
			ServiceAccountName:    serviceAccountName,
			SparkInfraConfig:      sparkInfraConfig,
		}

		ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
		defer cancel()

		sparkApp, err := c.CreateSparkApplication(ctx, namespace, req)
		assert.NotNil(t, sparkApp)
		assert.Nil(t, err)
	})
}

func TestGetSparkApplication(t *testing.T) {
	namespace := "test-ns"
	appName := "bicycle"
	t.Run("Success | nominal", func(t *testing.T) {
		cs := fake.NewSimpleClientset()
		cs.PrependReactor(
			reactorVerbs.Get,
			"sparkapplication",
			func(action k8stesting.Action) (bool, runtime.Object, error) {
				return true, &sparkv1beta2.SparkApplication{
					ObjectMeta: metav1.ObjectMeta{
						Name: "spark",
					},
				}, nil
			},
		)
		sparkClientSet := sparkOpFake.Clientset{}
		c := &controller{
			k8sSparkOperator: sparkClientSet.SparkoperatorV1beta2(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
		defer cancel()

		app, err := c.GetSparkApplication(ctx, namespace, appName)
		assert.NotNil(t, app)
		assert.Nil(t, err)
	})
}

func TestDeleteSparkApplication(t *testing.T) {
	namespace := "test-ns"
	appName := "bicycle"
	t.Run("Success | nominal", func(t *testing.T) {
		cs := fake.NewSimpleClientset()
		cs.PrependReactor(
			reactorVerbs.Get,
			"sparkapplication",
			func(action k8stesting.Action) (bool, runtime.Object, error) {
				return true, nil, nil
			},
		)
		sparkClientSet := sparkOpFake.Clientset{}
		c := &controller{
			k8sSparkOperator: sparkClientSet.SparkoperatorV1beta2(),
		}

		ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
		defer cancel()

		err := c.DeleteSparkApplication(ctx, namespace, appName)
		assert.Nil(t, err)
	})
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

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()

			err := c.CreateNamespace(ctx, namespace)
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
	labels := map[string]string{"key": "value"}
	cmap := ConfigMap{
		Name:     "my-data",
		FileName: "key",
		Data:     "value",
		Labels:   labels,
	}
	k8scmap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmap.Name,
			Namespace: namespace,
			Labels:    labels,
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

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()

			c := &controller{k8sCoreClient: cs.CoreV1()}
			err := c.ApplyConfigMap(ctx, namespace, &cmap)
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
		Labels: map[string]string{
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

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()

			// Run test
			err := c.CreateSecret(ctx, &secretConf)
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
		name           string
		reactors       []reactor
		ignoreNotFound bool
		hasErr         bool
	}{
		{
			"not_exists; ignore secret not found",
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
			false,
		},
		{
			"exists; ignore secret not found",
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
						assert.Equal(t, expAction, action)
						return true, nil, nil
					},
				},
			},
			true,
			false,
		},
		{
			"not_exists; do not ignore secret not found",
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
			false,
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create test controller
			c := createTestK8sController(cs, tc.reactors)

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()

			// Run test
			err := c.DeleteSecret(ctx, secretName, testNamespace, tc.ignoreNotFound)
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

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()

			// Run test
			err := c.ApplyPersistentVolumeClaim(ctx, testNamespace, &pvcConf)
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
		name           string
		reactors       []reactor
		ignoreNotFound bool
		hasErr         bool
	}{
		{
			"not_exists; ignore pvc not found",
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
			false,
		},
		{
			"exists; ignore pvc not found",
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
						assert.Equal(t, expAction, action)
						return true, nil, nil
					},
				},
			},
			true,
			false,
		},
		{
			"not_exists; do not ignore pvc not found",
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
			false,
			true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Create test controller
			c := createTestK8sController(cs, tc.reactors)

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()

			// Run test
			err := c.DeletePersistentVolumeClaim(ctx, pvcConf.Name, testNamespace, tc.ignoreNotFound)
			// Validate no error
			assert.Equal(t, tc.hasErr, err != nil)
		})
	}
}

func TestApplyIstioVirtualService(t *testing.T) {
	virtualServiceResource := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1beta1",
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

			ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
			defer cancel()

			// Run test
			err := c.ApplyIstioVirtualService(ctx, &vsConf)
			// Validate no error
			assert.Equal(t, tc.hasErr, err != nil)
		})
	}
}

func TestDeleteIstioVirtualService(t *testing.T) {
	virtualServiceResource := schema.GroupVersionResource{
		Group:    "networking.istio.io",
		Version:  "v1beta1",
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

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
	defer cancel()

	// Run test
	err := c.DeleteIstioVirtualService(ctx, vsConf.Name, testNamespace)
	// Validate no error
	assert.NoError(t, err)
}

func TestGetKnativePodTerminationMessage(t *testing.T) {
	testNamespace := "test-ns"
	testName := "test-name"

	// Create test controller
	c := createTestK8sController(fake.NewSimpleClientset(&corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("pod-%s", testName),
			Namespace: testNamespace,
			Labels: map[string]string{
				"serving.knative.dev/service": "test-name",
			},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{
					Name: "dummy-container",
					LastTerminationState: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							Message: "Test Dummy Message",
						},
					},
				},
				{
					Name: "user-container",
					LastTerminationState: corev1.ContainerState{
						Terminated: &corev1.ContainerStateTerminated{
							Message: "Test Termination Message",
						},
					},
				},
			},
		},
	}), []reactor{})

	ctx, cancel := context.WithTimeout(context.Background(), contextTimeoutDuration)
	defer cancel()

	// Run test
	msg := c.getKnativePodTerminationMessage(ctx, testName, testNamespace)
	assert.Equal(t, "Test Termination Message", msg)
}

func TestGetKnServiceStatusSummary(t *testing.T) {
	svc := &knservingv1.Service{
		Status: knservingv1.ServiceStatus{
			Status: duckv1.Status{
				ObservedGeneration: 1,
				Conditions: duckv1.Conditions{
					apis.Condition{
						Type:   apis.ConditionReady,
						Status: corev1.ConditionTrue,
					},
					apis.Condition{
						Type:    apis.ConditionSucceeded,
						Status:  corev1.ConditionFalse,
						Message: "Test Message",
					},
				},
			},
		},
	}

	expectedSummary := "Type: Ready, Status: true. \nType: Succeeded, Status: false. Test Message"
	summary := getKnServiceStatusMessages(svc)
	assert.Equal(t, expectedSummary, summary)
}

func createTestKnController(cs *knservingclientset.Clientset, reactors []reactor) *controller {
	// Add reactors
	for _, reactor := range reactors {
		cs.PrependReactor(reactor.verb, reactor.resource, reactor.rFunc)
	}
	// Create clientset
	client := cs.ServingV1()
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
	client := cs.NetworkingV1beta1()
	// Return test controller with a fake knative serving client
	return &controller{istioClient: client}
}
