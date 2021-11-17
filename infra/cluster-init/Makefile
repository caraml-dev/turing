SHELL := /bin/bash

APP_NAME := cluster-init

.PHONY: build-image
build-image: version
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)${APP_NAME}:${VERSION})
	@echo "Building docker image: ${IMAGE_TAG}"
	@docker build . --tag ${IMAGE_TAG}

.PHONY: version
version:
	$(eval VERSION=$(if $(OVERWRITE_VERSION),$(OVERWRITE_VERSION),v$(shell ../../scripts/vertagen/vertagen.sh -p ${APP_NAME}/)))
	@echo "turing-cluster-init version:" $(VERSION)
