ARG BASE_IMAGE

FROM ${BASE_IMAGE} as builder

ARG MLFLOW_ARTIFACT_STORAGE_TYPE

ARG MODEL_URL
ARG FOLDER_NAME
ARG GOOGLE_APPLICATION_CREDENTIALS

ARG AWS_ACCESS_KEY_ID
ARG AWS_SECRET_ACCESS_KEY
ARG AWS_DEFAULT_REGION
ARG AWS_ENDPOINT_URL

RUN if [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "gcs" ]; then  \
        if [ ! -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]; then \
            gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}; \
        fi \
    elif [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "s3" ]; then \
       echo "S3 credentials used"; \
    else \
       echo "No credentials are used"; \
    fi

RUN if [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "gcs" ]; then  \
        gsutil -m cp -r ${MODEL_URL} .; \
    elif [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "s3" ]; then \
        aws s3 cp ${MODEL_URL} artifacts --recursive ; \
    else \
        echo "No credentials are used"; \
    fi

RUN /bin/bash -c "conda env update --name ${CONDA_ENV_NAME} --file ./${FOLDER_NAME}/conda.yaml"

ENV FOLDER_NAME=$FOLDER_NAME
SHELL ["/bin/bash", "-c"]
ENTRYPOINT . activate ${CONDA_ENV_NAME} && \
  python -m pyfunc_ensembler_runner --mlflow_ensembler_dir ./${FOLDER_NAME} -l INFO
