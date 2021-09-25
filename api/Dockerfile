# Build turing-api binary
FROM golang:1.14-alpine as api-builder
ARG API_BIN_NAME=turing-api

ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64

ENV PROJECT_ROOT=github.com/gojek/turing/api/turing

WORKDIR /app
COPY . .

RUN go build \
    -mod=vendor \
    -o ./bin/${API_BIN_NAME} \
    -v ${PROJECT_ROOT}/cmd

# Clean image with turing-api binary and production build of turing-ui
FROM alpine:3.12

ARG API_BIN_NAME=turing-api
ENV API_BIN_NAME ${API_BIN_NAME}

ENV TURING_PORT "8080"

EXPOSE ${TURING_PORT}

RUN addgroup -S app && adduser -S app -G app && mkdir /app && chown -R app:app /app

COPY --chown=app:app --from=api-builder /app/bin/* /app
COPY --chown=app:app --from=api-builder /app/api /app/api

USER app
WORKDIR /app

ENTRYPOINT ["sh", "-c"]
CMD ["./${API_BIN_NAME} \"$@\""]
