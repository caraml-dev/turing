# PyFuncEnsembler Server for Real-Time Experiments

PyFuncEnsemblerRunner is a tool for deploying user-defined ensemblers (for use with Turing routers), written in 
MLflow's `pyfunc` flavour.

## Usage
To run the ensembler as a webservice:
```bash
python -m pyfunc_ensembler_runner --mlflow_ensembler_dir $ENSEMBLER_DIR [-l {DEBUG,INFO,WARNING,ERROR,CRITICAL}]

arguments: 
  --mlflow_ensembler_dir <path/to/ensembler/dir/>       Path to the ensembler folder containing the mlflow files
  --log-level <DEBUG||INFO||WARNING||ERROR||CRITICAL>   Set the logging level
  -h, --help                                            Show this help message and exit
```

## Docker Image Building

To create a docker image locally, you'll need to first download the model artifacts from the MLflow's model registry:
```bash
gsutil cp -r gs://[bucket-name]/mlflow/[project_id]/[run_id]/artifacts/ensembler .
```

To build the docker image, run the following:
```bash
make build-image
```
