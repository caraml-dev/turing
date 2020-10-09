FROM debian:buster-slim

RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

RUN mkdir -p /opt/turing_api

COPY . /opt/turing_api
COPY ./bin/turing-api /opt/turing_api

WORKDIR /opt/turing_api
