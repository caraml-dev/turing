from dataclasses import dataclass

import turing.generated.models
from typing import List, Dict, Union
from turing.generated.model_utils import OpenApiModel
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.common.env_var import EnvVar


@dataclass
class RouterEnsemblerConfig:
    """
    Class to create a new RouterEnsemblerConfig

    :param type: type of the ensembler; must be one of {'standard', 'docker'}
    :param id: id of the ensembler
    :param standard_config: EnsemblerStandardConfig instance containing mappings between routes and treatments
    :param docker_config: EnsemblerDockerConfig instance containing configs for the docker ensembler
    """
    type: str
    id: int = None
    standard_config: turing.generated.models.EnsemblerStandardConfig = None
    docker_config: turing.generated.models.EnsemblerDockerConfig = None

    def __init__(self,
                 type: str,
                 id: int = None,
                 standard_config: turing.generated.models.EnsemblerStandardConfig = None,
                 docker_config: turing.generated.models.EnsemblerDockerConfig = None,
                 **kwargs):
        self.id = id
        self.type = type
        self.standard_config = standard_config
        self.docker_config = docker_config

    @property
    def id(self) -> int:
        return self._id

    @id.setter
    def id(self, id: int):
        self._id = id

    @property
    def type(self) -> str:
        return self._type

    @type.setter
    def type(self, type: str):
        assert type in {'standard', 'docker'}
        self._type = type

    @property
    def standard_config(self) -> turing.generated.models.EnsemblerStandardConfig:
        return self._standard_config

    @standard_config.setter
    def standard_config(self, standard_config: Union[turing.generated.models.EnsemblerStandardConfig, Dict]):
        if isinstance(standard_config, turing.generated.models.EnsemblerStandardConfig):
            self._standard_config = standard_config
        elif isinstance(standard_config, dict):
            openapi_standard_config = standard_config.copy()
            openapi_standard_config['experiment_mappings'] = [
                turing.generated.models.EnsemblerStandardConfigExperimentMappings(**mapping)
                for mapping in standard_config['experiment_mappings']
            ]
            self._standard_config = turing.generated.models.EnsemblerStandardConfig(**openapi_standard_config)
        else:
            self._standard_config = standard_config

    @property
    def docker_config(self) -> turing.generated.models.EnsemblerDockerConfig:
        return self._docker_config

    @docker_config.setter
    def docker_config(self, docker_config: turing.generated.models.EnsemblerDockerConfig):
        if isinstance(docker_config, turing.generated.models.EnsemblerDockerConfig):
            self._docker_config = docker_config
        elif isinstance(docker_config, dict):
            openapi_docker_config = docker_config.copy()
            openapi_docker_config['resource_request'] = \
                turing.generated.models.ResourceRequest(**openapi_docker_config['resource_request'])
            openapi_docker_config['env'] = [turing.generated.models.EnvVar(**env_var) for env_var in openapi_docker_config['env']]
            self._docker_config = turing.generated.models.EnsemblerDockerConfig(
                **openapi_docker_config
            )
        else:
            self._docker_config = docker_config

    def to_open_api(self) -> OpenApiModel:
        kwargs = {}

        if self.standard_config is not None:
            kwargs["standard_config"] = self.standard_config
        if self.docker_config is not None:
            kwargs["docker_config"] = self.docker_config

        return turing.generated.models.RouterEnsemblerConfig(
            type=self.type,
            **kwargs
        )


