package cluster

import (
	"testing"

	apisparkv1beta2 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	"github.com/stretchr/testify/assert"
	apicorev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caraml-dev/turing/api/turing/batch"
	"github.com/caraml-dev/turing/api/turing/config"
	openapi "github.com/caraml-dev/turing/api/turing/generated"
)

var (
	tolerationName                          = "batch-job"
	sparkInfraConfig *config.SparkAppConfig = &config.SparkAppConfig{
		NodeSelector: map[string]string{
			"node-workload-type": "batch",
		},
		CorePerCPURequest:              1.0,
		CPURequestToCPULimit:           1.0,
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

func TestGetCoreRequest(t *testing.T) {
	tests := map[string]struct {
		given             resource.Quantity
		corePerCPURequest float64
		expected          int32
	}{
		"nominal | <1": {
			resource.MustParse("200m"),
			1,
			1,
		},
		"nominal | >1, fractional": {
			resource.MustParse("1200m"),
			1,
			1,
		},
		"nominal | >1, integer": {
			resource.MustParse("2000m"),
			1,
			2,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result := getCoreRequest(tt.given, tt.corePerCPURequest)
			assert.Equal(t, tt.expected, *result)
		})
	}
}

func TestToMegabyte(t *testing.T) {
	tests := map[string]struct {
		given         string
		expectedValue string
		expectedErr   error
	}{
		"nominal | <1": {
			"200Mi",
			"200m",
			nil,
		},
		"failure | unable to parse": {
			"brains",
			"",
			resource.ErrFormatWrong,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			result, err := toMegabyte(tt.given)
			if tt.expectedValue != "" {
				assert.Equal(t, tt.expectedValue, *result)
			} else {
				assert.Nil(t, result)
			}

			if tt.expectedErr != nil {
				assert.Equal(t, tt.expectedErr.Error(), err.Error())
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetCPURequestAndLimit(t *testing.T) {
	tests := map[string]struct {
		given                 resource.Quantity
		cpuRequestToCPULimit  float64
		expectedCPURequestStr string
		expectedCPULimitStr   string
	}{
		"nominal | <1": {
			resource.MustParse("200m"),
			1,
			"200m",
			"200m",
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			cpuRequestStr, cpuLimitStr := getCPURequestAndLimit(
				tt.given,
				tt.cpuRequestToCPULimit,
			)
			assert.Equal(t, tt.expectedCPURequestStr, *cpuRequestStr)
			assert.Equal(t, tt.expectedCPULimitStr, *cpuLimitStr)
		})
	}
}

var (
	jobName                  = "jobname"
	jobImageRef              = "gojek/nosuchimage"
	jobApplicationPath       = "main.py"
	jobArguments             = []string{"--config", "yaml.engineer"}
	cpuValue                 = "200Mi"
	memoryValue              = "200M"
	executorReplica    int32 = 32
	serviceAccountName       = "service-account"
	jobLabels                = make(map[string]string)
	memoryResult, _          = toMegabyte(memoryValue)
	barString                = "bar"
	envVars                  = []openapi.EnvVar{
		{
			Name:  "foo",
			Value: &barString,
		},
	}
	expectedEnvVars = []apicorev1.EnvVar{
		{
			Name:  envServiceAccountPathKey,
			Value: envServiceAccountPath,
		},
		{
			Name:  "foo",
			Value: barString,
		},
	}
)

func TestCreateSparkRequest(t *testing.T) {
	given := &CreateSparkRequest{
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
		EnvVars:               &envVars,
	}
	expected := &apisparkv1beta2.SparkApplication{
		ObjectMeta: apimetav1.ObjectMeta{
			Name:   jobName,
			Labels: jobLabels,
		},
		Spec: apisparkv1beta2.SparkApplicationSpec{
			Type:                sparkType,
			SparkVersion:        sparkInfraConfig.SparkVersion,
			Mode:                sparkMode,
			Image:               &jobImageRef,
			MainApplicationFile: &jobApplicationPath,
			Arguments:           jobArguments,
			HadoopConf:          defaultHadoopConf,
			NodeSelector:        sparkInfraConfig.NodeSelector,
			RestartPolicy: apisparkv1beta2.RestartPolicy{
				Type:                             "OnFailure",
				OnSubmissionFailureRetries:       &sparkInfraConfig.SubmissionFailureRetries,
				OnFailureRetries:                 &sparkInfraConfig.FailureRetries,
				OnSubmissionFailureRetryInterval: &sparkInfraConfig.SubmissionFailureRetryInterval,
				OnFailureRetryInterval:           &sparkInfraConfig.FailureRetryInterval,
			},
			PythonVersion:     &sparkInfraConfig.PythonVersion,
			TimeToLiveSeconds: &sparkInfraConfig.TTLSecond,
			Executor: apisparkv1beta2.ExecutorSpec{
				Instances:   &executorReplica,
				CoreRequest: &cpuValue,
				SparkPodSpec: apisparkv1beta2.SparkPodSpec{
					Cores: getCoreRequest(
						resource.MustParse(cpuValue),
						sparkInfraConfig.CorePerCPURequest,
					),
					CoreLimit: &cpuValue,
					Memory:    memoryResult,
					ConfigMaps: []apisparkv1beta2.NamePath{
						{
							Name: jobName,
							Path: batch.JobConfigMount,
						},
					},
					Secrets: []apisparkv1beta2.SecretInfo{
						{
							Name: jobName,
							Path: serviceAccountMount,
						},
					},
					Env:    expectedEnvVars,
					Labels: jobLabels,
					Tolerations: []apicorev1.Toleration{
						{
							Key:      "batch-job",
							Operator: apicorev1.TolerationOpEqual,
							Value:    "true",
							Effect:   apicorev1.TaintEffectNoSchedule,
						},
					},
				},
			},
			Driver: apisparkv1beta2.DriverSpec{
				CoreRequest: &cpuValue,
				SparkPodSpec: apisparkv1beta2.SparkPodSpec{
					Cores: getCoreRequest(
						resource.MustParse(cpuValue),
						sparkInfraConfig.CorePerCPURequest,
					),
					CoreLimit: &cpuValue,
					Memory:    memoryResult,
					ConfigMaps: []apisparkv1beta2.NamePath{
						{
							Name: jobName,
							Path: batch.JobConfigMount,
						},
					},
					Secrets: []apisparkv1beta2.SecretInfo{
						{
							Name: jobName,
							Path: serviceAccountMount,
						},
					},
					Env:    expectedEnvVars,
					Labels: jobLabels,
					Tolerations: []apicorev1.Toleration{
						{
							Key:      "batch-job",
							Operator: apicorev1.TolerationOpEqual,
							Value:    "true",
							Effect:   apicorev1.TaintEffectNoSchedule,
						},
					},
					ServiceAccount: &serviceAccountName,
				},
			},
		},
	}
	t.Run("success | nominal", func(t *testing.T) {
		result, err := createSparkRequest(given)
		assert.Nil(t, err)
		assert.EqualValues(t, *expected, *result)
	})
}
