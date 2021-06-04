SHELL := /bin/bash

CONDA_ENV_NAME ?= batch-ensembler
VERSION_NUMBER=$(if $(VERSION),$(VERSION),$(shell ../../scripts/vertagen/vertagen.sh -f docker))
ACTIVATE_ENV = source $$(conda info --base)/etc/profile.d/conda.sh ; conda activate ./env/$(CONDA_ENV_NAME)

.PHONY: setup
setup: $(CONDA_ENV_NAME)
$(CONDA_ENV_NAME):
	@conda env update -p env/$(CONDA_ENV_NAME) -f environment.yaml --prune
	$(ACTIVATE_ENV) && pip install -r requirements.dev.txt

.PHONY: type-check
type-check:
	@$(ACTIVATE_ENV) && mypy \
		--ignore-missing-imports \
		--allow-untyped-globals \
		ensembler

.PHONY: test-unit
test-unit: type-check
	@$(ACTIVATE_ENV) && \
		python -m pytest \
		--cov=ensembler \
		--ignore=env \
		-W ignore

.PHONY: proto
proto:
	protoc -I=. \
		--python_out=ensembler \
		--mypy_out=ensembler \
		--go_out=pkg \
		--go_opt=paths=source_relative \
		api/proto/v1/*.proto

.PHONY: build-image
build-image:
	@./build-image.sh \ 
		$(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)turing/batch-ensembler:${VERSION_NUMBER}

.PHONY: version
version:
	@echo $(VERSION_NUMBER)