@dataclass
class DockerRouterEnsemblerConfig(RouterEnsemblerConfig):
    def __init__(self,
                 id: int,
                 image: str,
                 resource_request: ResourceRequest,
                 endpoint: str,
                 timeout: str,
                 port: int,
                 env: List['EnvVar'],
                 service_account: str = None):
        """
        Method to create a new Docker ensembler

        :param id: id of the ensembler
        :param image: registry and name of the image
        :param resource_request: ResourceRequest instance containing configs related to the resources required
        :param endpoint: endpoint URL of the ensembler
        :param timeout: request timeout which when exceeded, the request to the ensembler will be terminated
        :param port: port number exposed by the container
        :param env: environment variables required by the container
        :param service_account: optional service account for the Docker deployment
        """
        self.image = image
        self.resource_request = resource_request
        self.endpoint = endpoint
        self.timeout = timeout
        self.port = port
        self.env = env
        self.service_account = service_account
        super().__init__(id=id, type="docker")

    @property
    def image(self) -> str:
        return self._image

    @image.setter
    def image(self, image: str):
        self._image = image

    @property
    def resource_request(self) -> ResourceRequest:
        return self._resource_request

    @resource_request.setter
    def resource_request(self, resource_request: ResourceRequest):
        self._resource_request = resource_request

    @property
    def endpoint(self) -> str:
        return self._endpoint

    @endpoint.setter
    def endpoint(self, endpoint: str):
        self._endpoint = endpoint

    @property
    def timeout(self) -> str:
        return self._timeout

    @timeout.setter
    def timeout(self, timeout: str):
        self._timeout = timeout

    @property
    def port(self) -> int:
        return self._port

    @port.setter
    def port(self, port: int):
        self._port = port

    @property
    def env(self) -> List['EnvVar']:
        return self._env

    @env.setter
    def env(self, env: List['EnvVar']):
        self._env = env

    @property
    def service_account(self) -> str:
        return self._service_account

    @service_account.setter
    def service_account(self, service_account: str):
        self._service_account = service_account

    def to_open_api(self) -> OpenApiModel:
        assert all(isinstance(env_var, EnvVar) for env_var in self.env)

        kwargs = {}
        if self.service_account is not None:
            kwargs['service_account'] = self.service_account

        self.docker_config = turing.generated.models.EnsemblerDockerConfig(
            image=self.image,
            resource_request=self.resource_request.to_open_api(),
            endpoint=self.endpoint,
            timeout=self.timeout,
            port=self.port,
            env=[env_var.to_open_api() for env_var in self.env],
            **kwargs
        )
        return super().to_open_api()


@dataclass
class StandardRouterEnsemblerConfig(RouterEnsemblerConfig):
    def __init__(self,
                 id: int,
                 experiment_mappings: List[Dict[str, str]]):
        """
        Method to create a new standard ensembler

        :param id: id of the ensembler
        :param experiment_mappings: configured mappings between routes and treatments
        """
        self.experiment_mappings = experiment_mappings
        super().__init__(id=id, type="standard")

    @property
    def experiment_mappings(self) -> List[Dict[str, str]]:
        return self._experiment_mappings

    @experiment_mappings.setter
    def experiment_mappings(self, experiment_mappings: List[Dict[str, str]]):
        StandardRouterEnsemblerConfig._verify_experiment_mappings(experiment_mappings)
        self._experiment_mappings = experiment_mappings

    @classmethod
    def _verify_experiment_mappings(cls, experiment_mappings: List[Dict[str, str]]):
        for experiment_mapping in experiment_mappings:
            try:
                assert isinstance(experiment_mapping["experiment"], str)
                assert isinstance(experiment_mapping["treatment"], str)
                assert isinstance(experiment_mapping["route"], str)
            except (KeyError, AssertionError):
                raise InvalidExperimentMappingException(
                    "Experiment mapping should contain the keys {'experiment', 'treatment', 'route'}; "
                    f"experiment_mapping passed: {experiment_mapping}"
                )

    def to_open_api(self) -> OpenApiModel:
        self.standard_config = turing.generated.models.EnsemblerStandardConfig(
            experiment_mappings=[
                turing.generated.models.EnsemblerStandardConfigExperimentMappings(**experiment_mapping) \
                for experiment_mapping in self.experiment_mappings
            ]
        )
        return super().to_open_api()


class InvalidExperimentMappingException(Exception):
    pass
