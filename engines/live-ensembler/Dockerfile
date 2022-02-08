FROM continuumio/miniconda3

RUN wget -qO- https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-367.0.0-linux-x86_64.tar.gz  | tar xzf -
ENV PATH=$PATH:/google-cloud-sdk/bin
COPY . .
COPY ../../sdk /sdk

RUN conda env create -f ./environment.yaml && \
    rm -rf /root/.cache

RUN /bin/bash -c ". activate live-ensembler && \
    conda env update --name live-ensembler --file /ensembler/conda.yaml && \
    python -m pyfunc_ensembler_runner --model_dir /ensembler --dry_run"

CMD ["/bin/bash", "./run.sh"]
