# Turing API

API for the Turing experimentation service. 

## Getting Started

### Local Development

#### Requirements
- Golang 1.14
- Docker
  
#### Setup
To set up and install necessary tools, run:

```bash
make setup
```

### Re-generate openapi models 
OAS3 specs are used to generate required Golang structs, that define 
the configuration of ensembling batch jobs. If openapi specs have been 
changed, then corresponding code can be re-generated with:

```bash
make gen-client
```

### Explore REST API
```bash
make swagger-ui
```
This will open a Swagger-UI in your browser (accessible under [http://localhost:8081](http://localhost:8081))

### Run API locally
```bash
make run
```
This command will launch a local instance of postgresql db, apply latest db-migrations to it and run turing-api on default port (`8080`)

You can also run API in debug mode from your preferred IDE. In this case, you will need a db instance running locally:
```bash
make local-db
```

#### Running Auth Server locally
To test authorization locally:
1. Set env var `AUTHORIZATION_ENABLED=true`
2. You can modify the sample policy at `keto/policies/example_policy.json`
3. Run `make auth-server`. This should make the server accessible at `http://localhost:4466/`
4. Issue requests to the app with the header `User-Email: test-user@gojek.com`
5. Use `make clean-auth-server` to remove the local auth db and the keto server

### Run Tests
To run tests:
```bash
make test
```
Note that this will attempt to set up a postgres db using docker, to run integration tests.

### Run End-to-end Tests

Tests are placed in `e2e/tests/` directory, in the `e2e` package. The entrypoint is `all_e2e_test.go`, which runs a number of tests sequentially. At the end of all tests, in the tear down phase, the cluster resources created by the tests will be explicitly removed. Each test case and some helper methods are placed in separate files in the same package.

**Notes:**
1. As the `environment_name` is supposed to be set in the test data, these tests are currently only able to run in the `dev` environment. For running tests in `staging` / `production` environments, some changes will be required.
2. Test runs are isolated by using different MLP projects and/or TEST_ID (see the [Local](#local) section below). This will prevent conflicts in the cluster resources, if tests are run simultaneously. However, the BQ outcome table is also shared and thus, tear down will not remove it.

### Running End-to-end tests locally

This is a work in progress and there changes required on this process as this is extremely clunky. A rough way of running the e2e tests are as follows:

First, spin up k8s and other services required for Turing:

```bash
#!/bin/bash

pushd ../infra/docker-compose/dev/
{
    docker-compose up -d
}
popd
```

Here we have to wait for all the pods to be ready.

```bash
watch KUBECONFIG=/tmp/kubeconfig kubectl get pod -A
```

An example of ready would look something like this:
```
NAMESPACE         NAME                                      READY   STATUS    RESTARTS   AGE
kube-system       metrics-server-86cbb8457f-phdvp           1/1     Running   0          20m
knative-serving   controller-66b8964655-88fxq               1/1     Running   0          20m
knative-serving   istio-webhook-6cd54997b5-k4nsk            1/1     Running   0          20m
knative-serving   webhook-7d44db89cc-m2vdp                  1/1     Running   0          20m
kube-system       coredns-6488c6fcc6-chxtl                  1/1     Running   0          20m
knative-serving   networking-istio-c57bb746-pqtl2           1/1     Running   0          20m
default           mockserver-7844d797f7-8v6rr               1/1     Running   0          20m
knative-serving   autoscaler-6f6d898c75-6qx6k               1/1     Running   0          20m
kube-system       local-path-provisioner-5ff76fc89d-5m9fp   1/1     Running   0          20m
istio-system      istiod-7684b696d6-fjf4z                   1/1     Running   0          19m
knative-serving   activator-7fdbbcf6dc-jkcfn                1/1     Running   0          20m
default           spark-spark-operator-59c685545-5mrtg      1/1     Running   0          20m
istio-system      svclb-istio-ingressgateway-b9z82          4/4     Running   0          19m
istio-system      svclb-istio-ingressgateway-n9zzk          4/4     Running   0          19m
istio-system      svclb-istio-ingressgateway-77scg          4/4     Running   0          19m
istio-system      svclb-istio-ingressgateway-l26k7          4/4     Running   0          19m
istio-system      cluster-local-gateway-5bf54b4999-p2bv7    1/1     Running   0          19m
istio-system      istio-ingressgateway-555bdcd566-xcn6h     1/1     Running   0          19m
```

Now fire up the turing API server (as a daemon or tmux or some other terminal process).

```
# Run turing api server (either as a daemon or some other terminal):
go run turing/cmd/main.go -config=config-dev.yaml
```

Now run E2E tests.

```
TEST_ID=$(date +%Y%m%d%H%M) \
MOCKSERVER_ENDPOINT=http://mockserver \
API_BASE_PATH=http://localhost:8080/v1 \
MODEL_CLUSTER_NAME="dev" \
PROJECT_ID="1" \
PROJECT_NAME=default \
KUBECONFIG_USE_LOCAL=true \
KUBECONFIG_FILE_PATH=/tmp/kubeconfig \
go test -v -parallel=2 ./e2e/... -tags=e2e -run TestEndToEnd
```

To clean up the Kubernetes cluster:

```bash
#!/bin/bash

pushd ../infra/docker-compose/dev/
{
    docker-compose down -v
}
popd
```

#### CI
The `turing-integration-test` MLP project is used by the CI to exercise the end to end tests. This project is pre-configured with the secret `ci_e2e_test_secret` that has the necessary access (JobUser, DataEditor) to the `gcp-project-id.dataset_id` dataset, as required by the tests for BQ logging with Fluentd. The CI step creates a `turing-api` deployment with an ingress and authorization disabled, to run tests. Additionally, the Fluentd flush interval is lowered for testing. A valid Google service account key must be set in the CI variables (`${SERVICE_ACCOUNT}`) for the test runner to access relevant resources (such as BigQuery results table).

#### Local
**Note:** Swarm mode must be enabled for this section.
1. Set the required env vars in `.env.development`, namely the `MLP_ECRYPTION_KEY` and `VAULT_TOKEN`
2. Set an MLP project id and the corresponding name in `e2e/local/.env.testing`. Set a unique `TEST_ID` (to prevent conflicts with CI runs / local runs triggered by other users). Set the `TEST_LITMUS_PASSKEY` and `TEST_XP_PASSKEY` values to the passkey of the client id in the Litmus/XP integration environment, respectively. The env var `MODEL_CLUSTER_NAME` is used to indicate the cluster in which resources are expected to be created, and will be used to initialise the cluster client for validation. Thus, this cluster name must correspond to the `environment_name` property in the test data (under `e2e/test/testdata`) and has been configured so.
3. Run `make test-e2e-local`. This will deploy a local docker swarm set up for serving the Turing API and will run the end to end tests from the local environment, against it. The local services are destroyed after the tests.

## Folder structure of Turing API

| folder     | description 
|------------|-------------
| /turing/api | HTTP handlers and routing for Turing API as defined in [openapi.yaml](./api/openapi.yaml)
| /turing/cluster | Packages for creating, updating and deleting Turing router deployment in Kubernetes cluster
| /turing/cmd | Turing API server binaries
| /turing/config | Configuration spec for Turing API
| /turing/generated | Models generated from the openapi specs ([openapi-sdk.yaml](./api/openapi-sdk.yaml))
| /turing/middleware | HTTP server middlewares e.g. authorization and request validation
| /turing/models | Contains Turing domain model. Ideally, this package should not depend on packages outside the standard library.
| /turing/service | Packages responsible for persisting domain objects and managing their relationships. Communication between services should use dependency injection when possible to reduce unnecessary tight coupling.
| /turing/web | HTTP handler for serving the frontend
