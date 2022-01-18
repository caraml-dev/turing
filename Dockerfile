ARG TURING_API_IMAGE
FROM ${TURING_API_IMAGE}

ARG TURING_USER=${TURING_USER:-turing}
ARG TURING_USER_GROUP=${TURING_USER_GROUP:-app}

ENV TURINGUICONFIG_SERVINGDIRECTORY "/app/turing-ui"

USER root
RUN apk add --no-cache bash

# Override the swagger ui config
ENV OPENAPICONFIG_SWAGGERUICONFIG_SERVINGDIRECTORY "/app/swagger-ui"
ENV OPENAPICONFIG_MERGEDSPECFILE "/app/swagger-ui/openapi.bundle.yaml"

# Build swagger ui deps
COPY ./scripts/swagger-ui-generator /app/swagger-ui-generator
RUN cd /app/swagger-ui-generator && ./swagger-ui-generator.sh \
    --spec-file ${OPENAPICONFIG_MERGEDSPECFILE} \
    --output ${OPENAPICONFIG_SWAGGERUICONFIG_SERVINGDIRECTORY}
RUN rm -rf /app/swagger-ui-generator

# Switch back to turing user
USER ${TURING_USER}
ARG TURING_UI_DIST_PATH=ui/build

COPY --chown=${TURING_USER}:${TURING_USER_GROUP} ${TURING_UI_DIST_PATH} ${TURINGUICONFIG_SERVINGDIRECTORY}/

COPY ./docker-entrypoint.sh ./

ENV TURING_UI_DIST_DIR ${TURINGUICONFIG_SERVINGDIRECTORY}


ENTRYPOINT ["./docker-entrypoint.sh"]
