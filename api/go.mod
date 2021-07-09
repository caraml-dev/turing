module github.com/gojek/turing/api

go 1.14

require (
	bou.ke/monkey v1.0.2
	github.com/DATA-DOG/go-sqlmock v1.3.3
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-20200424034326-7cd886c7ec44
	github.com/antihax/optional v1.0.0
	github.com/getkin/kin-openapi v0.20.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-playground/validator/v10 v10.3.0
	github.com/gojek/fiber v0.0.0-20201008181849-4f0f8284dc84
	github.com/gojek/merlin v0.0.0
	github.com/gojek/mlp v0.0.0
	github.com/gojek/turing/engines/batch-ensembler v0.0.0-00010101000000-000000000000
	github.com/gojek/turing/engines/experiment v0.0.0
	github.com/gojek/turing/engines/router v0.0.0
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/google/go-cmp v0.5.5
	github.com/google/go-containerregistry v0.0.0-20200331213917-3d03ed9b1ca2
	github.com/gorilla/mux v1.7.4
	github.com/gorilla/schema v1.1.0
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/jinzhu/gorm v1.9.12
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mitchellh/copystructure v1.0.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/ory/viper v1.7.5
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/tidwall/gjson v1.6.1
	github.com/xanzy/go-gitlab v0.31.0
	go.uber.org/zap v1.15.0
	golang.org/x/oauth2 v0.0.0-20201208152858-08078c50e5b5
	google.golang.org/protobuf v1.26.0
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v2.2.0+incompatible
	gotest.tools/v3 v3.0.3
	istio.io/api v0.0.0-20191115173247-e1a1952e5b81
	istio.io/client-go v0.0.0-20191120150049-26c62a04cdbc
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/pkg v0.0.0-20200519155757-14eb3ae3a5a7
	knative.dev/serving v0.15.0
	sigs.k8s.io/yaml v1.2.0
)

replace (
	// Ref: https://github.com/Azure/go-autorest/issues/414
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v38.2.0+incompatible
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.4.0+incompatible
	github.com/docker/docker => github.com/docker/engine v1.4.2-0.20200213202729-31a86c4ab209

	github.com/gojek/merlin => github.com/gojek/merlin/api v0.0.0-20200928080507-5e2221e40627
	github.com/gojek/merlin-pyspark-app => github.com/gojek/merlin/python/batch-predictor v0.0.0-20200928080507-5e2221e40627
	github.com/gojek/mlp => github.com/gojek/mlp/api v0.0.0-20200916102056-00ec9dccd758

	github.com/gojek/turing/engines/batch-ensembler => ../engines/batch-ensembler
	github.com/gojek/turing/engines/experiment => ../engines/experiment
	github.com/gojek/turing/engines/router => ../engines/router

	k8s.io/api => k8s.io/api v0.16.4

	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.16.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.4
	k8s.io/apiserver => k8s.io/apiserver v0.16.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.16.4
	k8s.io/client-go => k8s.io/client-go v0.16.4
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.16.6
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.16.6
	k8s.io/code-generator => k8s.io/code-generator v0.16.6
	k8s.io/component-base => k8s.io/component-base v0.16.6
	k8s.io/cri-api => k8s.io/cri-api v0.16.6
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.16.6
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.16.6
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.16.6
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.16.6
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.16.6
	k8s.io/kubectl => k8s.io/kubectl v0.16.4
	k8s.io/kubelet => k8s.io/kubelet v0.16.6
	k8s.io/kubernetes => k8s.io/kubernetes v1.16.4
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.16.6
	k8s.io/metrics => k8s.io/metrics v0.16.6
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.16.6
)
