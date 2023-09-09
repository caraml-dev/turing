ARG BASE_IMAGE

FROM ${BASE_IMAGE} as builder

ARG MODEL_URL
ARG FOLDER_NAME
ARG GOOGLE_APPLICATION_CREDENTIALS

# Run docker build using the credentials if provided
RUN if [[-z "$GOOGLE_APPLICATION_CREDENTIALS"]]; then gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}; fi
RUN gsutil -m cp -r ${MODEL_URL} .

RUN /bin/bash -c ". activate ${CONDA_ENV_NAME} && \
  conda env update --name ${CONDA_ENV_NAME} --file ./${FOLDER_NAME}/conda.yaml"

SHELL ["/bin/bash", "-c"]
ENTRYPOINT python -m pyfunc_ensembler_runner --mlflow_ensembler_dir ./${FOLDER_NAME} -l INFO
