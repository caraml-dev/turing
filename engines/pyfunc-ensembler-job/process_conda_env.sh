#!/bin/bash

# This script is designed to manipulate a Conda environment file to add turing dependencies.
#
# Usage:
# ./process_conda_env.sh CONDA_ENV_PATH TURING_DEP TURING_DEP_CONSTRAINT
#
# Input:
# CONDA_ENV_PATH: Path to the Conda environment YAML file.
# TURING_DEP: The dependency to be added or removed.
# TURING_DEP_CONSTRAINT: The constraint for the dependency, such as a version specification.


CONDA_ENV_PATH="$1"
TURING_DEP="$2"
TURING_DEP_CONSTRAINT="$3"

echo "Processing conda environment file: ${CONDA_ENV_PATH}"
echo "Current conda environment file content:"
cat "${CONDA_ENV_PATH}"

# Remove `mlflow` constraint
yq --inplace 'del(.dependencies[].pip[] | select(. == "mlflow*"))' "${CONDA_ENV_PATH}"
yq --inplace ".dependencies[].pip += [\"mlflow\"]" "${CONDA_ENV_PATH}"

# Remove `turing-sdk` from conda's pip dependencies
yq --inplace 'del(.dependencies[].pip[] | select(. == "turing-sdk*"))' "${CONDA_ENV_PATH}"

# Add `${TURING_DEP}` with its constaint (`${TURING_DEP_CONSTRAINT}`) to conda's pip dependencies, if not exist
yq --inplace "with(.dependencies[].pip; select(all_c(. != \"*${TURING_DEP}*\")) | . += [\"${TURING_DEP}${TURING_DEP_CONSTRAINT}\"] )" "${CONDA_ENV_PATH}"

echo "Processed conda environment file content:"
cat "${CONDA_ENV_PATH}"
