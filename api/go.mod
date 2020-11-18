module github.com/gojek/turing/api

go 1.14

require (
	bou.ke/monkey v1.0.2
	cloud.google.com/go/bigquery v1.9.0
	github.com/DATA-DOG/go-sqlmock v1.3.3
	github.com/getkin/kin-openapi v0.20.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-playground/validator/v10 v10.3.0
	github.com/gojek/fiber v0.0.0-20201008181849-4f0f8284dc84
	github.com/gojek/merlin v0.0.0
	github.com/gojek/mlp v0.0.0
	github.com/gojek/turing/engines/experiment v0.0.0
	github.com/gojek/turing/engines/router v0.0.0
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/google/go-cmp v0.5.0
	github.com/gorilla/mux v1.7.4
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/jinzhu/gorm v1.9.12
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/ory/viper v1.7.5
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/tidwall/gjson v1.6.1
	github.com/xanzy/go-gitlab v0.31.0
	go.uber.org/zap v1.15.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	google.golang.org/api v0.28.0
	gopkg.in/yaml.v2 v2.3.0
	gotest.tools v2.2.0+incompatible
	gotest.tools/v3 v3.0.2
	istio.io/api v0.0.0-20191115173247-e1a1952e5b81
	istio.io/client-go v0.0.0-20191120150049-26c62a04cdbc
	k8s.io/api v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/client-go v11.0.1-0.20190805182717-6502b5e7b1b5+incompatible
	knative.dev/pkg v0.0.0-20200519155757-14eb3ae3a5a7
	knative.dev/serving v0.15.0
)

replace (
	// Ref: https://github.com/Azure/go-autorest/issues/414
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v38.2.0+incompatible
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.4.0+incompatible

	github.com/docker/docker => github.com/docker/engine v1.4.2-0.20200213202729-31a86c4ab209

	github.com/gojek/merlin => github.com/gojek/merlin/api v0.0.0-20200928080507-5e2221e40627
	github.com/gojek/merlin-pyspark-app => github.com/gojek/merlin/python/batch-predictor v0.0.0-20200928080507-5e2221e40627
	github.com/gojek/mlp => github.com/gojek/mlp/api v0.0.0-20200916102056-00ec9dccd758

	github.com/gojek/turing/engines/experiment => ../engines/experiment
	github.com/gojek/turing/engines/router => ../engines/router

	k8s.io/api => k8s.io/api v0.16.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.4
	k8s.io/client-go => k8s.io/client-go v0.16.4
)
