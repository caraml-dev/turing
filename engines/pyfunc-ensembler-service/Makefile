SHELL := /bin/bash

PYTHON_VERSION ?= 3.8
APP_NAME := pyfunc-ensembler-service-py$(PYTHON_VERSION)
CONDA_ENV_NAME ?= $(APP_NAME)
ACTIVATE_ENV = source $$(conda info --base)/etc/profile.d/conda.sh ; conda activate $(CONDA_ENV_NAME)

.PHONY: setup
setup: $(CONDA_ENV_NAME)
$(CONDA_ENV_NAME):
	@conda env update -f env-$(PYTHON_VERSION).yaml -n $(CONDA_ENV_NAME) --prune
	$(ACTIVATE_ENV) && pip install -r requirements.dev.txt

.PHONY: lint
lint:
	@$(eval black_version=$(shell pip show black | grep Version | grep -oE '[0-9].*'))
	@[ "${black_version}" == "22.6.0" ] || sh -c "pip install black==22.6.0"
	@black .

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

.PHONY: build
build: version
	@rm -rf build dist
	@pip install "setuptools>=64,<75" "setuptools_scm>=8" "twine" "wheel"
	@tag=$$(python -m setuptools_scm -r ../.. | sed 's/\+.*//'); \
	sed -i -e "s|turing-sdk.*|turing-sdk==$$tag|g" ./requirements.txt; \
	sed -i -e "s|VERSION = \".*\"|VERSION = \"$$tag\"|g" ./version.py
	@python setup.py sdist bdist_wheel
	@TWINE_USERNAME=$(PYPI_USERNAME) TWINE_PASSWORD=$(PYPI_PASSWORD) twine upload dist/*
