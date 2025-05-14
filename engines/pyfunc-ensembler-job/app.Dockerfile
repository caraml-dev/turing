ARG BASE_IMAGE

FROM ${BASE_IMAGE}

ARG MLFLOW_ARTIFACT_STORAGE_TYPE

ARG MODEL_URL
ARG GOOGLE_APPLICATION_CREDENTIALS
ARG MODEL_DEPENDENCIES_URL

ARG AWS_ACCESS_KEY_ID
ARG AWS_SECRET_ACCESS_KEY
ARG AWS_DEFAULT_REGION
ARG AWS_ENDPOINT_URL

RUN if [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "gcs" ]; then  \
        if [ ! -z "${GOOGLE_APPLICATION_CREDENTIALS}" ]; then \
            gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS} > gcloud-auth.txt; \
        else echo "GOOGLE_APPLICATION_CREDENTIALS is not set. Skipping gcloud auth command." > gcloud-auth.txt; \
        fi \
    elif [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "s3" ]; then \
       echo "S3 credentials used"; \
    else \
       echo "No credentials are used"; \
    fi

# Download model dependencies
ARG MODEL_DEPENDENCIES_URL
RUN if [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "gcs" ]; then  \
        gsutil cp ${MODEL_DEPENDENCIES_URL} conda.yaml; \
    elif [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "s3" ]; then \
        S3_KEY=${MODEL_DEPENDENCIES_URL##*s3://}; \
        aws s3api get-object --bucket ${S3_KEY%%/*} --key ${S3_KEY#*/} conda.yaml; \
    else \
        echo "No credentials are used"; \
    fi

# Update conda.yaml to add turing-sdk
ARG TURING_DEP_CONSTRAINT
RUN process_conda_env.sh conda.yaml "turing-pyfunc-ensembler-job" "${TURING_DEP_CONSTRAINT}"
RUN /bin/bash -c "conda env create --name ${CONDA_ENV_NAME} --file ./conda.yaml"

# Download model artifact
RUN if [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "gcs" ]; then  \
        gsutil -m cp -r ${MODEL_URL} .; \
    elif [ "${MLFLOW_ARTIFACT_STORAGE_TYPE}" = "s3" ]; then \
        aws s3 cp ${MODEL_URL} artifacts --recursive; \
    else \
        echo "No credentials are used"; \
    fi

ARG FOLDER_NAME
RUN /bin/bash -c ". activate ${CONDA_ENV_NAME}"
