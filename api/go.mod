module github.com/caraml-dev/turing/api

go 1.18

require (
	bou.ke/monkey v1.0.2
	github.com/DATA-DOG/go-sqlmock v1.3.3
	github.com/GoogleCloudPlatform/spark-on-k8s-operator v0.0.0-20220113170521-22cd4a2c6990
	github.com/antihax/optional v1.0.0
	github.com/caraml-dev/turing/engines/experiment v0.0.0
	github.com/caraml-dev/turing/engines/router v0.0.0
	github.com/caraml-dev/universal-prediction-interface v0.3.4
	github.com/gavv/httpexpect/v2 v2.3.1
	github.com/getkin/kin-openapi v0.76.0
	github.com/ghodss/yaml v1.0.0
	github.com/go-playground/validator/v10 v10.11.1
	github.com/gojek/fiber v0.2.1-rc2
	github.com/gojek/merlin v0.0.0
	github.com/gojek/mlp v1.7.5-0.20230117024729-05ede139570e
	github.com/golang-collections/collections v0.0.0-20130729185459-604e922904d3
	github.com/golang-migrate/migrate/v4 v4.11.0
	github.com/google/go-cmp v0.5.9
	github.com/google/go-containerregistry v0.8.1-0.20220219142810-1571d7fdc46e
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.1.0
	github.com/heptiolabs/healthcheck v0.0.0-20180807145615-6ff867650f40
	github.com/mitchellh/copystructure v1.2.0
	github.com/mitchellh/mapstructure v1.5.0
	github.com/onsi/ginkgo/v2 v2.3.1
	github.com/onsi/gomega v1.22.1
	github.com/ory/viper v1.7.5
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/pkg/errors v0.9.1
	github.com/rs/cors v1.8.2
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.8.1
	github.com/xanzy/go-gitlab v0.32.0
	go.uber.org/zap v1.21.0
	golang.org/x/oauth2 v0.0.0-20221014153046-6fdb5e3db783
	google.golang.org/grpc v1.51.0
	google.golang.org/protobuf v1.28.1
	gopkg.in/yaml.v2 v2.4.0
	gopkg.in/yaml.v3 v3.0.1
	gorm.io/driver/postgres v1.4.5
	gorm.io/gorm v1.24.1-0.20221019064659-5dd2bb482755
	gotest.tools v2.2.0+incompatible
	gotest.tools/v3 v3.0.3
	istio.io/api v0.0.0-20220304035241-8c47cbbea144
	istio.io/client-go v1.12.5
	k8s.io/api v0.23.4
	k8s.io/apimachinery v0.26.0
	k8s.io/client-go v0.26.0
	knative.dev/pkg v0.0.0-20220222221138-929d328ad73c
	knative.dev/serving v0.27.2
	sigs.k8s.io/yaml v1.3.0
)

