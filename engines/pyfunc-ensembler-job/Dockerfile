FROM apache/spark-py:v3.1.3

# Switch to user root so we can add additional jars and configuration files.
USER root

ENV SPARK_OPERATOR_VERSION=v1beta2-1.3.7-3.1.1
ENV SPARK_BQ_CONNECTOR_VERSION=0.27.0

RUN apt-get update && apt-get install unzip

# Setup dependencies for Google Cloud Storage access.
RUN rm $SPARK_HOME/jars/guava-14.0.1.jar
ADD https://repo1.maven.org/maven2/com/google/guava/guava/23.0/guava-23.0.jar \
    $SPARK_HOME/jars
# Add the connector jar needed to access Google Cloud Storage using the Hadoop FileSystem API.
ADD https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector-hadoop2-2.2.8.jar \
    $SPARK_HOME/jars
ADD https://repo1.maven.org/maven2/com/google/cloud/spark/spark-bigquery-with-dependencies_2.12/${SPARK_BQ_CONNECTOR_VERSION}/spark-bigquery-with-dependencies_2.12-${SPARK_BQ_CONNECTOR_VERSION}.jar \
    $SPARK_HOME/jars
RUN chmod 644 -R $SPARK_HOME/jars/*

# Setup for the Prometheus JMX exporter.
RUN mkdir -p /etc/metrics/conf
# Add the Prometheus JMX exporter Java agent jar for exposing metrics sent to the JmxSink to Prometheus.
ADD https://repo1.maven.org/maven2/io/prometheus/jmx/jmx_prometheus_javaagent/0.11.0/jmx_prometheus_javaagent-0.11.0.jar /prometheus/
RUN chmod 644 /prometheus/jmx_prometheus_javaagent-0.11.0.jar

ADD https://raw.githubusercontent.com/GoogleCloudPlatform/spark-on-k8s-operator/$SPARK_OPERATOR_VERSION/spark-docker/conf/metrics.properties /etc/metrics/conf
ADD https://raw.githubusercontent.com/GoogleCloudPlatform/spark-on-k8s-operator/$SPARK_OPERATOR_VERSION/spark-docker/conf/prometheus.yaml /etc/metrics/conf
RUN chmod 644 -R /etc/metrics/conf/*

ENV LANG=C.UTF-8 LC_ALL=C.UTF-8

# Install wget and other libraries required by Miniconda3
RUN apt-get update --allow-releaseinfo-change-suite -q && \
    apt-get install -q -y \
    bzip2 \
    ca-certificates \
    git \
    libglib2.0-0 \
    libsm6 \
    libxext6 \
    libxrender1 \
    wget \
    && apt-get clean

# Install yq
ENV YQ_VERSION=v4.42.1
RUN wget https://github.com/mikefarah/yq/releases/download/${YQ_VERSION}/yq_linux_amd64 -O /usr/bin/yq && \
    chmod +x /usr/bin/yq

# Install gcloud SDK
ENV PATH=$PATH:/google-cloud-sdk/bin
ARG GCLOUD_VERSION=410.0.0
RUN wget -qO- \
    https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz | \
    tar xzf - -C /

# Install aws CLI
RUN wget -q https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip && unzip awscli-exe-linux-x86_64.zip && ./aws/install

COPY ./entrypoint.sh /opt/entrypoint.sh

# Configure non-root user
ARG username=spark
ARG spark_uid=185
ARG spark_gid=100
ENV USER=$username
ENV UID=$spark_uid
ENV GID=$spark_gid
ENV HOME=/home/$USER
RUN adduser --disabled-password --uid $UID --gid $GID --home $HOME $USER

# Switch to Spark user
USER ${USER}
WORKDIR $HOME

# Install miniconda
ENV CONDA_DIR=${HOME}/miniconda3
ENV PATH=${CONDA_DIR}/bin:$PATH
ENV MINIFORGE_VERSION=23.3.1-1

RUN wget --quiet https://github.com/conda-forge/miniforge/releases/download/${MINIFORGE_VERSION}/Miniforge3-${MINIFORGE_VERSION}-Linux-x86_64.sh -O miniconda.sh && \
    /bin/bash miniconda.sh -b -p ${CONDA_DIR} && \
    rm ~/miniconda.sh && \
    $CONDA_DIR/bin/conda clean -afy && \
    echo "source $CONDA_DIR/etc/profile.d/conda.sh" >> $HOME/.bashrc

# Copy PySpark application
ENV PROJECT_DIR=$HOME/batch-ensembler
RUN mkdir -p $PROJECT_DIR
COPY --chown=$UID:$GID . $PROJECT_DIR/
# Hack to ensure that it's compatible with the ../../sdk stated in requirements.txt
COPY --chown=$UID:$GID ./temp-deps/sdk $PROJECT_DIR/../../sdk

# Copy conda env processor script
COPY process_conda_env.sh /bin/process_conda_env.sh

ENTRYPOINT ["/opt/entrypoint.sh"]
