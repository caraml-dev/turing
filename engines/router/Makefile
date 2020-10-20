include .env.development
export

BIN_NAME=$(if $(APP_NAME),$(APP_NAME),turing-router)
VERSION_NUMBER=$(if $(VERSION),$(VERSION),$(shell ../../scripts/vertagen/vertagen.sh -f docker))
ALL_PACKAGES=$(shell go list ./...)
# TODO: Remove this step once Turing repo becomes public
GOPRIVATE=github.com/gojek/*
SRC_ROOT=./

USER=$(shell whoami)
HOST=$(shell hostname)
BRANCH=$(shell git rev-parse --abbrev-ref HEAD)
NOW=$(shell date +'%FT%T%z')
PROJECT_ROOT=github.com/gojek/turing/engines/router

.PHONY: default clean setup fmt dep-ci lint lint-ci tidy vendor test test-unit test-all build_image build build_docker run

default: test

clean:
	@test ! -e bin || rm -r bin
	@test ! -d vendor || rm -r vendor

setup:
	@test -x ${GOPATH}/bin/golint || go get -u golang.org/x/lint/golint

fmt:
	@echo "Fmt code..."
	gofmt -s -w ${SRC_ROOT}

dep-ci:
	@echo "Getting CI dependencies..."
	@test -x ${GOPATH}/bin/golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0

lint:
	@echo "Linting code..."
	golangci-lint -v run

lint-fix:
	@echo "Linting code..."
	golangci-lint -v run --fix

lint-ci: dep-ci lint

tidy: setup
	@echo "Fetching dependencies..."
	go mod tidy

vendor: setup
	@echo "Fetching dependencies..."
	go mod vendor

test: tidy
	@echo "Running tests..."
	go test -v -race -short -cover -coverprofile cover.out ${SRC_ROOT}/... -tags unit,integration
	go tool cover -func cover.out

test-unit: tidy
	@echo "Running tests..."
	go test -v -race -short -cover -coverprofile cover.out ${SRC_ROOT}/... -tags unit
	go tool cover -func cover.out

test-all: tidy
	@echo "Running all tests..."
	go test -v -race -cover -coverprofile cover.out ${SRC_ROOT}/... -tags unit,integration
	go test -bench=. -run=Bench ${SRC_ROOT}/...
	go tool cover -func cover.out

build: tidy fmt
	@echo "Building binary..."
	go build \
		-o ./bin/${BIN_NAME} \
		-ldflags " \
			-X ${PROJECT_ROOT}/missionctl/internal.Version=${VERSION_NUMBER} \
			-X ${PROJECT_ROOT}/missionctl/internal.Branch=${BRANCH} \
			-X ${PROJECT_ROOT}/missionctl/internal.BuildUser=${USER}@${HOST} \
			-X ${PROJECT_ROOT}/missionctl/internal.BuildDate=${NOW}" \
		${SRC_ROOT}/cmd/main.go

run: build
	./bin/${BIN_NAME}

build_image:
	echo "Building docker image: $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)${BIN_NAME}:${VERSION_NUMBER}"
	docker build \
		--tag $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)${BIN_NAME}:${VERSION_NUMBER} \
		--build-arg BIN_NAME \
		--build-arg VERSION=${VERSION_NUMBER} \
		--build-arg USER \
		--build-arg HOST \
		--build-arg BRANCH \
		--build-arg NOW \
		.

build_docker: vendor build_image clean

deploy_docker_stack:
	@echo "Using default Google Application Credentials for Auth..."
	@$(eval DEFAULT_CREDENTIALS=${HOME}/.config/gcloud/application_default_credentials.json)
	@$(eval GOOGLE_APPLICATION_CREDENTIALS=$(or $(GOOGLE_APPLICATION_CREDENTIALS),$(DEFAULT_CREDENTIALS)))
	@if test -f ${GOOGLE_APPLICATION_CREDENTIALS}; \
	then echo "... Found credentials in ${GOOGLE_APPLICATION_CREDENTIALS}"; \
		cat ${GOOGLE_APPLICATION_CREDENTIALS} | docker secret create google_app_creds -; \
	else echo "... No default credentials found, using a dummy value"; \
		echo dummy_cred | docker secret create google_app_creds -; \
	fi
	@mkdir -p ./fluentd_logs
	@cp .env.development compose/.env
	@IMAGE_VERSION=${VERSION_NUMBER} \
		LITMUS_PASSKEY=${LITMUS_PASSKEY} \
		XP_PASSKEY=${XP_PASSKEY} \
		docker stack deploy \
			-c compose/app.yaml \
			-c compose/docs.yaml \
			-c compose/fluentd.yaml \
			-c compose/kafka.yaml \
			-c compose/monitoring.yaml \
			-c compose/tracing.yaml \
			turing_stack
	@rm compose/.env

clean_docker_stack:
	@docker stack rm turing_stack
	@docker secret rm google_app_creds
	@echo "Removing fluentd logs..."
	@rm -rf ./fluentd_logs

swagger-ui:
	@docker-compose -f compose/docs.yaml up -d

jaeger-local:
	@docker-compose -f compose/tracing.yaml up -d

kafka-local:
	@docker-compose -f compose/kafka.yaml up -d

clean_services:
	@docker-compose \
		-f compose/app.yaml \
		-f compose/docs.yaml \
		-f compose/fluentd.yaml \
		-f compose/kafka.yaml \
		-f compose/monitoring.yaml \
		-f compose/tracing.yaml \
		down
