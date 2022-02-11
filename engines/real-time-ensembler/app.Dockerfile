ARG BASE_IMAGE

FROM ${BASE_IMAGE} as builder

ARG MODEL_URL
RUN gsutil -m cp -r ${MODEL_URL} .
ARG FOLDER_NAME

# Install dependencies required by the user-defined ensembler
RUN /bin/bash -c ". activate ${CONDA_ENVIRONMENT} && conda env update --name ${CONDA_ENVIRONMENT} --file /${HOME}/${FOLDER_NAME}/conda.yaml"

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
