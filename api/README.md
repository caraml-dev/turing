# Turing API

API for the Turing experimentation service. 

## Getting Started

### Local Development

#### Requirements
- Golang 1.18
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

#### <b>Spin up k8s and required dependencies</b>
---

First, spin up k8s and other services required for Turing:

```bash
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
knative-serving   autoscaler-6f6f6cf579-czrmw               1/1     Running   0          17h
knative-serving   controller-66f68465cd-d58mq               1/1     Running   0          17h
knative-serving   domain-mapping-948df9f76-mwkhj            1/1     Running   0          17h
knative-serving   webhook-7877b457fd-vhcvm                  1/1     Running   0          17h
knative-serving   net-istio-webhook-695d588d65-r6szn        1/1     Running   0          17h
knative-serving   autoscaler-hpa-66d97b559f-ts6zf           1/1     Running   0          17h
knative-serving   net-istio-controller-544874485d-mncx9     1/1     Running   0          17h
default           mockserver-7844d797f7-zwk2b               1/1     Running   0          17h
kube-system       local-path-provisioner-84bb864455-58c7z   1/1     Running   0          17h
kube-system       coredns-96cc4f57d-t8vsw                   1/1     Running   0          17h
knative-serving   domainmapping-webhook-74fc9b87b4-pmqdg    1/1     Running   0          17h
istio-system      istiod-7c595445b6-r6fwt                   1/1     Running   0          17h
knative-serving   activator-7d658db58b-29pnb                1/1     Running   0          17h
kube-system       metrics-server-ff9dbcb6c-bjwl2            1/1     Running   0          17h
istio-system      svclb-istio-ingressgateway-kxtmp          3/3     Running   0          17h
istio-system      svclb-istio-ingressgateway-xszc7          3/3     Running   0          17h
istio-system      svclb-istio-ingressgateway-xg6zg          3/3     Running   0          17h
default           spark-spark-operator-59c685545-qqn44      1/1     Running   0          17h
istio-system      svclb-istio-ingressgateway-mn9n4          3/3     Running   0          17h
istio-system      istio-ingressgateway-b7ffbd9c6-kfpl2      1/1     Running   0          17h
```

#### <b>Prepare proprietary experiment engine</b>
---
Next, let's get the proprietary experiment engine plugin ready!

1. Build the proprietary experiment engine plugin binary

When building the binary, make sure you're using the same GOOS and GOARCH that's compatible with your machine.

```bash
cd engines/experiment/examples/plugins/hardcoded
# If using MacOS M1
make proprietary_exp_plugin GOOS=darwin GOARCH=arm64
```

2. Build the proprietary experiment engine and push it to the local registry.

```bash
pushd ../experiment/examples/plugins/hardcoded
{
    docker build -t localhost:5000/proprietary-experiment-engine-plugin .
    docker push localhost:5000/proprietary-experiment-engine-plugin
}
popd
```

If you're facing issues building the proprietary experiment engine docker image locally, do the following steps.

1. Replace "turing/engines/experiment/examples/plugins/hardcoded" Dockerfile with the following Dockerfile.
```docker
FROM alpine:3.15

ARG TURING_USER="turing"
ARG TURING_USER_GROUP="app"

RUN addgroup -S ${TURING_USER_GROUP} \
    && adduser -S ${TURING_USER} -G ${TURING_USER_GROUP} -H

ENV PLUGIN_NAME ""
ENV PLUGINS_DIR "/app/plugins"

COPY --chown=${TURING_USER}:${TURING_USER_GROUP} ./bin/example-plugin /go/bin/plugin

CMD ["sh", "-c", "cp /go/bin/plugin ${PLUGINS_DIR}/${PLUGIN_NAME:?variable must be set}"]
```

2. Build and push the image to the local registry.

```bash
docker build --no-cache -t localhost:5000/proprietary-experiment-engine-plugin
docker push localhost:5000/proprietary-experiment-engine-plugin
```

#### <b>Prepare Turing router</b>
---
Don't forget to build the Turing routers and push it to the local registry. Alternatively, use one of our images [here](https://github.com/caraml-dev/turing/pkgs/container/turing%2Fturing-router).

```bash
pushd ../engines/router
{
    go mod vendor
    docker build -t localhost:5000/turing-router .
    docker push localhost:5000/turing-router
}
popd
```

#### <b>Start Turing API</b>
---
Now fire up the turing API server (as a daemon or tmux or some other terminal process).

```
go run turing/cmd/main.go -config=config-dev.yaml -config=config-dev-exp-engine.yaml
```

#### <b>Run E2E tests</b>
---
```
ginkgo ./e2e/new-test/... -p -tags=e2e -run TestEndToEnd -- -config config-local.yaml
```

#### <b>Cleanup k8s and required dependencies</b>
---
```bash
pushd ../infra/docker-compose/dev/
{
    docker-compose down -v
}
popd
```

#### <b>Common local E2E setup issues</b>
---
If this doesn't work, check the following:

1. Have all the containers started properly, does `KUBECONFIG=/tmp/kubeconfig kubectl get pod -A` work? Check `docker ps -a` if containers have been deployed correctly.
  - Run `docker volume prune` if you're getting the following error
  ```text
  Unable to connect to the server: x509: certificate signed by unknown authority
  ```
  - If enricher/ensembler pod is constantly stuck at 1/2 ready state and `queue-proxy` container shows the following error, increase the `resource_request` values.
  ```text
  // Error logs from queue-proxy container
  HTTP probe did not respond Ready, got status code: 503
  HTTP probe did not respond Ready, got status code: 503
  ...

  // Increase to at least the following values
  "resource_request": {
    "min_replica": 1,
    "max_replica": 1,
    "cpu_request": "200m",
    "memory_request": "256Mi"
  }
  ```
2. Are you out of disk space? Not enough CPU/Memory? Check `KUBECONFIG=/tmp/kubeconfig kubectl get pod -A` and describe the pods to see if there is some sort of pressure if it's being evicted.

#### CI
The `turing-integration-test` MLP project is used by the CI to exercise the end to end tests. This project is pre-configured with the secret `ci_e2e_test_secret` that has the necessary access (JobUser, DataEditor) to the `gcp-project-id.dataset_id` dataset, as required by the tests for BQ logging with Fluentd. The CI step creates a `turing-api` deployment with an ingress and authorization disabled, to run tests. Additionally, the Fluentd flush interval is lowered for testing. A valid Google service account key must be set in the CI variables (`${SERVICE_ACCOUNT}`) for the test runner to access relevant resources (such as BigQuery results table).

#### Local
**Note:** Swarm mode must be enabled for this section.
1. Set the required env vars in `.env.development`, namely the `MLP_ECRYPTION_KEY` and `VAULT_TOKEN`
2. Set an MLP project id and the corresponding name in `e2e/local/.env.testing`. Set a unique `TEST_ID` (to prevent conflicts with CI runs / local runs triggered by other users). The env var `MODEL_CLUSTER_NAME` is used to indicate the cluster in which resources are expected to be created, and will be used to initialise the cluster client for validation. Thus, this cluster name must correspond to the `environment_name` property in the test data (under `e2e/test/testdata`) and has been configured so.
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
