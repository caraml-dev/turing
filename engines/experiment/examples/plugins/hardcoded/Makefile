PLUGIN_NAME ?= plugin-example
VERSION ?= latest
GOOS ?= linux
GOARCH ?= amd64

.PHONY: tidy
tidy:
	@echo "Fetching dependencies..."
	go mod tidy

.PHONY: vendor
vendor:
	@echo "Fetching dependencies..."
	go mod vendor

.PHONY: build-image
build-image: vendor
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)${PLUGIN_NAME}-engine-plugin:${VERSION})
	@echo "Building docker image: ${IMAGE_TAG}"
	docker build --tag ${IMAGE_TAG} .

.PHONY: proprietary_exp_plugin
proprietary_exp_plugin:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./bin/example-plugin ./cmd/main.go
	mv ./bin/example-plugin ../../../../../api/bin/example-plugin
