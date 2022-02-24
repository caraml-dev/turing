ARG BASE_IMAGE

FROM ${BASE_IMAGE} as builder

ARG MODEL_URL
ARG FOLDER_NAME

RUN gsutil -m cp -r ${MODEL_URL} .

# Install dependencies required by the user-defined ensembler
RUN /bin/bash -c ". activate ${CONDA_ENV_NAME}"
RUN conda env update --name ${CONDA_ENV_NAME} --file ./${FOLDER_NAME}/conda.yaml

# Use conda-pack to create a standalone environment
# in /venv:
RUN conda-pack -n ${CONDA_ENV_NAME} -o /tmp/env.tar && \
  mkdir /venv && cd /venv && tar xf /tmp/env.tar && \
  rm /tmp/env.tar

RUN /venv/bin/conda-unpack

FROM debian:bullseye-slim

COPY --from=builder /pyfunc_ensembler_runner ./pyfunc_ensembler_runner
COPY --from=builder /venv /venv

RUN /bin/bash -c ". /venv/bin/activate"

ENTRYPOINT python -m pyfunc_ensembler_runner --mlflow_ensembler_dir /ensembler -l INFO
