import re
import turing.generated.models
from typing import List, Dict
from turing.generated.model_utils import OpenApiModel
from turing.router.config.resource_request import ResourceRequest


class RouterEnsemblerConfig:
    def __init__(self,
                 id: int,
                 type: str,
                 standard_config: turing.generated.models.EnsemblerStandardConfig = None,
                 docker_config: turing.generated.models.EnsemblerDockerConfig = None):
        """
        Method to create a new RouterEnsemblerConfig

        :param id: id of the ensembler
        :param type: type of the ensembler; must be one of {'standard', 'docker'}
        :param standard_config: EnsemblerStandardConfig instance containing mappings between routes and treatments
        :param docker_config: EnsemblerDockerConfig instance containing configs for the docker ensembler
        """
        assert type in {'standard', 'docker'}
        self._id = id
        self._type = type
        self._standard_config = standard_config
        self._docker_config = docker_config

    @property
    def id(self) -> int:
        return self._id

    @property
    def type(self) -> str:
        return self._type

    @property
    def standard_config(self):
        return self._standard_config

    @property
    def docker_config(self):
        return self._docker_config

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.RouterEnsemblerConfig(
            id=self.id,
            type=self.type,
            standard_config=self.standard_config,
            docker_config=self.docker_config,
        )


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
        :param service_account: optional service account for the Docker deployment
        """
        DockerRouterEnsemblerConfig._verify_image(image)
        DockerRouterEnsemblerConfig._verify_timeout(timeout)
        DockerRouterEnsemblerConfig._verify_env(env)

        self._image = image
        self._resource_request = resource_request
        self._endpoint = endpoint
        self._timeout = timeout
        self._port = port
        self._env = env
        self._service_account = service_account
        super().__init__(id=id, type="docker")

    @property
    def image(self) -> str:
        return self._image

    @image.setter
    def image(self, image):
        DockerRouterEnsemblerConfig._verify_image(image)
        self._image = image

    @property
    def resource_request(self) -> ResourceRequest:
        return self._resource_request

    @property
    def endpoint(self) -> str:
        return self._endpoint

    @property
    def timeout(self) -> str:
        return self._timeout

    @timeout.setter
    def timeout(self, timeout):
        DockerRouterEnsemblerConfig._verify_timeout(timeout)
        self._timeout = timeout

    @property
    def port(self) -> int:
        return self._port

    @property
    def env(self) -> List['EnvVar']:
        return self._env

    @env.setter
    def env(self, env):
        DockerRouterEnsemblerConfig._verify_env(env)
        self._env = env

    @property
    def service_account(self) -> str:
        return self._service_account

    @classmethod
    def _verify_image(cls, image):
        matched = re.fullmatch(
            r"^([a-z0-9]+(?:[._-][a-z0-9]+)*(?::\d{2,5})?\/)?([a-z0-9]+(?:[._-][a-z0-9]+)*\/)*([a-z0-9]+(?:[._-][a-z0-9]+)*)(?::[a-z0-9]+(?:[._-][a-z0-9]+)*)?$",
            image,
            re.IGNORECASE
        )
        if bool(matched) is False:
            raise InvalidImageException(
                f"Valid Docker Image value should be provided, e.g. kennethreitz/httpbin:latest; "
                f"image passed: {image}"
            )

    @classmethod
    def _verify_timeout(cls, timeout):
        matched = re.fullmatch(r"^[0-9]+(ms|s|m|h)$", timeout)
        if bool(matched) is False:
            raise InvalidTimeoutException(
                f"Valid duration is required; timeout passed: {timeout}"
            )

    @classmethod
    def _verify_env(cls, env):
        for env_var in env:
            EnvVar.verify_name(env_var.name)

    def to_open_api(self) -> OpenApiModel:
        DockerRouterEnsemblerConfig._verify_env(self.env)

        self._docker_config = turing.generated.models.EnsemblerDockerConfig(
            image=self.image,
            resource_request=self.resource_request.to_open_api(),
            endpoint=self.endpoint,
            timeout=self.timeout,
            port=self.port,
            env=[env_var.to_open_api() for env_var in self.env],
            service_account=self.service_account
        )
        return super().to_open_api()


class EnvVar:
    def __init__(self,
                 name: str,
                 value: str):
        self._name = name
        self._value = value

    @property
    def name(self) -> str:
        return self._name

    @property
    def value(self) -> str:
        return self._value

    @classmethod
    def verify_name(cls, name):
        matched = re.fullmatch(r"^[a-z0-9_]*$", name, re.IGNORECASE)
        if bool(matched) is False:
            raise InvalidEnvironmentVariableNameException(
                f"The name of a variable can contain only alphanumeric character or the underscore; "
                f"name passed: {name}"
            )

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.EnvVar(
            name=self.name,
            value=self.value
        )


class InvalidImageException(Exception):
    pass


class InvalidTimeoutException(Exception):
    pass


class InvalidEnvironmentVariableNameException(Exception):
    pass


class StandardRouterEnsemblerConfig(RouterEnsemblerConfig):
    def __init__(self,
                 id: int,
                 experiment_mappings: List[Dict[str, str]]):
        """
        Method to create a new standard ensembler

        :param id: id of the ensembler
        :param experiment_mappings: configured mappings between routes and treatments
        """
        StandardRouterEnsemblerConfig._verify_experiment_mappings(experiment_mappings)
        self._experiment_mappings = experiment_mappings
        super().__init__(id=id, type="standard")

    @property
    def experiment_mappings(self) -> List[Dict[str, str]]:
        return self._experiment_mappings

    @experiment_mappings.setter
    def experiment_mappings(self, experiment_mappings):
        StandardRouterEnsemblerConfig._verify_experiment_mappings(experiment_mappings)
        self._experiment_mappings = experiment_mappings

    @classmethod
    def _verify_experiment_mappings(cls, experiment_mappings):
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
        self._standard_config = turing.generated.models.EnsemblerStandardConfig(
            experiment_mappings=[
                turing.generated.models.EnsemblerStandardConfigExperimentMappings(**experiment_mapping) \
                for experiment_mapping in self.experiment_mappings
            ]
        )
        return super().to_open_api()


class InvalidExperimentMappingException(Exception):
    pass
