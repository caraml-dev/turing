package cluster

import (
	"fmt"
	"strconv"

	apisparkv1beta2 "github.com/GoogleCloudPlatform/spark-on-k8s-operator/pkg/apis/sparkoperator.k8s.io/v1beta2"
	apicorev1 "k8s.io/api/core/v1"
	apirbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	apimetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/caraml-dev/turing/api/turing/config"
	openapi "github.com/caraml-dev/turing/api/turing/generated"
)

const (
	// ServiceAccountFileName is the name of the service account
	ServiceAccountFileName = "service-account.json"

	serviceAccountMount      = "/mnt/secrets/"
	envServiceAccountPathKey = "GOOGLE_APPLICATION_CREDENTIALS"
	envServiceAccountPath    = serviceAccountMount + ServiceAccountFileName

	sparkType = "Python"
	sparkMode = "cluster"

	hadoopConfEnableServiceAccountKey = "google.cloud.auth.service.account.enable"
	hadoopConfEnableServiceAccount    = "true"
	haddopConfServiceAccountPathKey   = "google.cloud.auth.service.account.json.keyfile"
	haddopConfServiceAccountPath      = envServiceAccountPath
)

var (
	defaultEnvVars = []apicorev1.EnvVar{
		{
			Name:  envServiceAccountPathKey,
			Value: envServiceAccountPath,
		},
	}

	defaultHadoopConf = map[string]string{
		hadoopConfEnableServiceAccountKey: hadoopConfEnableServiceAccount,
		haddopConfServiceAccountPathKey:   haddopConfServiceAccountPath,
	}
)

// DefaultSparkDriverRoleRules is the default spark required policies
var DefaultSparkDriverRoleRules = []apirbacv1.PolicyRule{
	{
		// Allow driver to manage pods
		APIGroups: []string{
			"", // indicates the core API group
		},
		Resources: []string{
			"pods",
		},
		Verbs: []string{
			"*",
		},
	},
	{
		// Allow driver to manage services
		APIGroups: []string{
			"", // indicates the core API group
		},
		Resources: []string{
			"services",
		},
		Verbs: []string{
			"*",
		},
	},
}

// CreateSparkRequest is the request for creating a spark driver
type CreateSparkRequest struct {
	JobName               string
	JobLabels             map[string]string
	JobImageRef           string
	JobApplicationPath    string
	JobArguments          []string
	JobConfigMount        string
	DriverCPURequest      string
	DriverMemoryRequest   string
	ExecutorCPURequest    string
	ExecutorMemoryRequest string
	ExecutorReplica       int32
	ServiceAccountName    string
	SparkInfraConfig      *config.SparkAppConfig
	EnvVars               *[]openapi.EnvVar
}

func createSparkRequest(request *CreateSparkRequest) (*apisparkv1beta2.SparkApplication, error) {
	driver, err := createSparkDriver(request)
	if err != nil {
		return nil, err
	}
	executor, err := createSparkExecutor(request)
	if err != nil {
		return nil, err
	}
	spec := &apisparkv1beta2.SparkApplicationSpec{
		Type:                sparkType,
		SparkVersion:        request.SparkInfraConfig.SparkVersion,
		Mode:                sparkMode,
		Image:               &request.JobImageRef,
		MainApplicationFile: &request.JobApplicationPath,
		Arguments:           request.JobArguments,
		HadoopConf:          defaultHadoopConf,
		Driver:              *driver,
		Executor:            *executor,
		NodeSelector:        request.SparkInfraConfig.NodeSelector,
		RestartPolicy: apisparkv1beta2.RestartPolicy{
			Type:                             "OnFailure",
			OnSubmissionFailureRetries:       &request.SparkInfraConfig.SubmissionFailureRetries,
			OnFailureRetries:                 &request.SparkInfraConfig.FailureRetries,
			OnSubmissionFailureRetryInterval: &request.SparkInfraConfig.SubmissionFailureRetryInterval,
			OnFailureRetryInterval:           &request.SparkInfraConfig.FailureRetryInterval,
		},
		PythonVersion:     &request.SparkInfraConfig.PythonVersion,
		TimeToLiveSeconds: &request.SparkInfraConfig.TTLSecond,
	}

	return &apisparkv1beta2.SparkApplication{
		ObjectMeta: apimetav1.ObjectMeta{
			Name:   request.JobName,
			Labels: request.JobLabels,
		},
		Spec: *spec,
	}, nil
}

func getEnvVarFromRequest(request *CreateSparkRequest) []apicorev1.EnvVar {
	envVars := []apicorev1.EnvVar{}
	if request.EnvVars == nil {
		return envVars
	}

	for _, envVar := range *request.EnvVars {
		envVars = append(envVars, apicorev1.EnvVar{
			Name:  envVar.GetName(),
			Value: envVar.GetValue(),
		})
	}

	return envVars
}

