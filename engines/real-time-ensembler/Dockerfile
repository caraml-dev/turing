FROM continuumio/miniconda3 AS builder

RUN wget -qO- https://dl.google.com/dl/cloudsdk/channels/rapid/downloads/google-cloud-sdk-367.0.0-linux-x86_64.tar.gz  | tar xzf -
ENV PATH=$PATH:/google-cloud-sdk/bin
COPY . .
COPY ./temp-deps/sdk ./../../sdk

RUN conda env create -f ./environment.yaml &&  \
    conda env update --name real-time-ensembler --file /ensembler/conda.yaml && \
    rm -rf /root/.cache

# Install conda-pack:
RUN conda install -c conda-forge conda-pack

# Use conda-pack to create a standalone enviornment
# in /venv:
RUN conda-pack -n real-time-ensembler -o /tmp/env.tar && \
  mkdir /venv && cd /venv && tar xf /tmp/env.tar && \
  rm /tmp/env.tar

RUN /venv/bin/conda-unpack

FROM debian:latest

COPY --from=builder /ensembler ./ensembler
COPY --from=builder /pyfunc_ensembler_runner ./pyfunc_ensembler_runner
COPY --from=builder /run.sh /run.sh
COPY --from=builder /venv /venv

RUN /bin/bash -c ". /venv/bin/activate && \
    python -m pyfunc_ensembler_runner --mlflow_ensembler_dir /ensembler --dry_run"

CMD ["/bin/bash", "./run.sh"]
