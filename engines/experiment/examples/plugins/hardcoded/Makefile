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

.PHONY: proprietary-exp-plugin
proprietary-exp-plugin:
	GOOS=$(GOOS) GOARCH=$(GOARCH) go build -o ./bin/example-plugin ./cmd/main.go
	ln -s ./bin/example-plugin ../../../../../api/bin/example-plugin

.PHONY: build-local-proprietary-exp-plugin-image
build-local-proprietary-exp-plugin-image:
	docker build --no-cache -t localhost:5000/proprietary-experiment-engine-plugin .
    docker push localhost:5000/proprietary-experiment-engine-plugin
