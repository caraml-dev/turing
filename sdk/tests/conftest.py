import json
from datetime import datetime, timedelta
import pytest
import random
import tests
from turing.ensembler import PyFuncEnsembler
import turing.generated.models
import turing.batch.config
import turing.batch.config.source
import turing.batch.config.sink
import uuid
from tests.fixtures.mlflow import mock_mlflow
from tests.fixtures.gcs import mock_gcs


@pytest.fixture
def turing_api() -> str:
    return "http://turing.local.svc:8080"


@pytest.fixture
def use_google_oauth() -> bool:
    return False


@pytest.fixture
def project():
    return turing.generated.models.Project(
        id=10,
        name=f"project_name",
        mlflow_tracking_url="http://localhost:5000",
        created_at=datetime.now(),
        updated_at=datetime.now()
    )


@pytest.fixture
def active_project(responses, project):
    responses.add(
        method="GET",
        url=f"/v1/projects?name={project.name}",
        body=json.dumps([project], default=tests.json_serializer),
        match_querystring=True,
        status=200,
        content_type="application/json"
    )
    return project


@pytest.fixture
def projects(num_projects):
    return [
        turing.generated.models.Project(
            id=i,
            name=f"project_{i}",
            mlflow_tracking_url="http://localhost:5000",
            created_at=datetime.now() + timedelta(seconds=i + 10),
            updated_at=datetime.now() + timedelta(seconds=i + 10)
        ) for i in range(1, num_projects + 1)]


@pytest.fixture
def generic_ensemblers(project, num_ensemblers):
    return [
        turing.generated.models.GenericEnsembler(
            id=i,
            project_id=project.id,
            type=PyFuncEnsembler.TYPE,
            name=f"test_{i}",
            created_at=datetime.now() + timedelta(seconds=i + 10),
            updated_at=datetime.now() + timedelta(seconds=i + 10)
        ) for i in range(1, num_ensemblers + 1)]


@pytest.fixture
def bucket_name():
    return "bucket-name"


@pytest.fixture
def experiment_name(project, ensembler_name):
    return f"{project.name}/ensemblers/{ensembler_name}"


@pytest.fixture
def experiment_id():
    return random.randint(1, 100)


@pytest.fixture
def run_id():
    return uuid.uuid4().hex


@pytest.fixture
def artifact_uri(bucket_name, experiment_id, run_id):
    return f"gs://{bucket_name}/mlflow/{experiment_id}/{run_id}"


@pytest.fixture
def pyfunc_ensembler(project, ensembler_name, experiment_id, run_id, artifact_uri):
    return turing.generated.models.PyFuncEnsembler(
        id=1,
        project_id=project.id,
        type=PyFuncEnsembler.TYPE,
        name=ensembler_name,
        mlflow_experiment_id=experiment_id,
        mlflow_run_id=run_id,
        artifact_uri=artifact_uri,
        created_at=datetime.now(),
        updated_at=datetime.now()
    )


@pytest.fixture
def ensembling_job_config():
    source = turing.batch.config.source.BigQueryDataset(
        table="project.dataset.features",
        features=["feature_1", "feature_2", "features_3"]
    ).join_on(columns=["feature_1"])

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

    result_config = turing.batch.config.ResultConfig(
        type=turing.batch.config.ResultType.INTEGER,
        column_name="prediction_result"
    )

    sink = turing.batch.config.sink.BigQuerySink(
        table="project.dataset.ensembling_results",
        staging_bucket="staging_bucket"
    ).save_mode(turing.batch.config.sink.SaveMode.OVERWRITE) \
        .select(columns=["feature_1", "feature_2", "prediction_result"])

    resource_request = turing.batch.config.ResourceRequest(
        driver_cpu_request="1",
        driver_memory_request="1G",
        executor_replica=5,
        executor_cpu_request="500Mi",
        executor_memory_request="800M"
    )

    return turing.batch.config.EnsemblingJobConfig(
        source=source,
        predictions=predictions,
        result_config=result_config,
        sink=sink,
        service_account="service-account",
        resource_request=resource_request
    )


@pytest.fixture
def generic_router_status():
    return turing.generated.models.RouterStatus(random.choice(["deployed", "undeployed", "failed", "pending"]))


@pytest.fixture
def generic_router_version_status():
    return turing.generated.models.RouterVersionStatus(random.choice(["deployed", "undeployed", "failed", "pending"]))


@pytest.fixture
def generic_resource_request():
    return turing.generated.models.ResourceRequest(
        min_replica=0,
        max_replica=2,
        cpu_request='500m',
        memory_request='512Mi'
    )


@pytest.fixture
def generic_log_level():
    return turing.generated.models.LogLevel(random.choice(["DEBUG", "INFO", "WARN", "ERROR"]))


@pytest.fixture
def generic_result_logger_type():
    return turing.generated.models.ResultLoggerType(random.choice(["nop", "console", "bigquery", "kafka"]))


@pytest.fixture
def generic_bigquery_config():
    return turing.generated.models.BigQueryConfig(
        table="bigquerytable",
        service_account_secret="abc123"
    )


@pytest.fixture
def generic_kafka_config():
    return turing.generated.models.KafkaConfig(
        brokers="1.1.1.1,2.2.2.2,3.3.3.3",
        topic="new",
        serialization_format=random.choice(["json", "protobuf"])
    )


