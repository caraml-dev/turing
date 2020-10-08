SRC_ROOT=.

.PHONY: default tidy dep-ci lint lint-ci

default: test

tidy:
	@echo "Fetching dependencies..."
	go mod tidy

fmt:
	@echo "Fmt code..."
	gofmt -s -w ${SRC_ROOT}

lint-install:
	@echo "Installing golangci-lint..."
	@test -x ${GOPATH}/bin/golangci-lint || go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0

lint: lint-install
	@echo "Linting code..."
	golangci-lint -v run

lint-fix:
	@echo "Linting code..."
	golangci-lint -v run --fix

lint-ci: dep-ci lint

test: tidy fmt
	@echo "Running tests..."
	go test -v -race -short -cover -coverprofile cover.out ${SRC_ROOT}/... -tags unit,integration
	go tool cover -func cover.out