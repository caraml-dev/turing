# Turing Batch Ensembler

Pyspark application for running batch ensembling jobs in Turing.

## Usage
```shell
usage: main.py [-h] --job-spec <job-spec.yaml> [-l {DEBUG,INFO,WARNING,ERROR,CRITICAL}]

Run a PySpark batch ensembling job

optional arguments:
  --job-spec  <path/to/job-spec/yaml>                   Path to the ensembling job YAML file specification
  --log-level <DEBUG||INFO||WARNING||ERROR||CRITICAL>   Set the logging level
  -h, --help                                            show this help message and exit
```

### Job Specification
Application expects to receive a path of a job specification via `--job-spec` argument.
Job specification contains the information on how to configure data source and sink for 
the job, as well as the user-defined ensembler configuration. The complete schema of 
the job specification is defined at this [proto](./api/proto/v1/batch_ensembling_job.proto).

#### Job Spec example
```yaml
version: v1
kind: BatchEnsemblingJob
metadata:
  name: batch-ensembling
  annotations:
    spark/spark.jars: "https://storage.googleapis.com/hadoop-lib/gcs/gcs-connector-hadoop2-2.0.1.jar"
    spark/spark.jars.packages: "com.google.cloud.spark:spark-bigquery-with-dependencies_2.12:0.19.1"
    hadoopConfiguration/fs.gs.impl: "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFileSystem"
    hadoopConfiguration/fs.AbstractFileSystem.gs.impl: "com.google.cloud.hadoop.fs.gcs.GoogleHadoopFS"    
spec:
  source:
    dataset:
      type: BQ
      bq_config:
        query: |-
          WITH customer_filter AS (
            SELECT customer_id, target_date
            FROM `project.dataset.customers_list`
            WHERE target_date = DATE("2021-03-15", "Asia/Jakarta")
          ),

          serving_features AS (
            SELECT *
            FROM `project.dataset.customers_features`
          )

          SELECT
            serving_features.*,
          FROM
            serving_features
            INNER JOIN customer_filter USING (customer_id)
        options:
          viewsEnabled: "true"
          materializationDataset: dataset
    join_on:
      - customer_id
      - target_date
  predictions:
    model_a:
      dataset:
        type: BQ
        bq_config:
          table: "project.dataset.predictions_model_a"
          features:
            - customer_id
            - target_date
            - predictions
      columns:
        - predictions
      join_on:
        - customer_id
        - target_date
    model_b:
      dataset:
        type: BQ
        bq_config:
          query: |-
            SELECT *
            FROM `project.dataset.predictions_model_b`
      columns:
        - predictions
      join_on:
        - customer_id
        - target_date
  ensembler:
    uri: gs://bucket-name/my-ensembler/artifacts/ensembler
    result:
      column_name: prediction_score
      type: ARRAY
      item_type: FLOAT
  sink:
    type: BQ
    save_mode: OVERWRITE
    columns:
      - customer_id as customerId
      - target_date
      - results
    bq_config:
      table: project.dataset.ensembling_results
      staging_bucket: bucket-name
      options:
        partitionField: target_date
```

## Development
### Requirements:
* Python 3.8
* Miniconda3 ([Instructions](https://docs.conda.io/en/latest/miniconda.html))
* Docker ([Instructions](https://docs.docker.com/install/))

### Setup Dev Dependencies
```shell script
make setup
```

### Run Unit tests
```shell script
make test-unit
```

### Build a Docker image
```shell script
make build-image
```
