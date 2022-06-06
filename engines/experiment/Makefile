SRC_ROOT=.

.PHONY: default
default: test

.PHONY: setup
setup:
	@echo "Setting up tools..."
	@test -x $(shell go env GOPATH)/bin/golangci-lint || \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/v1.46.2/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.46.2

.PHONY: tidy
tidy:
	@echo "Fetching dependencies..."
	go mod tidy

.PHONY: fmt
fmt:
	@echo "Formatting code..."
	gofmt -s -w ${SRC_ROOT}

.PHONY: lint
lint: setup
	@echo "Linting code..."
	golangci-lint -v run $(if $(filter true,$(fix)),--fix,)

.PHONY: test
test: tidy
	@echo "Running tests..."
	go test -v -race -short -cover -coverprofile cover.out ${SRC_ROOT}/... -tags integration
	go tool cover -func cover.out