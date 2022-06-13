# Build application binary
FROM golang:1.18-alpine as builder
ARG BIN_NAME=turing-router
ARG VERSION
ARG USER
ARG HOST
ARG BRANCH
ARG NOW

ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64

ENV PROJECT_ROOT=github.com/gojek/turing/engines/router/missionctl

WORKDIR /app
COPY . .

# Install gcc
RUN apk add build-base librdkafka

RUN go build \
    -mod=vendor \
    -tags musl \
    -o ./bin/${BIN_NAME} \
    -ldflags "\
        -X ${PROJECT_ROOT}/internal.Version=${VERSION} \
        -X ${PROJECT_ROOT}/internal.Branch=${BRANCH} \
        -X ${PROJECT_ROOT}/internal.BuildUser=${USER}@${HOST} \
        -X ${PROJECT_ROOT}/internal.BuildDate=${NOW}" \
    -v ${PROJECT_ROOT}/cmd

# Build the application image
FROM alpine:latest
ARG BIN_NAME=turing-router
ENV BIN_NAME ${BIN_NAME}

RUN addgroup -S app && adduser -S app -G app
WORKDIR /app
RUN chown -R app:app /app
COPY --chown=app:app --from=builder /app/bin/* ./
USER app
ENTRYPOINT ./${BIN_NAME}
