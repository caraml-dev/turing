ARG BASE_IMAGE

FROM ${BASE_IMAGE}

ARG MODEL_URL
ARG GOOGLE_APPLICATION_CREDENTIALS
# Run docker build using the credentials if provided
RUN if [[-z "$GOOGLE_APPLICATION_CREDENTIALS"]]; then gcloud auth activate-service-account --key-file=${GOOGLE_APPLICATION_CREDENTIALS}; fi
RUN gsutil -m cp -r ${MODEL_URL} .
ARG FOLDER_NAME
RUN /bin/bash -c ". activate ${CONDA_ENVIRONMENT} && conda env update --name ${CONDA_ENVIRONMENT} --file /${HOME}/${FOLDER_NAME}/conda.yaml"
