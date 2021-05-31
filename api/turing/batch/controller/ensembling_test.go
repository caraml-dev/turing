package batchcontroller

import (
	"fmt"
	"testing"

	"github.com/gojek/turing/api/turing/cluster"
	clustermock "github.com/gojek/turing/api/turing/cluster/mocks"
	"github.com/gojek/turing/api/turing/config"
	"github.com/gojek/turing/api/turing/models"
	"github.com/gojek/turing/api/turing/service"
	servicemock "github.com/gojek/turing/api/turing/service/mocks"
	batchensembler "github.com/gojek/turing/engines/batch-ensembler/pkg/api/proto/v1"
	"github.com/stretchr/testify/mock"
	"gotest.tools/assert"
	apicorev1 "k8s.io/api/core/v1"
	apirbacv1 "k8s.io/api/rbac/v1"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	imageRef  = "gojek/testimage:1"
	namespace = "test-namespace"
)

var (
	standardLabels map[string]string = map[string]string{
		"model": "T800",
	}
	sparkInfraConfig *config.SparkInfraConfig = &config.SparkInfraConfig{
		NodeSelector: map[string]string{
			"node-workload-type": "batch",
		},
		CorePerCPURequest:              1.5,
		CPURequestToCPULimit:           1.25,
		SparkVersion:                   "2.4.5",
		TolerationName:                 "batch-job",
		SubmissionFailureRetries:       3,
		SubmissionFailureRetryInterval: 10,
		FailureRetries:                 3,
		FailureRetryInterval:           10,
		PythonVersion:                  "3",
		TTLSecond:                      86400,
	}
)

