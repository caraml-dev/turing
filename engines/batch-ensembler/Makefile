SHELL := /bin/bash

APP_NAME := batch-ensembler
CONDA_ENV_NAME ?= $(APP_NAME)
ACTIVATE_ENV = source $$(conda info --base)/etc/profile.d/conda.sh ; conda activate ./env/$(CONDA_ENV_NAME)

.PHONY: setup
setup: $(CONDA_ENV_NAME)
$(CONDA_ENV_NAME):
	@conda env update -p env/$(CONDA_ENV_NAME) -f environment.yaml --prune
	$(ACTIVATE_ENV) && pip install -r requirements.dev.txt

.PHONY: type-check
type-check:
	@$(ACTIVATE_ENV) && mypy \
		--install-types \
		--non-interactive \
		--ignore-missing-imports \
		--allow-untyped-globals \
		ensembler

.PHONY: test
test: type-check
	@$(ACTIVATE_ENV) && \
		python -m pytest \
		--cov=ensembler \
		--ignore=env \
		-W ignore

.PHONY: build-image
build-image: version
	@mkdir -p temp-deps
	@cp -r ../../sdk temp-deps/
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)${APP_NAME}:${VERSION})
	@echo "Building docker image: ${IMAGE_TAG}"
	@docker build . --tag ${IMAGE_TAG}
	@rm -rf temp-deps

.PHONY: version
version:
	$(eval VERSION=$(if $(OVERWRITE_VERSION),$(OVERWRITE_VERSION),v$(shell ../../scripts/vertagen/vertagen.sh -p ${APP_NAME}/)))
	@echo "turing-batch-ensembler version:" $(VERSION)
