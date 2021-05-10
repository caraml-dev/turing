.PHONY: setup

env:
	pip install -r requirements.txt

dev: env
	pip install -r requirements.dev.txt

gen-client:
	rm -rf turing/generated
	@openapi-generator generate \
		--input-spec ../api/openapi-sdk.yaml \
		--generator-name python \
		--output . \
		--config openapi-codegen.yaml

build:
	python setup.py check
	python setup.py sdist