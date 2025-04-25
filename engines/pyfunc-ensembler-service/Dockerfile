FROM condaforge/miniforge3:23.3.1-1 AS builder

RUN apt-get update && apt-get install unzip

ARG APP_NAME
ARG CONDA_ENV_NAME

ENV APP_NAME=$APP_NAME
ENV CONDA_ENV_NAME=$CONDA_ENV_NAME

# Install yq
ENV YQ_VERSION=v4.42.1
RUN wget https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64 -O /usr/bin/yq && \
    chmod +x /usr/bin/yq

# Install gcloud SDK
RUN wget -qO- https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-367.0.0-linux-x86_64.tar.gz  | tar xzf -
ENV PATH=$PATH:/google-cloud-sdk/bin

# Install aws CLI
RUN wget -q https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip && unzip awscli-exe-linux-x86_64.zip && ./aws/install

COPY . .
COPY ./temp-deps/sdk ./../../sdk
COPY process_conda_env.sh /bin/process_conda_env.sh