func generateEnsemblingJobFixture() *models.EnsemblingJob {
	return &models.EnsemblingJob{
		Name:            "test-ensembler-1",
		ProjectID:       models.ID(1),
		EnsemblerID:     models.ID(1),
		EnvironmentName: "dev",
		InfraConfig: &models.InfraConfig{
			ArtifactURI:        "gs://bucket/ensembler",
			EnsemblerName:      "ensembler",
			ServiceAccountName: "test-service-account",
			Resources: &models.BatchEnsemblingJobResources{
				DriverCPURequest:      "1",
				DriverMemoryRequest:   "1Gi",
				ExecutorReplica:       10,
				ExecutorCPURequest:    "1",
				ExecutorMemoryRequest: "1Gi",
			},
		},
		JobConfig: &models.JobConfig{
			JobConfig: batchensembler.BatchEnsemblingJob{
				Version: "v1",
				Kind:    batchensembler.BatchEnsemblingJob_BatchEnsemblingJob,
				Metadata: &batchensembler.BatchEnsemblingJobMetadata{
					Name:        "test-batch-ensembling",
					Annotations: map[string]string{},
				},
				Spec: &batchensembler.BatchEnsemblingJobSpec{
					Source: &batchensembler.Source{
						Dataset: &batchensembler.Dataset{
							Type: batchensembler.Dataset_DatasetType(
								batchensembler.Dataset_BQ,
							),
							Config: &batchensembler.Dataset_BqConfig{
								BqConfig: &batchensembler.Dataset_BigQueryDatasetConfig{
									Query: "select * from helloworld where customer_id = 4",
									Options: map[string]string{
										"viewsEnabled":           "true",
										"materializationDataset": "dataset",
									},
								},
							},
						},
						JoinOn: []string{"customer_id", "target_date"},
					},
					Predictions: map[string]*batchensembler.PredictionSource{
						"model_a": {
							Dataset: &batchensembler.Dataset{
								Type: batchensembler.Dataset_DatasetType(
									batchensembler.Dataset_BQ,
								),
								Config: &batchensembler.Dataset_BqConfig{
									BqConfig: &batchensembler.Dataset_BigQueryDatasetConfig{
										Table: "project.dataset.predictions_model_a",
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
							Dataset: &batchensembler.Dataset{
								Type: batchensembler.Dataset_DatasetType(
									batchensembler.Dataset_BQ,
								),
								Config: &batchensembler.Dataset_BqConfig{
									BqConfig: &batchensembler.Dataset_BigQueryDatasetConfig{
										Query: "select * from helloworld where customer_id = 3",
									},
								},
							},
							Columns: []string{"predictions"},
							JoinOn:  []string{"customer_id", "target_date"},
						},
					},
					Ensembler: &batchensembler.Ensembler{
						Uri: "/home/spark/ensembler",
						Result: &batchensembler.Ensembler_Result{
							ColumnName: "prediction_score",
							Type:       batchensembler.Ensembler_FLOAT,
							ItemType:   batchensembler.Ensembler_FLOAT,
						},
					},
					Sink: &batchensembler.Sink{
						Type: batchensembler.Sink_BQ,
						Columns: []string{
							"customer_id as customerId",
							"target_date",
							"results",
						},
						SaveMode: batchensembler.Sink_OVERWRITE,
						Config: &batchensembler.Sink_BqConfig{
							BqConfig: &batchensembler.Sink_BigQuerySinkConfig{
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
		},
	}
}

func TestCreateEnsemblingJobController(t *testing.T) {
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
				ctrler.On("CreateNamespace", mock.Anything).Return(nil)
				ctrler.On("CreateServiceAccount", mock.Anything, mock.Anything).Return(
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
					mock.Anything,
				).Return(nil, nil)
				ctrler.On("CreateSecret", mock.Anything, mock.Anything).Return(nil, nil)
				ctrler.On("ApplyConfigMap", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("CreateSparkApplication", mock.Anything, mock.Anything).Return(nil, nil)
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
				ctrler.On("CreateNamespace", mock.Anything).Return(fmt.Errorf("hi"))
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything).Return(nil)
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
				ctrler.On("CreateNamespace", mock.Anything).Return(nil)
				ctrler.On("CreateServiceAccount", mock.Anything, mock.Anything).Return(
					nil,
					fmt.Errorf("hi"),
				)
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything).Return(nil)
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
				ctrler.On("CreateNamespace", mock.Anything).Return(nil)
				ctrler.On("CreateServiceAccount", mock.Anything, mock.Anything).Return(
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
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything).Return(nil)
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
				ctrler.On("CreateNamespace", mock.Anything).Return(nil)
				ctrler.On("CreateServiceAccount", mock.Anything, mock.Anything).Return(
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
					mock.Anything,
				).Return(nil, fmt.Errorf("hi"))
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything).Return(nil)
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
				ctrler.On("CreateNamespace", mock.Anything).Return(nil)
				ctrler.On("CreateServiceAccount", mock.Anything, mock.Anything).Return(
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
					mock.Anything,
				).Return(nil, nil)
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything).Return(nil)
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
				ctrler.On("CreateNamespace", mock.Anything).Return(nil)
				ctrler.On("CreateServiceAccount", mock.Anything, mock.Anything).Return(
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
					mock.Anything,
				).Return(nil, nil)
				ctrler.On("CreateSecret", mock.Anything, mock.Anything).Return(nil, nil)
				ctrler.On("ApplyConfigMap", mock.Anything, mock.Anything).Return(fmt.Errorf("hi"))
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything).Return(nil)
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
				ctrler.On("CreateNamespace", mock.Anything).Return(nil)
				ctrler.On("CreateServiceAccount", mock.Anything, mock.Anything).Return(
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
					mock.Anything,
				).Return(nil, nil)
				ctrler.On("CreateSecret", mock.Anything, mock.Anything).Return(nil, nil)
				ctrler.On("ApplyConfigMap", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("CreateSparkApplication", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("hi"))
				ctrler.On("DeleteSecret", mock.Anything, mock.Anything).Return(nil)
				ctrler.On("DeleteConfigMap", mock.Anything, mock.Anything).Return(nil)
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
