# PyFuncEnsembler Server for Real-Time Experiments

PyFuncEnsemblerRunner is a tool for deploying user-defined ensemblers (for use with Turing routers), written in 
MLflow's `pyfunc` flavour.

## Usage

```bash
python -m pyfunc_ensembler_runner --mlflow_ensembler_uri $ENSEMBLER_URI
```

## Docker Image Building

To create a docker image locally, you'll need to first download the model artifacts from the MLflow's model registry:
```bash
gsutil cp -r gs://[bucket-name]/mlflow/[project_id]/[run_id]/artifacts/ensembler .
```

To build the docker image, run the following:
```bash
docker build -t my_pyfunc_ensembler:latest -f Dockerfile .
```

To run the ensembler service
```bash
docker run -e 
```