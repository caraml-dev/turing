.PHONY: version
version:
	$(eval VERSION=$(if $(OVERWRITE_VERSION),$(OVERWRITE_VERSION),v$(shell ../vertagen/vertagen.sh -f docker)))
	@echo "version:" $(VERSION)


.PHONY: build-image
build-image: vendor version
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)fluentd:${VERSION})
	@echo "Building docker image: ${IMAGE_TAG}"
	docker build . --tag ${IMAGE_TAG}