require (
	cloud.google.com/go/compute v1.14.0 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/ajg/form v1.5.1 // indirect
	github.com/andybalholm/brotli v1.0.3 // indirect
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/buger/jsonparser v1.1.1 // indirect
	github.com/certifi/gocertifi v0.0.0-20200922220541-2c3bb06c6054 // indirect
	github.com/cespare/xxhash v1.1.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgraph-io/ristretto v0.0.1 // indirect
	github.com/docker/cli v20.10.12+incompatible // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.17+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.4 // indirect
	github.com/evanphx/json-patch v4.11.0+incompatible // indirect
	github.com/evanphx/json-patch/v5 v5.5.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fatih/structs v1.1.0 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/getsentry/raven-go v0.2.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-openapi/analysis v0.19.5 // indirect
	github.com/go-openapi/errors v0.19.4 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/loads v0.19.4 // indirect
	github.com/go-openapi/runtime v0.19.15 // indirect
	github.com/go-openapi/spec v0.20.2 // indirect
	github.com/go-openapi/strfmt v0.19.5 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-openapi/validate v0.19.7 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator v9.31.0+incompatible // indirect
	github.com/go-stack/stack v1.8.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.11.2 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v0.16.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.4.3 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.6 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/yamux v0.0.0-20181012175058-2f1d1f20f75d // indirect
	github.com/imdario/mergo v0.3.13 // indirect
	github.com/imkira/go-interpol v1.0.0 // indirect
	github.com/jackc/chunkreader/v2 v2.0.1 // indirect
	github.com/jackc/pgconn v1.13.0 // indirect
	github.com/jackc/pgio v1.0.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgproto3/v2 v2.3.1 // indirect
	github.com/jackc/pgservicefile v0.0.0-20200714003250-2b9c44734f2b // indirect
	github.com/jackc/pgtype v1.12.0 // indirect
	github.com/jackc/pgx/v4 v4.17.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lib/pq v1.10.3 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.11 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2-0.20181231171920-c182affec369 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/mitchellh/hashstructure/v2 v2.0.2 // indirect
	github.com/mitchellh/reflectwalk v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/newrelic/go-agent v3.19.2+incompatible // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.3-0.20220114050600-8b9d41f48198 // indirect
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/ory/keto-client-go v0.4.4-alpha.1 // indirect
	github.com/pelletier/go-toml v1.9.5 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_golang v1.11.1 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.31.1 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/sergi/go-diff v1.1.0 // indirect
	github.com/sirupsen/logrus v1.9.0 // indirect
	github.com/spf13/afero v1.9.2 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	github.com/subosito/gotenv v1.2.0 // indirect
	github.com/tidwall/pretty v1.0.2 // indirect
	github.com/uber/jaeger-client-go v2.25.0+incompatible // indirect
	github.com/uber/jaeger-lib v2.4.0+incompatible // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.30.0 // indirect
	github.com/xeipuuv/gojsonpointer v0.0.0-20180127040702-4e3ac2762d5f // indirect
	github.com/xeipuuv/gojsonreference v0.0.0-20180127040603-bd5ef7bd5415 // indirect
	github.com/xeipuuv/gojsonschema v1.2.0 // indirect
	github.com/yalp/jsonpath v0.0.0-20180802001716-5cc68e5049a0 // indirect
	github.com/yudai/gojsondiff v1.0.0 // indirect
	github.com/yudai/golcs v0.0.0-20170316035057-ecda9a501e82 // indirect
	github.com/zaffka/zap-to-hclog v0.10.5 // indirect
	go.mongodb.org/mongo-driver v1.1.2 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.8.0 // indirect
	golang.org/x/crypto v0.5.0 // indirect
	golang.org/x/net v0.5.0 // indirect
	golang.org/x/sync v0.1.0 // indirect
	golang.org/x/sys v0.4.0 // indirect
	golang.org/x/term v0.4.0 // indirect
	golang.org/x/text v0.6.0 // indirect
	golang.org/x/time v0.0.0-20220210224613-90d013bbcef8 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20230104163317-caabf589fcbf // indirect
	gopkg.in/errgo.v2 v2.1.0 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/ini.v1 v1.63.2 // indirect
	istio.io/gogo-genproto v0.0.0-20210113155706-4daf5697332f // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
	k8s.io/kube-openapi v0.0.0-20211109043538-20434351676c // indirect
	k8s.io/utils v0.0.0-20221107191617-1a15be271d1d // indirect
	knative.dev/networking v0.0.0-20211101215640-8c71a2708e7d // indirect
	moul.io/http2curl v1.0.1-0.20190925090545-5cd742060b0e // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)

replace (
	cloud.google.com/go => cloud.google.com/go v0.104.0
	// Ref: https://github.com/Azure/go-autorest/issues/414
	github.com/Azure/azure-sdk-for-go => github.com/Azure/azure-sdk-for-go v38.2.0+incompatible
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v13.4.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.9.0

	github.com/caraml-dev/turing/engines/experiment => ../engines/experiment
	github.com/caraml-dev/turing/engines/router => ../engines/router

	// The older version of k8 lib uses 0.4, UPI indirect depencies uses 1.2 which is compatible
	github.com/go-logr/logr => github.com/go-logr/logr v0.4.0

	github.com/gojek/merlin => github.com/gojek/merlin/api v0.0.0-20210723093139-cc0240032d58
	github.com/gojek/merlin-pyspark-app => github.com/gojek/merlin/python/batch-predictor v0.0.0-20210723093139-cc0240032d58

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
	k8s.io/klog/v2 => k8s.io/klog/v2 v2.9.0
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
