FROM continuumio/miniconda3 AS builder

ARG CONDA_ENV_NAME
ARG APP_NAME

ENV CONDA_ENV_NAME=$CONDA_ENV_NAME
ENV APP_NAME=$APP_NAME

RUN wget -qO- https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-367.0.0-linux-x86_64.tar.gz  | tar xzf -
ENV PATH=$PATH:/google-cloud-sdk/bin

COPY . .
COPY ./temp-deps/sdk ./../../sdk

RUN conda env create -f ./environment.yaml &&  \
    rm -rf /root/.cache

# Install conda-pack:
RUN conda install -c conda-forge conda-pack