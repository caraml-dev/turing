package batchensembling

import (
	"fmt"
	"testing"

	"github.com/caraml-dev/turing/api/turing/batch"

	apisparkv1beta2 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	apicorev1 "k8s.io/api/core/v1"
	apirbacv1 "k8s.io/api/rbac/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caraml-dev/turing/api/turing/cluster"
	clustermock "github.com/caraml-dev/turing/api/turing/cluster/mocks"
	"github.com/caraml-dev/turing/api/turing/config"
	openapi "github.com/caraml-dev/turing/api/turing/generated"
	"github.com/caraml-dev/turing/api/turing/internal/ref"
	"github.com/caraml-dev/turing/api/turing/models"
	"github.com/caraml-dev/turing/api/turing/service"
	servicemock "github.com/caraml-dev/turing/api/turing/service/mocks"
)

const (
	imageRef  = "gojek/testimage:1"
	namespace = "test-namespace"
)

var (
	standardLabels = map[string]string{
		"model": "T800",
	}
	tolerationName                          = "batch-job"
	sparkInfraConfig *config.SparkAppConfig = &config.SparkAppConfig{
		NodeSelector: map[string]string{
			"node-workload-type": "batch",
		},
		CorePerCPURequest:              1.5,
		CPURequestToCPULimit:           1.25,
		SparkVersion:                   "2.4.5",
		TolerationName:                 &tolerationName,
		SubmissionFailureRetries:       3,
		SubmissionFailureRetryInterval: 10,
		FailureRetries:                 3,
		FailureRetryInterval:           10,
		PythonVersion:                  "3",
		TTLSecond:                      86400,
	}
)

func generateEnsemblingJobFixture() *models.EnsemblingJob {
	nullableEnsemblingResources := openapi.NullableEnsemblingResources{}
	nullableEnsemblingResources.Set(&openapi.EnsemblingResources{
		DriverCpuRequest:      ref.String("1"),
		DriverMemoryRequest:   ref.String("1Gi"),
		ExecutorReplica:       ref.Int32(10),
		ExecutorCpuRequest:    ref.String("1"),
		ExecutorMemoryRequest: ref.String("1Gi"),
	})

	return &models.EnsemblingJob{
		Name:            "test-ensembler-1",
		ProjectID:       models.ID(1),
		EnsemblerID:     models.ID(1),
		EnvironmentName: "dev",
		InfraConfig: &models.InfraConfig{
			EnsemblerInfraConfig: openapi.EnsemblerInfraConfig{
				RunId:              ref.String("abc123"),
				ArtifactUri:        ref.String("gs://bucket/ensembler"),
				EnsemblerName:      ref.String("ensembler"),
				Resources:          nullableEnsemblingResources,
				ServiceAccountName: ref.String("test-service-account"),
			},
		},
		JobConfig: &models.JobConfig{
			Version: "v1",
			Kind:    openapi.ENSEMBLERCONFIGKIND_BATCH_ENSEMBLING_JOB,
			Metadata: &openapi.EnsemblingJobMeta{
				Name:        "test-batch-ensembling",
				Annotations: map[string]string{},
			},
			Spec: openapi.EnsemblingJobSpec{
				Source: openapi.EnsemblingJobSource{
					Dataset: openapi.Dataset{
						BigQueryDataset: &openapi.BigQueryDataset{
							Type: batch.DatasetTypeBQ,
							BqConfig: openapi.BigQueryDatasetConfig{
								Query: ref.String("select * from hello_world where customer_id = 4"),
								Options: map[string]string{
									"viewsEnabled":           "true",
									"materializationDataset": "dataset",
								},
							},
						},
					},
					JoinOn: []string{"customer_id", "target_date"},
				},
				Predictions: map[string]openapi.EnsemblingJobPredictionSource{
					"model_a": {
						Dataset: openapi.Dataset{
							BigQueryDataset: &openapi.BigQueryDataset{
								Type: batch.DatasetTypeBQ,
								BqConfig: openapi.BigQueryDatasetConfig{
									Table: ref.String("project.dataset.predictions_model_a"),
									Features: []string{
										"customer_id",
										"target_date",
										"predictions",
									},
								},
							},
						},
						Columns: []string{"predictions"},
						JoinOn:  []string{"customer_id", "target_date"},
					},
					"model_b": {
						Dataset: openapi.Dataset{
							BigQueryDataset: &openapi.BigQueryDataset{
								Type: batch.DatasetTypeBQ,
								BqConfig: openapi.BigQueryDatasetConfig{
									Query: ref.String("select * from helloworld where customer_id = 3"),
								},
							},
						},
						Columns: []string{"predictions"},
						JoinOn:  []string{"customer_id", "target_date"},
					},
				},
				Ensembler: openapi.EnsemblingJobEnsemblerSpec{
					Uri: "/home/spark/ensembler",
					Result: openapi.EnsemblingJobEnsemblerSpecResult{
						ColumnName: "prediction_score",
						Type:       openapi.ENSEMBLINGJOBRESULTTYPE_FLOAT,
						ItemType:   ref.EnsemblingJobResultType(openapi.ENSEMBLINGJOBRESULTTYPE_FLOAT),
					},
				},
				Sink: openapi.EnsemblingJobSink{
					BigQuerySink: &openapi.BigQuerySink{
						Type: batch.SinkTypeBQ,
						Columns: []string{
							"customer_id as customerId",
							"target_date",
							"results",
						},
						SaveMode: openapi.SAVEMODE_OVERWRITE,
						BqConfig: openapi.BigQuerySinkConfig{
							Table:         "project.dataset.ensembling_results",
							StagingBucket: "bucket-name",
							Options: map[string]string{
								"partitionField": "target_date",
							},
						},
					},
				},
			},
		},
	}
}

