SHELL := /bin/bash

APP_NAME := pyfunc-ensembler-service
CONDA_ENV_NAME ?= $(APP_NAME)
PYTHON_VERSION ?= 3.7
ACTIVATE_ENV = source $$(conda info --base)/etc/profile.d/conda.sh ; conda activate $(CONDA_ENV_NAME)

.PHONY: setup
setup: $(CONDA_ENV_NAME)
$(CONDA_ENV_NAME):
	@conda env update -f env-$(PYTHON_VERSION).yaml --prune
	$(ACTIVATE_ENV) && pip install -r requirements.dev.txt

.PHONY: test
test:
	@$(ACTIVATE_ENV) && \
		python -m pytest \
		--cov=pyfunc_ensembler_runner \
		--cov-report term-missing \
		-W ignore

.PHONY: build-image
build-image: version
	@mkdir -p temp-deps
	@cp -r ../../sdk temp-deps/
	@$(eval IMAGE_TAG = $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)${APP_NAME}:${VERSION})
	@echo "Building docker image: ${IMAGE_TAG}"
	@docker build . --tag ${IMAGE_TAG} \
		--build-arg APP_NAME=${APP_NAME} \
		--build-arg CONDA_ENV_NAME=${CONDA_ENV_NAME} \
		--build-arg PYTHON_VERSION=${PYTHON_VERSION}
	@rm -rf temp-deps

.PHONY: version
version:
	$(eval VERSION=$(if $(OVERWRITE_VERSION),$(OVERWRITE_VERSION),v$(shell ../../scripts/vertagen/vertagen.sh -p ${APP_NAME}/)))
	@echo "turing-pyfunc-ensembler-service version:" $(VERSION)
