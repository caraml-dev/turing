.PHONY: bundle-openapi
bundle-openapi:
	@docker run --rm \
		--volume ${PWD}:/local \
		 openapitools/openapi-generator-cli:v5.1.1 generate \
        	--input-spec /local/openapi.yaml \
        	--generator-name openapi-yaml \
        	--additional-properties outputFile=openapi.bundle.yaml \
        	--output /local
	@for f in ".openapi-generator" ".openapi-generator-ignore" "README.md" ; do \
		rm -rf ${PWD}/api/$${f} ;\
	done

.PHONY: swagger-ui-dist
swagger-ui-dist: bundle-openapi
	@scripts/swagger-ui-generator/swagger-ui-cli.sh \
		--spec-file api/openapi.bundle.yaml \
		--output static/swagger-ui