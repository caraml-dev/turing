# Build turing-api binary
FROM golang:1.18-alpine as api-builder
ARG API_BIN_NAME=turing-api

ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64

ENV PROJECT_ROOT=github.com/gojek/turing/api/turing

WORKDIR /app
COPY . .

# Build Turing binary
RUN go build \
    -mod=vendor \
    -o ./bin/${API_BIN_NAME} \
    -v ${PROJECT_ROOT}/cmd

# Clean image with turing-api binary
FROM alpine:3.13

ENV TURING_PORT "8080"
ENV TURING_USER "turing"
ENV TURING_USER_GROUP "app"

EXPOSE ${TURING_PORT}

RUN addgroup -S ${TURING_USER_GROUP} \
    && adduser -S ${TURING_USER} -G ${TURING_USER_GROUP} -H \
    && mkdir /app \
    && chown -R ${TURING_USER}:${TURING_USER_GROUP} /app

COPY --chown=${TURING_USER}:${TURING_USER_GROUP} --from=api-builder /app/bin/* /app
COPY --chown=${TURING_USER}:${TURING_USER_GROUP} --from=api-builder /app/db-migrations /app/db-migrations
COPY --chown=${TURING_USER}:${TURING_USER_GROUP} --from=api-builder /app/api/openapi.bundle.yaml /app/api/openapi.bundle.yaml

USER ${TURING_USER}
WORKDIR /app

ARG API_BIN_NAME=turing-api
ENV TURING_API_BIN "./${API_BIN_NAME}"

ENTRYPOINT ["sh", "-c", "${TURING_API_BIN} \"$@\"", "--"]
