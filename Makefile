APP_NAME := $(or ${APP_NAME},${APP_NAME},turing)

.PHONY: build-ui
build-ui:
	@echo "Creating an optimized production build of Turing UI..."
	@cd ui && \
	yarn install --network-concurrency 1 && \
 	NODE_OPTIONS=--max_old_space_size=4096 yarn build

.PHONY: build-swagger-ui
build-swagger-ui:
	$(MAKE) -C api/api swagger-ui-dist

.PHONY: build-image
build-image: version
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)${APP_NAME}:${VERSION})
	@echo "Building docker image: ${IMAGE_TAG}"
	docker build . \
		--tag ${IMAGE_TAG} \
		--build-arg TURING_API_IMAGE \
		--build-arg TURING_UI_DIST_PATH

.PHONY: version
version:
	$(eval VERSION=$(if $(OVERWRITE_VERSION),$(OVERWRITE_VERSION),v$(shell scripts/vertagen/vertagen.sh)))
	@echo "turing version:" $$VERSION