func createSparkExecutor(request *CreateSparkRequest) (*apisparkv1beta2.ExecutorSpec, error) {
	userCPURequest, err := resource.ParseQuantity(request.ExecutorCPURequest)
	if err != nil {
		return nil, fmt.Errorf("invalid executor cpu request: %s", request.ExecutorCPURequest)
	}

	core := getCoreRequest(userCPURequest, request.SparkInfraConfig.CorePerCPURequest)
	cpuRequest, cpuLimit := getCPURequestAndLimit(userCPURequest, request.SparkInfraConfig.CPURequestToCPULimit)

	memoryRequest, err := toMegabyte(request.ExecutorMemoryRequest)
	if err != nil {
		return nil, fmt.Errorf("invalid executor memory request: %s", request.ExecutorMemoryRequest)
	}

	s := &apisparkv1beta2.ExecutorSpec{
		Instances:   &request.ExecutorReplica,
		CoreRequest: cpuRequest,
		SparkPodSpec: apisparkv1beta2.SparkPodSpec{
			Cores:     core,
			CoreLimit: cpuLimit,
			Memory:    memoryRequest,
			ConfigMaps: []apisparkv1beta2.NamePath{
				{
					Name: request.JobName,
					Path: request.JobConfigMount,
				},
			},
			Secrets: []apisparkv1beta2.SecretInfo{
				{
					Name: request.JobName,
					Path: serviceAccountMount,
				},
			},
			Env:    append(defaultEnvVars, getEnvVarFromRequest(request)...),
			Labels: request.JobLabels,
		},
	}
	if request.SparkInfraConfig.TolerationName != nil {
		s.SparkPodSpec.Tolerations = []apicorev1.Toleration{
			{
				Key:      *request.SparkInfraConfig.TolerationName,
				Operator: apicorev1.TolerationOpEqual,
				Value:    "true",
				Effect:   apicorev1.TaintEffectNoSchedule,
			},
		}
	}

	return s, nil
}

func createSparkDriver(request *CreateSparkRequest) (*apisparkv1beta2.DriverSpec, error) {
	userCPURequest, err := resource.ParseQuantity(request.DriverCPURequest)
	if err != nil {
		return nil, fmt.Errorf("invalid driver cpu request: %s", request.DriverCPURequest)
	}

	core := getCoreRequest(userCPURequest, request.SparkInfraConfig.CorePerCPURequest)
	cpuRequest, cpuLimit := getCPURequestAndLimit(userCPURequest, request.SparkInfraConfig.CPURequestToCPULimit)

	memoryRequest, err := toMegabyte(request.DriverMemoryRequest)
	if err != nil {
		return nil, fmt.Errorf("invalid driver memory request: %s", request.DriverMemoryRequest)
	}

	s := &apisparkv1beta2.DriverSpec{
		CoreRequest: cpuRequest,
		SparkPodSpec: apisparkv1beta2.SparkPodSpec{
			Cores:     core,
			CoreLimit: cpuLimit,
			Memory:    memoryRequest,
			ConfigMaps: []apisparkv1beta2.NamePath{
				{
					Name: request.JobName,
					Path: request.JobConfigMount,
				},
			},
			Secrets: []apisparkv1beta2.SecretInfo{
				{
					Name: request.JobName,
					Path: serviceAccountMount,
				},
			},
			Env:            append(defaultEnvVars, getEnvVarFromRequest(request)...),
			Labels:         request.JobLabels,
			ServiceAccount: &request.ServiceAccountName,
		},
	}
	if request.SparkInfraConfig.TolerationName != nil {
		s.SparkPodSpec.Tolerations = []apicorev1.Toleration{
			{
				Key:      *request.SparkInfraConfig.TolerationName,
				Operator: apicorev1.TolerationOpEqual,
				Value:    "true",
				Effect:   apicorev1.TaintEffectNoSchedule,
			},
		}
	}

	return s, nil
}

func getCoreRequest(cpuRequest resource.Quantity, corePerCPURequest float64) *int32 {
	var core int32
	core = int32(float64(cpuRequest.MilliValue()) / (corePerCPURequest * float64(1000)))
	if core < 1 {
		core = 1
	}
	return &core
}

func getCPURequestAndLimit(cpuRequest resource.Quantity, cpuRequestToCPULimit float64) (*string, *string) {
	cpuRequestStr := cpuRequest.String()

	cpuLimitMilli := cpuRequestToCPULimit * float64(cpuRequest.MilliValue())
	cpuLimit := resource.NewMilliQuantity(int64(cpuLimitMilli), resource.BinarySI)
	cpuLimitStr := cpuLimit.String()

	return &cpuRequestStr, &cpuLimitStr
}

func toMegabyte(request string) (*string, error) {
	req, err := resource.ParseQuantity(request)
	if err != nil {
		return nil, err
	}

	inBytes, ok := req.AsInt64()
	if !ok {
		return nil, fmt.Errorf("unable to convert to int64: %v", req)
	}

	inMegaBytes := inBytes / (1024 * 1024)
	strVal := fmt.Sprintf("%sm", strconv.Itoa(int(inMegaBytes)))
	return &strVal, nil
}
