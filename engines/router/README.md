# Turing Router

Turing enables DS teams to run ML experiments in the production environment.

## Local Development
Requirements:
* Golang 1.18

### Running the Turing Router locally
1. Provide configs for the turing router. Some sample configs are in `configs/`
2. Provide the environment variables in `.env.development`
3. Check that `librdkafka` is available locally (on Mac, this is usually under `/usr/local/opt`). If not, install it (on Mac, you can use homebrew: `brew install librdkafka`)
4. To run the app locally, do:
```
make build
make run
```
5. Issue a http request.
Sample test for `configs/default_router.yaml`:
```
curl -X POST \
  -H "Content-type: application/json" \
  -H "Accept: application/json" \
  -d '{}' \
  "http://localhost:8080/v1/predict"
```

### Running Tests
Run `make test` to run all unit tests and integration tests.

### API Docs
Run `make swagger-ui` to build the Swagger doc. This can then be accessed at <http://localhost:8081/>

### Running all Services

__Notes:__
1. Swarm mode must be enabled for this section.
2. If any of the services are unable to replicate, check the service logs. If there are no logs, make sure the docker image is available locally, and if not, issue a `docker pull ... `.

Compose files under `compose/` define various services for the Turing Router app, Fleuntd, Prometheus monitoring tool, Jaeger tracing backend, a Kafka broker / UI and the Swagger docs. This can be used as follows:

1. Set env var `GOOGLE_APPLICATION_CREDENTIALS` on the terminal if using BQ logging. For using Fluentd, the `GOOGLE_APPLICATION_CREDENTIALS` MUST belong to a service account.
2. Run `make build_docker` to build the Turing app's image. This is a time consuming step and may be skipped on subsequent runs if the Golang app does not need to be re-built for its code changes and there are no new commits in the branch since the last time the image was built.
3. Run `make deploy_docker_stack` to create the services. __Note:__ If not all services are required, simply remove the corresponding `-c <filename>` from the make target.
4. Access the Swagger Docs at <http://localhost:8081/>
5. Make a request to the Turing app at <http://localhost:8080/v1/predict>. If instrumentation is enabled (by `APP_CUSTOM_METRICS` in `.env.development`), the metrics exposed by the Prometheus client are available at <http://localhost:8080/metrics>.
6. Query the metrics from the Prometheus web UI at <http://localhost:9090/>
7. Test Fluentd `in_http` plugin by posting a request that matches, minimally, the required fields in the BigQuery output table schema. The logs would appear in the `./fluentd_logs` folder as they are being buffered. Eg: If `APP_FLUENTD_TAG` is set to `response.log`:
```
curl -v -X POST -d \
  'json={"turing_req_id": "ID1", "ts": "2020-05-11 03:29:41.367452 UTC", "request": {"header": "h1", "body": "b1"}}' \
  http://localhost:9880/response.log
```
__Note:__ The app uses the `in_forward` plugin to write data to a TCP socket.
8. The Jaeger tracing service runs the `all-in-one` image that includes the agent, collector and a UI. The stack set up is configured to trace all requests, for testing. The UI can be accessed at <http://localhost:16686/>
9. Test Kafka result logger using a local Kafka set up created by `make kafka-local`. You can use the Kafdrop UI exposed at <http://localhost:9001/> to explore the topics and messages. 
10. Clean up the deployment with `make clean_docker_stack`

## Dev Notes
1. Knative Serving maps port 8080 of the user container by default. This can be configured by setting `ports.containerPort` field. Refer to [PR](https://github.com/knative/serving/pull/2642).
2. Custom env vars can be added to each of the Turing app's user containers by setting the `env` key under the `enricher`, `ensembler` or `router` sections of the values file, as follows:
```
  env:
  - name: TEST_ENV1
    value: value1
  - name: TEST_ENV2
    value: value2
```
3. Jaeger client is initialised to trace all requests (`const` mode). However, the tracer's `IsEnabled()` methods determine whether the app adds the trace to the requests. To run Jaeger locally, do `make jaeger-local`. Simulating incoming requests with Spans:
```
  curl -v -X GET http://localhost:8080/v1/predict \
    -H "X-B3-Sampled: 1" \
    -H "X-B3-Spanid: a30ec88c39471716" \
    -H "X-B3-Traceid: 950f2de0b8430e9fa30ec88c39471716" \
    -d '{}'
```