include .env.development*
export

BIN_NAME=$(if $(APP_NAME),$(APP_NAME),turing-api)
VERSION_NUMBER=$(if $(VERSION),$(VERSION),$(shell ../scripts/vertagen/vertagen.sh -f docker))
ALL_PACKAGES=$(shell go list ./...)
GOPRIVATE=github.com/gojek/*
SRC_ROOT=./turing

E2E_TEST_ROOT=./e2e
E2E_API_PORT=8082
E2E_API_BASE_PATH=http://localhost:${E2E_API_PORT}/v1

.PHONY: default test clean setup tidy test build run local-db swagger-ui auth-server

default: test

clean:
	@test ! -e bin || rm -r bin

setup:
	@test -x ${GOPATH}/bin/golint || go get -u golang.org/x/lint/golint

fmt:
	@echo "Formatting code..."
	gofmt -s -w ${SRC_ROOT}

dep-ci:
	@echo "Getting CI dependencies..."
	@test -x ${GOPATH}/bin/golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0

lint:
	@echo "Linting code..."
	golangci-lint -v run --timeout 3m

lint-fix:
	@echo "Linting code..."
	golangci-lint -v run --timeout 3m --fix

lint-ci: dep-ci lint

tidy:
	@echo "Fetching dependencies..."
	go mod tidy

vendor: setup
	@echo "Fetching dependencies..."
	go mod vendor

test: tidy local-db
	@echo "Running tests..."
	go test -v -race -short -cover -coverprofile cover.out ${SRC_ROOT}/... -tags unit,integration
	go tool cover -func cover.out

test-ci: tidy
	echo "Running tests..."
	DATABASE_HOST=$${POSTGRES_HOST} \
	DATABASE_USER=$${POSTGRES_USER} \
	DATABASE_PASSWORD=$${POSTGRES_PASSWORD} \
	DATABASE_NAME=$${POSTGRES_DB} \
	go test -v -race -short -cover -coverprofile cover.out ${SRC_ROOT}/... -tags unit,integration
	go tool cover -func cover.out

test-e2e-local: deploy_docker_stack
	@$(eval include ./e2e/local/.env.testing)
	@echo "Running e2e tests..."
	go clean -testcache
	API_BASE_PATH=${E2E_API_BASE_PATH} \
		TEST_ID=$(if $(TEST_ID),$(TEST_ID),$(shell date +%Y%m%d%H%M)) \
		go test -v -parallel=2 ${E2E_TEST_ROOT}/... -tags e2e -timeout 15m -run TestEndToEnd || true
	make clean_docker_stack

build: tidy
	@echo "Building binary..."
	go build -o ./bin/${BIN_NAME} ${SRC_ROOT}/cmd/main.go

run: build local-db
	./bin/${BIN_NAME}

local-db:
	@docker-compose up -d postgres && docker-compose run migrations

swagger-ui:
	@docker-compose up -d swagger-ui
	@xdg-open 2>/dev/null http://localhost:8081 || open http://localhost:8081

auth-server:
	@docker-compose up -d postgres-auth && docker-compose run keto-server migrate sql -e
	@docker-compose up -d keto-server
	@docker-compose run keto-server-bootstrap-policies engines acp ory policies import glob /policies/example_policy.json

clean-auth-server:
	@docker stop keto-server && docker rm keto-server
	@docker stop postgres-auth && docker rm postgres-auth

build_image:
	echo "Building docker image: $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)${BIN_NAME}:${VERSION_NUMBER}"
	CGO_ENABLED=0 GOOS=linux go build -o ./bin/${BIN_NAME} ${SRC_ROOT}/cmd/main.go
	docker build \
		--tag $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)${BIN_NAME}:${VERSION_NUMBER} .

build_docker: vendor build_image clean

deploy_docker_stack: build_docker
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
	@IMAGE_VERSION=${VERSION_NUMBER} \
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
