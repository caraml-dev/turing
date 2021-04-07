SHELL := /bin/bash

CONDA_ENV_NAME ?= batch-ensembler
ACTIVATE_ENV = source $$(conda info --base)/etc/profile.d/conda.sh ; conda activate ./env/$(CONDA_ENV_NAME)

.PHONY: setup
setup: $(CONDA_ENV_NAME)
$(CONDA_ENV_NAME):
	@conda env update -p env/$(CONDA_ENV_NAME) -f environment.yaml --prune
	$(ACTIVATE_ENV) && pip install -r requirements.dev.txt

.PHONY: type-check
type-check: setup
	@$(ACTIVATE_ENV) && mypy \
		--ignore-missing-imports \
		--allow-untyped-globals \
		ensembler

.PHONY: test
test: type-check
	@$(ACTIVATE_ENV) && python -m pytest --cov=ensembler -W ignore --ignore=env