FROM alpine:3

WORKDIR /app
RUN apk add --no-cache curl bash && \
    curl -Lo /usr/bin/kubectl "https://dl.k8s.io/release/v1.22.16/bin/linux/amd64/kubectl" && \
    chmod +x /usr/bin/kubectl

COPY . .
