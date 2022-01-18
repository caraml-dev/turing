ARG TURING_API_IMAGE
FROM bash:alpine3.14 as swagger-ui

WORKDIR /app

ARG openapi_spec_file="/app/swagger-ui/openapi.bundle.yaml"

# Build swagger ui deps
COPY ./scripts/swagger-ui-generator /app/swagger-ui-generator

RUN cd /app/swagger-ui-generator && ./swagger-ui-generator.sh \
    --spec-file ${openapi_spec_file} \
    --output /app/swagger-ui

FROM ${TURING_API_IMAGE}

ARG TURING_USER=${TURING_USER:-turing}
ARG TURING_USER_GROUP=${TURING_USER_GROUP:-app}

# Install bash
USER root
RUN apk add --no-cache bash
USER ${TURING_USER}

ARG openapi_spec_file=/app/swagger-ui/openapi.bundle.yaml
ARG turing_ui_dist_path=ui/build

ENV TURINGUICONFIG_SERVINGDIRECTORY "/app/turing-ui"

# Override the swagger ui config
ENV OPENAPICONFIG_SWAGGERUICONFIG_SERVINGDIRECTORY "/app/swagger-ui"
ENV OPENAPICONFIG_MERGEDSPECFILE ${openapi_spec_file}

COPY --chown=${TURING_USER}:${TURING_USER_GROUP} --from=swagger-ui /app/swagger-ui ${OPENAPICONFIG_SWAGGERUICONFIG_SERVINGDIRECTORY}/
COPY --chown=${TURING_USER}:${TURING_USER_GROUP} ${turing_ui_dist_path} ${TURINGUICONFIG_SERVINGDIRECTORY}/

COPY ./docker-entrypoint.sh ./

ENV TURING_UI_DIST_DIR ${TURINGUICONFIG_SERVINGDIRECTORY}

ENTRYPOINT ["./docker-entrypoint.sh"]
