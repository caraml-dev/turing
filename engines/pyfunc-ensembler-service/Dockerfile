FROM continuumio/miniconda3 AS builder

RUN wget -qO- https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-367.0.0-linux-x86_64.tar.gz  | tar xzf -
ENV PATH=$PATH:/google-cloud-sdk/bin
COPY . .
COPY ./temp-deps/sdk ./../../sdk

RUN conda env create -f ./environment.yaml &&  \
    conda env update --name pyfunc-ensembler-service --file /ensembler/conda.yaml && \
    rm -rf /root/.cache

# Install conda-pack:
RUN conda install -c conda-forge conda-pack