.PHONY: setup
setup:
	pip install -r requirements.txt -r requirements.dev.txt

.PHONY: gen-client
gen-client:
	rm -rf turing/generated
	@docker run --rm -v ${PWD}/..:/local openapitools/openapi-generator-cli:v5.1.1 generate \
		--input-spec /local/api/api/openapi-sdk.yaml \
		--generator-name python \
		--output /local/sdk/. \
		--config /local/sdk/openapi-codegen.yaml

.PHONY: version
version:
	$(eval VERSION=$(if $(OVERWRITE_VERSION),$(OVERWRITE_VERSION),$(shell ../scripts/vertagen/vertagen.sh -p sdk/ -y)))
	@echo 'VERSION = "$(VERSION)"' > turing/version.py
	@echo "turing-sdk version:" $(VERSION)

.PHONY: lint
lint:
	@$(eval black_version=$(shell pip show black | grep Version | grep -oE '[0-9].*'))
	@[ "${black_version}" == "22.6.0" ] || sh -c "pip install black==22.6.0"
	@black . --exclude=turing/generated/

.PHONY: build
build: version
	@rm -rf build dist
	@python setup.py sdist bdist_wheel

.PHONY: test
test:
	@unset http_proxy
	@unset https_proxy
	@python -m pytest --cov=turing tests/

e2e-sdk:
	@python -m pytest -s -v e2e
