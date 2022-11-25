ARG BASE_IMAGE

FROM ${BASE_IMAGE}

ARG MODEL_URL
RUN echo "[Credentials]\\ngs_service_key_file=${GOOGLE_APPLICATION_CREDENTIALS}" > ~/.boto
RUN cat ~/.boto
RUN gsutil -m cp -r ${MODEL_URL} .
ARG FOLDER_NAME
RUN /bin/bash -c ". activate ${CONDA_ENVIRONMENT} && conda env update --name ${CONDA_ENVIRONMENT} --file /${HOME}/${FOLDER_NAME}/conda.yaml"