@pytest.fixture
def log_config(generic_log_level, generic_result_logger_type, generic_bigquery_config, generic_kafka_config):
    return turing.generated.models.RouterVersionLogConfig(
        log_level=generic_log_level,
        custom_metrics_enabled=True,
        fiber_debug_log_enabled=True,
        jaeger_enabled=True,
        result_logger_type=generic_result_logger_type,
        bigquery_config=generic_bigquery_config,
        kafka_config=generic_kafka_config
    )


@pytest.fixture
def generic_route():
    return turing.generated.models.Route(
        id="route_1",
        type="PROXY",
        endpoint="http://models.internal/predict_1",
        annotations={
            "annotation_1": "value_1",
            "annotation_2": ["value_2a", "value_2b"]
        },
        timeout="20ms"
    )


@pytest.fixture
def generic_field_source():
    return turing.generated.models.FieldSource(random.choice(["header", "payload"]))


@pytest.fixture
def generic_traffic_rule_condition(generic_field_source):
    return turing.generated.models.TrafficRuleCondition(
        field_source=generic_field_source,
        field="taxi",
        operator="in",
        values=["departures", "arrivals"]
    )


@pytest.fixture
def generic_traffic_rule(generic_traffic_rule_condition):
    return turing.generated.models.TrafficRule(
        conditions=[generic_traffic_rule_condition],
        routes=["test"]
    )


@pytest.fixture
def generic_ensembler_standard_config():
    return turing.generated.models.EnsemblerStandardConfig(
        experiment_mappings=[
            turing.generated.models.EnsemblerStandardConfigExperimentMappings(
                experiment="experiment-1",
                treatment="treatment-1",
                route="route-1"
            ),
            turing.generated.models.EnsemblerStandardConfigExperimentMappings(
                experiment="experiment-2",
                treatment="treatment-2",
                route="route-2"
            )
        ]
    )


@pytest.fixture
def generic_env_var():
    return turing.generated.models.EnvVar(
        name="env_name",
        value="env_val"
    )


@pytest.fixture
def generic_ensembler_docker_config(generic_resource_request, generic_env_var):
    return turing.generated.models.EnsemblerDockerConfig(
        image="test.io/gods-test/turing-ensembler:0.0.0-build.0",
        resource_request=generic_resource_request,
        endpoint=f"http://localhost:5000/ensembler_endpoint",
        timeout="500ms",
        port=5120,
        env=[generic_env_var],
        service_account="secret-name-for-google-service-account"
    )


@pytest.fixture(params=["standard", "docker"])
def ensembler(request, generic_ensembler_standard_config, generic_ensembler_docker_config):
    ensembler_type = request.param
    return turing.generated.models.RouterEnsemblerConfig(
        id=1,
        type=ensembler_type,
        standard_config=generic_ensembler_standard_config,
        docker_config=generic_ensembler_docker_config,
        created_at=datetime.now() + timedelta(seconds=10),
        updated_at=datetime.now() + timedelta(seconds=10)
    )


@pytest.fixture(params=["nop", "random_engine"])
def experiment_config(request):
    experiment_type = request.param
    if experiment_type == "nop":
        config = {}
    elif experiment_type == "random_engine":
        config = {
            "client": {
                "id": 1,
                "passkey": "abc"
            },
            "experiments": [
                {
                    "client_id": 1,
                    "name": "experiment_1"
                }
            ],
            "variables": {
                "client_variables": [
                    {
                        "variable_name": "version",
                        "required": False
                    }
                ]
            }
        }
    else:
        config = None
    return turing.generated.models.ExperimentConfig(
        type=experiment_type,
        config=config
    )


@pytest.fixture
def router_version(
        generic_router_version_status,
        generic_route,
        generic_traffic_rule,
        experiment_config,
        generic_resource_request,
        log_config,
        ensembler
):
    return turing.generated.models.RouterVersion(
        id=2,
        created_at=datetime.now() + timedelta(seconds=20),
        updated_at=datetime.now() + timedelta(seconds=20),
        router=None,
        version=1,
        status=generic_router_version_status,
        error="NONE",
        image="test.io/gods-test/turing-router:0.0.0-build.0",
        routes=[generic_route for _ in range(2)],
        default_route="http://models.internal/default",
        default_route_id="control",
        rules=[generic_traffic_rule for _ in range(2)],
        experiment_engine=experiment_config,
        resource_request=generic_resource_request,
        timeout="100ms",
        log_config=log_config,
        ensembler=ensembler,
        monitoring_url="https://lookhere.io/"
    )


@pytest.fixture
def generic_routers(project, num_routers, generic_router_status, router_version):
    return [
        turing.generated.models.RouterDetails(
            id=i,
            name=f"router_{i}",
            endpoint=f"http://localhost:5000/endpoint_{i}",
            environment_name=f"env_{i}",
            monitoring_url=f"http://localhost:5000/dashboard_{i}",
            project_id=project.id,
            status=generic_router_status,
            created_at=datetime.now() + timedelta(seconds=i + 10),
            updated_at=datetime.now() + timedelta(seconds=i + 10),
            config=router_version
        ) for i in range(1, num_routers + 1)]
