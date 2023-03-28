include .env.development*
export

BIN_NAME=$(if $(APP_NAME),$(APP_NAME),turing-api)
SRC_ROOT=./turing

default: test

.PHONY: clean
clean:
	@test ! -e bin || rm -r bin
	@test ! -e vendor || rm -rf vendor

.PHONY: setup
setup:
	@echo "Setting up tools..."
	@test -x ${GOPATH}/bin/golangci-lint || go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.2

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	gofmt -s -w ${SRC_ROOT}

.PHONY: lint
lint: setup
	@echo "Linting code..."
	golangci-lint -v run --timeout 3m $(if $(filter true,$(fix)),--fix,)

.PHONY: tidy
tidy:
	@echo "Fetching dependencies..."
	go mod tidy

.PHONY: vendor
vendor:
	@echo "Fetching dependencies..."
	go mod vendor

.PHONY: test
test: tidy
	@echo "Running tests..."
	go test -v -race -short -cover -coverprofile cover.out ${SRC_ROOT}/... -tags integration
	go tool cover -func cover.out

.PHONY: build
build: tidy
	@echo "Building binary..."
	go build -o ./bin/${BIN_NAME} ${SRC_ROOT}/cmd/main.go

.PHONY: build-image
build-image: vendor version
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)${BIN_NAME}:${VERSION})
	@echo "Building docker image: ${IMAGE_TAG}"
	docker build --tag ${IMAGE_TAG} .

.PHONY: run
run: build local-db
	./bin/${BIN_NAME} -config config-dev.yaml

.PHONY: local-db
local-db:
	@docker-compose up -d postgres && docker-compose run migrations

.PHONY: swagger-ui
swagger-ui:
	@docker-compose up -d swagger-ui
	@xdg-open 2>/dev/null http://localhost:8081 || open http://localhost:8081

.PHONY: gen-client
gen-client:
	rm -rf turing/generated
	@docker run --rm -v ${PWD}:/local openapitools/openapi-generator-cli:v5.1.1 generate \
		--input-spec /local/api/openapi-sdk.yaml \
		--generator-name go \
		--output /local/turing/generated \
		--config /local/openapi-codegen.yaml

.PHONY: version
version:
	$(eval VERSION=$(if $(OVERWRITE_VERSION),$(OVERWRITE_VERSION),v$(shell ../scripts/vertagen/vertagen.sh)))
	@echo "turing-api version:" $(VERSION)

# ==================================
# E2E tests
# ==================================

.PHONY: build-run-local-api
build-run-local-api:
	@echo "Building proprietary experiment engine plugin..."
	cd ../engines/experiment/examples/plugins/hardcoded && make proprietary-exp-plugin GOOS=darwin
	@echo "Building proprietary experiment engine image..."
	cd ../engines/experiment/examples/plugins/hardcoded && make build-local-proprietary-exp-plugin-image
	@echo "Building router image..."
	cd ../engines/router && make build-local-router-image DOCKER_REGISTRY=localhost:5000 OVERWRITE_VERSION=latest
	@echo "Generating cluster cred"
	sh ../infra/docker-compose/dev/extract_creds.sh
	@echo "Starting Turing API server in a background process..."
	nohup go run turing/cmd/main.go -config=config-dev-w-creds.yaml -config=config-dev-exp-engine.yaml &
	@echo "Started Turing API server on port 8080 in a background process..."

.PHONY: clean-local-infra
clean-local-infra:
	@echo "Stopping background process Turing API server..."
	lsof -ti:8080 | xargs kill -9
	@echo "Cleaning up all E2E tests dependencies..."
	cd ../infra/docker-compose/dev && docker-compose down -v

.PHONY: run-local-e2e
run-local-e2e:
	ginkgo ./e2e/test/... -p -tags=e2e -run TestEndToEnd -- -config config-local.yaml

.PHONY: setup-e2e
setup-e2e:
	go get github.com/onsi/ginkgo/v2/ginkgo/generators@v2.3.1
	go get github.com/onsi/ginkgo/v2/ginkgo/internal@v2.3.1
	go get github.com/onsi/gomega@v1.22.1
	go install github.com/onsi/ginkgo/v2/ginkgo@v2.3.1

.PHONY: test-e2e
test-e2e:
	@$(shell go env GOPATH)/bin/ginkgo ./e2e/test/... -p -run TestEndToEnd

#   R E V I E W   R E Q U I R E D
#
#   Following Makefile targets aren't expected to be working
#   correctly and require further review and refactoring
#

E2E_TEST_ROOT=./e2e
E2E_API_PORT=8082
E2E_API_BASE_PATH=http://localhost:${E2E_API_PORT}/v1

test-e2e-local: deploy_docker_stack
	@$(eval include ./e2e/local/.env.testing)
	@echo "Running e2e tests..."
	go clean -testcache
	API_BASE_PATH=${E2E_API_BASE_PATH} \
		TEST_ID=$(if $(TEST_ID),$(TEST_ID),$(shell date +%Y%m%d%H%M)) \
		go test -v -parallel=2 ${E2E_TEST_ROOT}/... -tags e2e -timeout 15m -run TestEndToEnd || true
	make clean_docker_stack

auth-server:
	@docker-compose up -d postgres-auth && docker-compose run keto-server migrate sql -e
	@docker-compose up -d keto-server
	@docker-compose run keto-server-bootstrap-policies engines acp ory policies import glob /policies/example_policy.json

clean-auth-server:
	@docker stop keto-server && docker rm keto-server
	@docker stop postgres-auth && docker rm postgres-auth

deploy_docker_stack: build-docker version
	@echo "Using default Google Application Credentials for Auth..."
	@$(eval DEFAULT_CREDENTIALS=${HOME}/.config/gcloud/application_default_credentials.json)
	@$(eval GOOGLE_APPLICATION_CREDENTIALS=$(or $(GOOGLE_APPLICATION_CREDENTIALS),$(DEFAULT_CREDENTIALS)))
	@if test -f ${GOOGLE_APPLICATION_CREDENTIALS}; \
	then echo "... Found credentials in ${GOOGLE_APPLICATION_CREDENTIALS}"; \
		cat ${GOOGLE_APPLICATION_CREDENTIALS} | docker secret create google_app_creds -; \
	else echo "... No default credentials found, exiting."; \
		exit 1; \
	fi
	@echo "Deploying local docker stack..."
	@cp .env.development ./e2e/local/.env
	@IMAGE_VERSION=${VERSION} \
		API_PORT=${E2E_API_PORT} \
		docker stack deploy --compose-file=./e2e/local/docker-compose.yaml turing_api_stack
	@rm ./e2e/local/.env
	@echo "Waiting for the API service to become ready..."
	@$(shell chmod +x ./e2e/local/readiness_probe.sh)
	@./e2e/local/readiness_probe.sh ${E2E_API_BASE_PATH}

clean_docker_stack:
	@echo "Cleaning up local docker stack..."
	@docker stack rm turing_api_stack
	@docker secret rm google_app_creds
