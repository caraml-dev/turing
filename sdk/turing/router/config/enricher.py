import turing.generated.models
from dataclasses import dataclass
from typing import List, Union, Dict
from turing.generated.model_utils import OpenApiModel
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.common.env_var import EnvVar


@dataclass
class Enricher:
    """
    Class to create a new Enricher

    :param image: registry and name of the image
    :param resource_request: ResourceRequest instance containing configs related to the resources required
    :param endpoint: endpoint URL of the enricher
    :param timeout: request timeout which when exceeded, the request to the enricher will be terminated
    :param port: port number exposed by the container
    :param env: environment variables required by the container
    :param id: id of the enricher
    :param service_account: optional service account for the Docker deployment
    """
    image: str
    resource_request: ResourceRequest
    endpoint: str
    timeout: str
    port: int
    env: List['EnvVar']
    id: int = None
    service_account: str = None

    def __init__(self,
                 image: str,
                 resource_request: ResourceRequest,
                 endpoint: str,
                 timeout: str,
                 port: int,
                 env: List['EnvVar'],
                 id: int = None,
                 service_account: str = None,
                 **kwargs):
        self.id = id
        self.image = image
        self.resource_request = resource_request
        self.endpoint = endpoint
        self.timeout = timeout
        self.port = port
        self.env = env
        self.service_account = service_account

    @property
    def id(self) -> int:
        return self._id

    @id.setter
    def id(self, id: int):
        self._id = id

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
    def resource_request(self, resource_request: Union[ResourceRequest, Dict[str, Union[str, int]]]):
        if isinstance(resource_request, ResourceRequest):
            self._resource_request = resource_request
        elif isinstance(resource_request, dict):
            self._resource_request = ResourceRequest(**resource_request)
        else:
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
    def env(self, env: Union[List['EnvVar'], List[Dict[str, str]]]):
        if isinstance(env, list):
            if all(isinstance(env_var, EnvVar) for env_var in env):
                self._env = env
            elif all(isinstance(env_var, dict) for env_var in env):
                self._env = [EnvVar(**env_var) for env_var in env]
            else:
                self._env = env
        else:
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
        if self.id is not None:
            kwargs['id'] = self.id
        if self.service_account is not None:
            kwargs['service_account'] = self.service_account

        return turing.generated.models.Enricher(
            image=self.image,
            resource_request=self.resource_request.to_open_api(),
            endpoint=self.endpoint,
            timeout=self.timeout,
            port=self.port,
            env=[env_var.to_open_api() for env_var in self.env],
            **kwargs
        )
