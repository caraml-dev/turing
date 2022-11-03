module github.com/caraml-dev/turing/engines/router

go 1.18

require (
	bou.ke/monkey v1.0.2
	cloud.google.com/go/bigquery v1.14.0
	github.com/buger/jsonparser v1.1.1
	github.com/caraml-dev/turing/engines/experiment v0.0.0
	github.com/caraml-dev/universal-prediction-interface v0.0.0-20221026045401-50e7d79e4b73
	github.com/fluent/fluent-logger-golang v1.5.0
	github.com/go-playground/validator/v10 v10.3.0
	github.com/gojek/fiber v0.1.1-0.20221018054323-013517aeaf8f
	github.com/gojek/mlp v1.4.7
	github.com/google/go-cmp v0.5.8
	github.com/google/uuid v1.3.0
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/json-iterator/go v1.1.11
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/opentracing/opentracing-go v1.1.0
	github.com/pierrec/lz4 v2.4.1+incompatible
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/soheilhy/cmux v0.1.5
	github.com/stretchr/testify v1.8.0
	github.com/uber/jaeger-client-go v2.23.1+incompatible
	go.einride.tech/protobuf-bigquery v0.7.0
	go.uber.org/zap v1.21.0
	google.golang.org/grpc v1.50.1
	google.golang.org/protobuf v1.28.1
	gopkg.in/confluentinc/confluent-kafka-go.v1 v1.4.2
)

require (
	cloud.google.com/go v0.74.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/certifi/gocertifi v0.0.0-20191021191039-0944d244cd40 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/confluentinc/confluent-kafka-go v1.4.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/fatih/color v1.7.0 // indirect
	github.com/frankban/quicktest v1.8.1 // indirect
	github.com/getsentry/raven-go v0.2.0 // indirect
	github.com/ghodss/yaml v1.0.0 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator v9.31.0+incompatible // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/googleapis/gax-go/v2 v2.0.5 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.2 // indirect
	github.com/hashicorp/go-hclog v0.16.0 // indirect
	github.com/hashicorp/go-plugin v1.4.3 // indirect
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d // indirect
	github.com/jstemmer/go-junit-report v0.9.1 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/philhofer/fwd v1.0.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.26.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/stretchr/objx v0.4.0 // indirect
	github.com/tinylib/msgp v1.1.2 // indirect
	github.com/uber/jaeger-lib v2.2.0+incompatible // indirect
	github.com/zaffka/zap-to-hclog v0.10.5 // indirect
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/lint v0.0.0-20210508222113-6edffad5e616 // indirect
	golang.org/x/mod v0.6.0-dev.0.20220419223038-86c51ed26bb4 // indirect
	golang.org/x/net v0.0.0-20220809184613-07c6da5e1ced // indirect
	golang.org/x/oauth2 v0.0.0-20220722155238-128564f6959c // indirect
	golang.org/x/sys v0.0.0-20220728004956-3c1f35247d10 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/tools v0.1.12 // indirect
	google.golang.org/api v0.36.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220810155839-1856144b1d9c // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/caraml-dev/turing/engines/experiment => ../experiment

replace github.com/gojek/fiber => /Users/user/Documents/Code/github/fiber
