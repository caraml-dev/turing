import time
import turing
import turing.batch
import turing.batch.config
from samples.common import MyEnsembler

SERVICE_ACCOUNT_NAME = "service-account@gcp-project.iam.gserviceaccount.com"


def main(turing_api: str, project: str):
    # Initialize Turing client:
    turing.set_url(turing_api)
    turing.set_project(project)

    # Save pyfunc ensembler in Turing's backend:
    ensembler = turing.PyFuncEnsembler.create(
        name="my-ensembler",
        ensembler_instance=MyEnsembler(),
        conda_env={
            'dependencies': [
                'python>=3.8.0',
                # other dependencies, if required
            ]
        }
    )
    print("Ensembler created:\n", ensembler)

    # Or fetch existing ensembler by its ID:
    # ensembler_id = < ENSEMBLER_ID >
    # ensembler = turing.PyFuncEnsembler.get_by_id(ensembler_id)

    # Define configuration of the batch ensembling job

    # Configure datasource, that contains input features:
    source = turing.batch.config.source.BigQueryDataset(
        table="project.dataset.features",
        features=["feature_1", "feature_2", "features_3"]
    ).join_on(columns=["feature_1"])

    # Configure dataset(s), that contain predictions of individual models:
    predictions = {
        'model_odd':
            turing.batch.config.source.BigQueryDataset(
                table="project.dataset.scores_model_odd",
                features=["feature_1", "prediction_score"]
            ).join_on(columns=["feature_1"]).select(columns=["prediction_score"]),

        'model_even':
            turing.batch.config.source.BigQueryDataset(
                query="""
                    SELECT feature_1, prediction_score
                    FROM `project.dataset.scores_model_even`
                    WHERE target_date = DATE("2021-03-15", "Asia/Jakarta")
                """,
                options={
                    "viewsEnabled": "true",
                    "materializationDataset": "my_dataset"
                }
            ).join_on(columns=["feature_1"]).select(columns=["prediction_score"])
    }

    # Configure ensembling result:
    result_config = turing.batch.config.ResultConfig(
        type=turing.batch.config.ResultType.INTEGER,
        column_name="prediction_result"
    )

    # Configure destination, where ensembling results will be stored:
    sink = turing.batch.config.sink.BigQuerySink(
        table="project.dataset.ensembling_results",
        staging_bucket="staging_bucket"
    ).save_mode(turing.batch.config.sink.SaveMode.OVERWRITE) \
        .select(columns=["feature_1", "feature_2", "prediction_result"])

    # (Optional) Configure resources allocation for the job execution:
    resource_request = turing.batch.config.ResourceRequest(
        driver_cpu_request="1",
        driver_memory_request="1G",
        executor_replica=5,
        executor_cpu_request="500Mi",
        executor_memory_request="800M"
    )

    # (Optional) Configure environment variables here
    env_vars = {
        "SOME_VAR": "SOME_VALUE",
    }

    # Submit the job for execution:
    job = ensembler.submit_job(
        turing.batch.config.EnsemblingJobConfig(
            source=source,
            predictions=predictions,
            result_config=result_config,
            sink=sink,
            service_account=SERVICE_ACCOUNT_NAME,
            resource_request=resource_request,
            env_vars=env_vars,
        )
    )
    print(job)

    # You can also retrieve the instance of existing job by its ID:
    # job_id = < JOB_ID >
    # job = turing.batch.EnsemblingJob.get_by_id(job_id=job_id)
    #
    # # Or list all ensembling jobs within the project
    # jobs = turing.batch.EnsemblingJob.list(status=[
    #     turing.batch.EnsemblingJobStatus.PENDING,
    #     turing.batch.EnsemblingJobStatus.RUNNING,
    # ])

    # Refresh the status of the job
    for i in range(3):
        time.sleep(5)
        job.refresh()
        print(f"Refresh #{i+1}: {job}")

    # It's also possible to terminate a running job:
    job.terminate()
    print(f"Job's termination in process: {job}")


if __name__ == '__main__':
    import fire
    fire.Fire(main)
