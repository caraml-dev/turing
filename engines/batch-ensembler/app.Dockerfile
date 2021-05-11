ARG BASE_IMAGE

FROM ${BASE_IMAGE}

ARG MODEL_URL
RUN gsutil -m cp -r ${MODEL_URL} .
ARG FOLDER_NAME
RUN /bin/bash -c ". activate ${CONDA_ENVIRONMENT} && conda env update --name ${CONDA_ENVIRONMENT} --file /${HOME}/${FOLDER_NAME}/conda.yaml"
