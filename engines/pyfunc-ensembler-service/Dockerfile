FROM condaforge/miniforge3:23.3.1-1 AS builder

RUN apt-get update && apt-get install unzip

ARG APP_NAME
ARG CONDA_ENV_NAME
ARG PYTHON_VERSION

ENV APP_NAME=$APP_NAME
ENV CONDA_ENV_NAME=$CONDA_ENV_NAME

# Install gcloud SDK
RUN wget -qO- https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-367.0.0-linux-x86_64.tar.gz  | tar xzf -
ENV PATH=$PATH:/google-cloud-sdk/bin
# Install aws CLI
RUN wget -q https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip && unzip awscli-exe-linux-x86_64.zip && ./aws/install

COPY . .
COPY ./temp-deps/sdk ./../../sdk

RUN conda env create -f ./env-${PYTHON_VERSION}.yaml -n $CONDA_ENV_NAME &&  \
    rm -rf /root/.cache
