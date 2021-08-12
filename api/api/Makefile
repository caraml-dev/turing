SHELL := /bin/bash
VERSION_NUMBER=$(if $(VERSION),$(VERSION),$(shell ../../scripts/vertagen/vertagen.sh -f docker))

.PHONY: build-image
build-image:
	@docker build . \
		--tag $(if $(DOCKER_REGISTRY),$(DOCKER_REGISTRY)/,)turing/turing-swagger:${VERSION_NUMBER}

.PHONY: version
version:
	@echo $(VERSION_NUMBER)
