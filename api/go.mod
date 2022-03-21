module github.com/gojek/turing/api

go 1.14

require (
	bou.ke/monkey v1.0.2
	github.com/DATA-DOG/go-sqlmock v1.3.3
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-20220113170521-22cd4a2c6990
	github.com/antihax/optional v1.0.0
	github.com/getkin/kin-openapi v0.76.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-playground/validator/v10 v10.9.0
	github.com/gojek/fiber v0.0.0-20201008181849-4f0f8284dc84
	github.com/gojek/merlin v0.0.0
	github.com/gojek/mlp v1.4.7
	github.com/gojek/turing/engines/experiment v0.0.0
	github.com/gojek/turing/engines/router v0.0.0
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/google/go-cmp v0.5.7
	github.com/google/go-containerregistry v0.8.1-0.20220219142810-1571d7fdc46e
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.1.0
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/jinzhu/gorm v1.9.12
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mitchellh/copystructure v1.0.0
	github.com/mitchellh/mapstructure v1.4.3
	github.com/ory/viper v1.7.5
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.6.5
	github.com/xanzy/go-gitlab v0.31.0
	go.uber.org/zap v1.19.1
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
	gotest.tools v2.2.0+incompatible
	gotest.tools/v3 v3.0.3
	istio.io/api v0.0.0-20220304035241-8c47cbbea144
	istio.io/client-go v1.12.5
	k8s.io/api v0.22.7
	k8s.io/apimachinery v0.22.7
	k8s.io/client-go v0.22.7
	knative.dev/pkg v0.0.0-20220222221138-929d328ad73c
	knative.dev/serving v0.27.2
	sigs.k8s.io/yaml v1.3.0
)

replace (
	// Ref: https://github.com/Azure/go-autorest/issues/414
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v38.2.0+incompatible
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.4.0+incompatible
	github.com/docker/docker => github.com/docker/engine v1.4.2-0.20200213202729-31a86c4ab209

	github.com/gojek/merlin => github.com/gojek/merlin/api v0.0.0-20210723093139-cc0240032d58
	github.com/gojek/merlin-pyspark-app => github.com/gojek/merlin/python/batch-predictor v0.0.0-20210723093139-cc0240032d58

	github.com/gojek/turing/engines/experiment => ../engines/experiment
	github.com/gojek/turing/engines/router => ../engines/router

	k8s.io/api => k8s.io/api v0.22.7

	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.7
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.7
	k8s.io/apiserver => k8s.io/apiserver v0.22.7
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.22.7
	k8s.io/client-go => k8s.io/client-go v0.22.7
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.22.7
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.22.7
	k8s.io/code-generator => k8s.io/code-generator v0.22.7
	k8s.io/component-base => k8s.io/component-base v0.22.7
	k8s.io/component-helpers => k8s.io/component-helpers v0.22.7
	k8s.io/controller-manager => k8s.io/controller-manager v0.22.7
	k8s.io/cri-api => k8s.io/cri-api v0.22.7
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.22.7
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.22.7
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.22.7
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.22.7
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.22.7
	k8s.io/kubectl => k8s.io/kubectl v0.22.7
	k8s.io/kubelet => k8s.io/kubelet v0.22.7
	k8s.io/kubernetes => k8s.io/kubernetes v1.22.7
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.22.7
	k8s.io/metrics => k8s.io/metrics v0.22.7
	k8s.io/mount-utils => k8s.io/mount-utils v0.22.7
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.22.7
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.22.7
)
