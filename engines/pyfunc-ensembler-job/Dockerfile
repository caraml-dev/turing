ARG SPARK_VERSION=v3.0.0

FROM gcr.io/spark-operator/spark-py:$SPARK_VERSION

ARG SPARK_OPERATOR_VERSION=v1beta2-1.2.2-3.0.0
ARG SPARK_BQ_CONNECTOR_VERSION=0.19.1

# Reset to root to include spark dependencies
USER 0
# Setup dependencies for Google Cloud Storage access.
RUN rm $SPARK_HOME/jars/guava-14.0.1.jar
ADD https://repo1.maven.org/maven2/com/google/guava/guava/23.0/guava-23.0.jar \
    $SPARK_HOME/jars
# Add the connector jar needed to access Google Cloud Storage using the Hadoop FileSystem API.
ADD https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector-hadoop2-2.0.1.jar \
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

# Install gcloud SDK
ENV PATH=$PATH:/google-cloud-sdk/bin
ARG GCLOUD_VERSION=332.0.0

RUN wget -qO- \
    https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-${GCLOUD_VERSION}-linux-x86_64.tar.gz | \
    tar xzf - -C /

COPY ./entrypoint.sh /opt/entrypoint.sh

# Configure non-root user
ARG username=spark
ARG spark_uid=185
ARG spark_gid=100
ENV USER $username
ENV UID $spark_uid
ENV GID $spark_gid
ENV HOME /home/$USER
RUN adduser --disabled-password --uid $UID --gid $GID --home $HOME $USER

# Switch to Spark user
USER ${USER}
WORKDIR $HOME

# Install miniconda
ARG CONDA_VERSION=py38_4.9.2
ARG CONDA_MD5=122c8c9beb51e124ab32a0fa6426c656
ENV CONDA_DIR $HOME/miniconda3
ENV PATH $CONDA_DIR/bin:$PATH

RUN wget --quiet https://repo.anaconda.com/miniconda/Miniconda3-${CONDA_VERSION}-Linux-x86_64.sh -O miniconda.sh && \
    echo "$CONDA_MD5  miniconda.sh" > miniconda.md5 && \
    if ! md5sum --status -c miniconda.md5; then exit 1; fi && \
    sh miniconda.sh -b -p $CONDA_DIR && \
    rm miniconda.sh miniconda.md5 && \
    echo "source $CONDA_DIR/etc/profile.d/conda.sh" >> $HOME/.bashrc && \
    $CONDA_DIR/bin/conda clean -afy

# Copy PySpark application
ENV PROJECT_DIR $HOME/batch-ensembler
RUN mkdir -p $PROJECT_DIR
COPY --chown=$UID:$GID . $PROJECT_DIR/
# Hack to ensure that it's compatible with the ../../sdk stated in requirements.txt
COPY ./temp-deps/sdk $PROJECT_DIR/../../sdk

# Setup base conda environment
ARG PYTHON_VERSION
ENV CONDA_ENVIRONMENT turing-batch-ensembler
RUN conda env create -f $PROJECT_DIR/env-${PYTHON_VERSION}.yaml -n $CONDA_ENVIRONMENT && \
    rm -rf $HOME/.cache

RUN echo "conda activate $CONDA_ENVIRONMENT" >> $HOME/.bashrc

ENV PATH $CONDA_DIR/envs/$CONDA_ENVIRONMENT/bin:$PATH

ENTRYPOINT ["/opt/entrypoint.sh"]