func TestCreate(t *testing.T) {
	tests := map[string]struct {
		expected          error
		clusterController func() cluster.Controller
		mlpService        func() service.MLPService
		request           *CreateEnsemblingJobRequest
	}{
		"success | nominal": {
			expected: nil,
			clusterController: func() cluster.Controller {
				ctrler := &clustermock.Controller{}
				ctrler.On("CreateNamespace", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("CreateServiceAccount", mock.Anything, mock.Anything, mock.Anything).Return(
					&apicorev1.ServiceAccount{
						ObjectMeta: apimetav1.ObjectMeta{
							Name: fmt.Sprintf("%s-driver-sa", namespace),
						},
					},
					nil,
				)
				ctrler.On("CreateRole", mock.Anything, mock.Anything, mock.Anything).Return(
					&apirbacv1.Role{
						ObjectMeta: apimetav1.ObjectMeta{
							Name: fmt.Sprintf("%s-driver-role", namespace),
						},
					},
					nil,
				)
				ctrler.On(
					"CreateRoleBinding",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)
				ctrler.On("CreateSecret", mock.Anything, mock.Anything).Return(nil, nil)
				ctrler.On("ApplyConfigMap", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				ctrler.On("CreateSparkApplication", mock.Anything, mock.Anything, mock.Anything).Return(nil, nil)
				return ctrler
			},
			mlpService: func() service.MLPService {
				svc := &servicemock.MLPService{}
				svc.On(
					"GetSecret",
					mock.Anything,
					mock.Anything,
				).Return("Alright then. Keep your secrets.", nil)
				return svc
			},
			request: &CreateEnsemblingJobRequest{
				EnsemblingJob: generateEnsemblingJobFixture(),
				Labels:        standardLabels,
				ImageRef:      imageRef,
				Namespace:     namespace,
			},
		},
		"failure | fail to create namespace": {
			expected: fmt.Errorf("failed creating namespace %s: %v", namespace, fmt.Errorf("hi")),
			clusterController: func() cluster.Controller {
				ctrler := &clustermock.Controller{}
				ctrler.On("CreateNamespace", mock.Anything, mock.Anything).Return(fmt.Errorf("hi"))
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				return ctrler
			},
			mlpService: func() service.MLPService {
				svc := &servicemock.MLPService{}
				svc.On(
					"GetSecret",
					mock.Anything,
					mock.Anything,
				).Return("Alright then. Keep your secrets.", nil)
				return svc
			},
			request: &CreateEnsemblingJobRequest{
				EnsemblingJob: generateEnsemblingJobFixture(),
				Labels:        standardLabels,
				ImageRef:      imageRef,
				Namespace:     namespace,
			},
		},
		"failure | fail to create authorization: fail service account creation": {
			expected: fmt.Errorf("failed creating spark driver authorization in namespace %s: %v",
				namespace,
				fmt.Errorf("hi"),
			),
			clusterController: func() cluster.Controller {
				ctrler := &clustermock.Controller{}
				ctrler.On("CreateNamespace", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("CreateServiceAccount", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
					fmt.Errorf("hi"),
				)
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				return ctrler
			},
			mlpService: func() service.MLPService {
				svc := &servicemock.MLPService{}
				svc.On(
					"GetSecret",
					mock.Anything,
					mock.Anything,
				).Return("Alright then. Keep your secrets.", nil)
				return svc
			},
			request: &CreateEnsemblingJobRequest{
				EnsemblingJob: generateEnsemblingJobFixture(),
				Labels:        standardLabels,
				ImageRef:      imageRef,
				Namespace:     namespace,
			},
		},
		"failure | fail to create authorization: fail role creation": {
			expected: fmt.Errorf("failed creating spark driver authorization in namespace %s: %v",
				namespace,
				fmt.Errorf("hi"),
			),
			clusterController: func() cluster.Controller {
				ctrler := &clustermock.Controller{}
				ctrler.On("CreateNamespace", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("CreateServiceAccount", mock.Anything, mock.Anything, mock.Anything).Return(
					&apicorev1.ServiceAccount{
						ObjectMeta: apimetav1.ObjectMeta{
							Name: fmt.Sprintf("%s-driver-sa", namespace),
						},
					},
					nil,
				)
				ctrler.On("CreateRole", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
					fmt.Errorf("hi"),
				)
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				return ctrler
			},
			mlpService: func() service.MLPService {
				svc := &servicemock.MLPService{}
				svc.On(
					"GetSecret",
					mock.Anything,
					mock.Anything,
				).Return("Alright then. Keep your secrets.", nil)
				return svc
			},
			request: &CreateEnsemblingJobRequest{
				EnsemblingJob: generateEnsemblingJobFixture(),
				Labels:        standardLabels,
				ImageRef:      imageRef,
				Namespace:     namespace,
			},
		},
		"failure | fail to create authorization: fail role binding creation": {
			expected: fmt.Errorf("failed creating spark driver authorization in namespace %s: %v",
				namespace,
				fmt.Errorf("hi"),
			),
			clusterController: func() cluster.Controller {
				ctrler := &clustermock.Controller{}
				ctrler.On("CreateNamespace", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("CreateServiceAccount", mock.Anything, mock.Anything, mock.Anything).Return(
					&apicorev1.ServiceAccount{
						ObjectMeta: apimetav1.ObjectMeta{
							Name: fmt.Sprintf("%s-driver-sa", namespace),
						},
					},
					nil,
				)
				ctrler.On("CreateRole", mock.Anything, mock.Anything, mock.Anything).Return(
					&apirbacv1.Role{
						ObjectMeta: apimetav1.ObjectMeta{
							Name: fmt.Sprintf("%s-driver-role", namespace),
						},
					},
					nil,
				)
				ctrler.On(
					"CreateRoleBinding",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, fmt.Errorf("hi"))
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				return ctrler
			},
			mlpService: func() service.MLPService {
				svc := &servicemock.MLPService{}
				svc.On(
					"GetSecret",
					mock.Anything,
					mock.Anything,
				).Return("Alright then. Keep your secrets.", nil)
				return svc
			},
			request: &CreateEnsemblingJobRequest{
				EnsemblingJob: generateEnsemblingJobFixture(),
				Labels:        standardLabels,
				ImageRef:      imageRef,
				Namespace:     namespace,
			},
		},
		"failure | fail to get secret": {
			expected: fmt.Errorf("service account %s is not found within %s project: %s",
				"test-service-account",
				namespace,
				fmt.Errorf("hi"),
			),
			clusterController: func() cluster.Controller {
				ctrler := &clustermock.Controller{}
				ctrler.On("CreateNamespace", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("CreateServiceAccount", mock.Anything, mock.Anything, mock.Anything).Return(
					&apicorev1.ServiceAccount{
						ObjectMeta: apimetav1.ObjectMeta{
							Name: fmt.Sprintf("%s-driver-sa", namespace),
						},
					},
					nil,
				)
				ctrler.On("CreateRole", mock.Anything, mock.Anything, mock.Anything).Return(
					&apirbacv1.Role{
						ObjectMeta: apimetav1.ObjectMeta{
							Name: fmt.Sprintf("%s-driver-role", namespace),
						},
					},
					nil,
				)
				ctrler.On(
					"CreateRoleBinding",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				return ctrler
			},
			mlpService: func() service.MLPService {
				svc := &servicemock.MLPService{}
				svc.On(
					"GetSecret",
					mock.Anything,
					mock.Anything,
				).Return("", fmt.Errorf("hi"))
				return svc
			},
			request: &CreateEnsemblingJobRequest{
				EnsemblingJob: generateEnsemblingJobFixture(),
				Labels:        standardLabels,
				ImageRef:      imageRef,
				Namespace:     namespace,
			},
		},
		"failure | fail to create job config map": {
			expected: fmt.Errorf("failed creating job specification configmap for job %s in namespace %s: %v",
				"test-ensembler-1",
				namespace,
				fmt.Errorf("hi"),
			),
			clusterController: func() cluster.Controller {
				ctrler := &clustermock.Controller{}
				ctrler.On("CreateNamespace", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("CreateServiceAccount", mock.Anything, mock.Anything, mock.Anything).Return(
					&apicorev1.ServiceAccount{
						ObjectMeta: apimetav1.ObjectMeta{
							Name: fmt.Sprintf("%s-driver-sa", namespace),
						},
					},
					nil,
				)
				ctrler.On("CreateRole", mock.Anything, mock.Anything, mock.Anything).Return(
					&apirbacv1.Role{
						ObjectMeta: apimetav1.ObjectMeta{
							Name: fmt.Sprintf("%s-driver-role", namespace),
						},
					},
					nil,
				)
				ctrler.On(
					"CreateRoleBinding",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)
				ctrler.On("CreateSecret", mock.Anything, mock.Anything).Return(nil, nil)
				ctrler.On("ApplyConfigMap", mock.Anything, mock.Anything, mock.Anything).Return(fmt.Errorf("hi"))
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				return ctrler
			},
			mlpService: func() service.MLPService {
				svc := &servicemock.MLPService{}
				svc.On(
					"GetSecret",
					mock.Anything,
					mock.Anything,
				).Return("Alright then. Keep your secrets.", nil)
				return svc
			},
			request: &CreateEnsemblingJobRequest{
				EnsemblingJob: generateEnsemblingJobFixture(),
				Labels:        standardLabels,
				ImageRef:      imageRef,
				Namespace:     namespace,
			},
		},
		"failure | fail to create spark application": {
			expected: fmt.Errorf(
				"failed submitting spark application to spark controller for job %s in namespace %s: %v",
				"test-ensembler-1",
				namespace,
				fmt.Errorf("hi"),
			),
			clusterController: func() cluster.Controller {
				ctrler := &clustermock.Controller{}
				ctrler.On("CreateNamespace", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("CreateServiceAccount", mock.Anything, mock.Anything, mock.Anything).Return(
					&apicorev1.ServiceAccount{
						ObjectMeta: apimetav1.ObjectMeta{
							Name: fmt.Sprintf("%s-driver-sa", namespace),
						},
					},
					nil,
				)
				ctrler.On("CreateRole", mock.Anything, mock.Anything, mock.Anything).Return(
					&apirbacv1.Role{
						ObjectMeta: apimetav1.ObjectMeta{
							Name: fmt.Sprintf("%s-driver-role", namespace),
						},
					},
					nil,
				)
				ctrler.On(
					"CreateRoleBinding",
					mock.Anything,
					mock.Anything,
					mock.Anything,
				).Return(nil, nil)
				ctrler.On("CreateSecret", mock.Anything, mock.Anything).Return(nil, nil)
				ctrler.On("ApplyConfigMap", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				ctrler.On("CreateSparkApplication", mock.Anything, mock.Anything, mock.Anything).Return(nil, fmt.Errorf("hi"))
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				return ctrler
			},
			mlpService: func() service.MLPService {
				svc := &servicemock.MLPService{}
				svc.On(
					"GetSecret",
					mock.Anything,
					mock.Anything,
				).Return("Alright then. Keep your secrets.", nil)
				return svc
			},
			request: &CreateEnsemblingJobRequest{
				EnsemblingJob: generateEnsemblingJobFixture(),
				Labels:        standardLabels,
				ImageRef:      imageRef,
				Namespace:     namespace,
			},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clusterController := tt.clusterController()
			mlpSvc := tt.mlpService()
			ensemblingController := NewBatchEnsemblingController(
				clusterController,
				mlpSvc,
				sparkInfraConfig,
			)
			err := ensemblingController.Create(tt.request)
			if tt.expected != nil {
				assert.Equal(t, tt.expected.Error(), err.Error())
			} else {
				assert.Equal(t, tt.expected, err)
			}
		})
	}
}

func TestGetStatus(t *testing.T) {
	tests := map[string]struct {
		clusterController func() cluster.Controller
		expectedVal       SparkApplicationState
		err               bool
	}{
		"failure | unable to get spark app": {
			clusterController: func() cluster.Controller {
				ctrler := &clustermock.Controller{}
				ctrler.On("GetSparkApplication", mock.Anything, mock.Anything, mock.Anything).Return(
					// Important to note that spark always returns a object rather than nil!!!
					&apisparkv1beta2.SparkApplication{},
					fmt.Errorf("hello"),
				)
				return ctrler
			},
			expectedVal: SparkApplicationStateUnknown,
			err:         true,
		},
		"success | completed job": {
			clusterController: func() cluster.Controller {
				ctrler := &clustermock.Controller{}
				ctrler.On("GetSparkApplication", mock.Anything, mock.Anything, mock.Anything).Return(
					&apisparkv1beta2.SparkApplication{
						Status: apisparkv1beta2.SparkApplicationStatus{
							AppState: apisparkv1beta2.ApplicationState{
								State: apisparkv1beta2.CompletedState,
							},
						},
					},
					nil,
				)
				return ctrler
			},
			expectedVal: SparkApplicationStateCompleted,
			err:         false,
		},
		"success | failed job": {
			clusterController: func() cluster.Controller {
				ctrler := &clustermock.Controller{}
				ctrler.On("GetSparkApplication", mock.Anything, mock.Anything, mock.Anything).Return(
					&apisparkv1beta2.SparkApplication{
						Status: apisparkv1beta2.SparkApplicationStatus{
							AppState: apisparkv1beta2.ApplicationState{
								State: apisparkv1beta2.FailedState,
							},
						},
					},
					nil,
				)
				return ctrler
			},
			expectedVal: SparkApplicationStateFailed,
			err:         false,
		},
		"success | unknown state": {
			clusterController: func() cluster.Controller {
				ctrler := &clustermock.Controller{}
				ctrler.On("GetSparkApplication", mock.Anything, mock.Anything, mock.Anything).Return(
					&apisparkv1beta2.SparkApplication{
						Status: apisparkv1beta2.SparkApplicationStatus{
							AppState: apisparkv1beta2.ApplicationState{
								State: apisparkv1beta2.UnknownState,
							},
						},
					},
					nil,
				)
				return ctrler
			},
			expectedVal: SparkApplicationStateUnknown,
			err:         false,
		},
		"success | other cases": {
			clusterController: func() cluster.Controller {
				ctrler := &clustermock.Controller{}
				ctrler.On("GetSparkApplication", mock.Anything, mock.Anything, mock.Anything).Return(
					&apisparkv1beta2.SparkApplication{
						Status: apisparkv1beta2.SparkApplicationStatus{
							AppState: apisparkv1beta2.ApplicationState{
								State: apisparkv1beta2.PendingRerunState,
							},
						},
					},
					nil,
				)
				return ctrler
			},
			expectedVal: SparkApplicationStateRunning,
			err:         false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			namespace := "test-ns"
			ensemblingJob := &models.EnsemblingJob{
				Name: "ensembling-job",
			}
			cc := tt.clusterController()
			var c EnsemblingController = &ensemblingController{
				clusterController: cc,
			}
			val, err := c.GetStatus(namespace, ensemblingJob)
			assert.Equal(t, tt.expectedVal, val)
			assert.True(t, (err != nil) == tt.err)
		})
	}
}

func TestDelete(t *testing.T) {
	tests := map[string]struct {
		clusterController func() cluster.Controller
		hasErr            bool
	}{
		"success | delete spark application": {
			clusterController: func() cluster.Controller {
				ctrler := &clustermock.Controller{}
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)

				ctrler.On("GetSparkApplication", mock.Anything, mock.Anything, mock.Anything).Return(
					&apisparkv1beta2.SparkApplication{},
					nil,
				)
				ctrler.On("DeleteSparkApplication", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
				)
				return ctrler
			},
			hasErr: false,
		},
		"success | no such job": {
			clusterController: func() cluster.Controller {
				ctrler := &clustermock.Controller{}
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything, mock.Anything, false).Return(nil)

				ctrler.On("GetSparkApplication", mock.Anything, mock.Anything, mock.Anything).Return(
					nil,
					fmt.Errorf("hello"),
				)
				return ctrler
			},
			// It is important that no action is to be done, we ignore the error
			// Just that there is no further action required on the delete function
			hasErr: false,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			clusterController := tt.clusterController()
			ensemblingController := NewBatchEnsemblingController(
				clusterController,
				nil,
				sparkInfraConfig,
			)
			err := ensemblingController.Delete("", &models.EnsemblingJob{})
			if tt.hasErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
