FROM alpine:3

COPY --from=bitnami/kubectl:1.22.16 /opt/bitnami/kubectl/bin/kubectl /usr/bin/kubectl

WORKDIR /app
RUN apk add --no-cache curl bash

COPY . .
