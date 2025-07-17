SHELL := /bin/bash

APP_NAME := turing-pyfunc-ensembler-job
CONDA_ENV_NAME ?= $(APP_NAME)
ACTIVATE_ENV = source $$(conda info --base)/etc/profile.d/conda.sh ; conda activate $(CONDA_ENV_NAME)

.PHONY: setup
setup: build
	@conda env update -f env.yaml -n $(CONDA_ENV_NAME) --prune
	@DIST_VERSION=$$(echo $(VERSION) | \
		sed -E 's/^v([0-9]+\.[0-9]+\.[0-9]+)(-rc([0-9]+))?/\1rc\3/'); \
		$(ACTIVATE_ENV) && pip install "dist/turing_pyfunc_ensembler_job-$${DIST_VERSION}-py3-none-any.whl[dev]"

.PHONY: type-check
type-check:
	@$(ACTIVATE_ENV) && mypy \
		--install-types \
		--non-interactive \
		--ignore-missing-imports \
		--allow-untyped-globals \
		ensembler

.PHONY: lint
lint:
	@$(eval black_version=$(shell pip show black | grep Version | grep -oE '[0-9].*'))
	@[ "${black_version}" == "22.6.0" ] || sh -c "pip install black==22.6.0"
	@black .

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
	$(eval VERSION=$(if $(OVERWRITE_VERSION),$(OVERWRITE_VERSION),v$(shell ../../scripts/vertagen/vertagen.sh -p python/ -y)))
	@echo "turing-pyfunc-ensembler-job version:" $(VERSION)

.PHONY: build
build: version
	@rm -rf build dist
	@pip install "setuptools>=64,<75" "setuptools_scm>=8" "twine" "wheel"
	@sed -i -e "s|turing-sdk.*|turing-sdk==$(VERSION)|g" ./requirements.txt
	@sed -i -e "s|VERSION = \".*\"|VERSION = \"$(VERSION)\"|g" ./version.py
	@python setup.py sdist bdist_wheel

.PHONY: build-and-publish
build-and-publish: build
	@twine upload dist/* --verbose
