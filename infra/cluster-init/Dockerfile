FROM alpine:3

COPY --from=alpine/helm:3.7.1 /usr/bin/helm /usr/bin/helm
COPY --from=bitnami/kubectl:1.21.5 /opt/bitnami/kubectl/bin/kubectl /usr/bin/kubectl

WORKDIR /app
RUN apk add --no-cache curl bash

COPY . .
