module github.com/gojek/turing/engines/router

go 1.14

require (
	bou.ke/monkey v1.0.2
	cloud.google.com/go/bigquery v1.14.0
	github.com/confluentinc/confluent-kafka-go v1.4.2 // indirect
	github.com/fluent/fluent-logger-golang v1.5.0
	github.com/frankban/quicktest v1.8.1 // indirect
	github.com/go-playground/validator/v10 v10.3.0
	github.com/gojek/fiber v0.0.0-20201008181849-4f0f8284dc84
	github.com/gojek/mlp v1.4.7
	github.com/gojek/turing/engines/experiment v0.0.0
	github.com/google/go-cmp v0.5.5
	github.com/google/uuid v1.1.2
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/json-iterator/go v1.1.11
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pierrec/lz4 v2.4.1+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/stretchr/testify v1.7.0
	github.com/tinylib/msgp v1.1.2 // indirect
	github.com/uber/jaeger-client-go v2.23.1+incompatible
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	go.einride.tech/protobuf-bigquery v0.7.0
	go.uber.org/zap v1.19.1
	google.golang.org/protobuf v1.26.0-rc.1
	gopkg.in/confluentinc/confluent-kafka-go.v1 v1.4.2
)

replace github.com/gojek/turing/engines/experiment => ../experiment
