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
		--install-types \
		--non-interactive \
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

.PHONY: build-image
build-image:
	mkdir -p temp-deps
	cp -r ../../sdk temp-deps/
	@docker build . \
		--tag $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)turing/batch-ensembler:${VERSION_NUMBER} \
		--tag $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)turing/batch-ensembler:latest
	rm -rf temp-deps

.PHONY: push-image
push-image:
	@docker push $(DOCKER_REGISTRY)/turing/batch-ensembler:$(VERSION_NUMBER)
	@docker push $(DOCKER_REGISTRY)/turing/batch-ensembler:latest

.PHONY: version
version:
	@echo $(VERSION_NUMBER)
