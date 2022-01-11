import re
import turing.generated.models
from typing import List, Dict
from turing.generated.model_utils import OpenApiModel
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.common.common import EnvVar
from turing.router.config.common.schemas import DockerImageSchema, TimeoutSchema


class Enricher:
    def __init__(self,
                 id: int,
                 image: str,
                 resource_request: ResourceRequest,
                 endpoint: str,
                 timeout: str,
                 port: int,
                 env: List['EnvVar'],
                 service_account: str = None):
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
        DockerImageSchema.verify_schema(image)
        self._image = image

    @property
    def resource_request(self) -> ResourceRequest:
        return self._resource_request

    @resource_request.setter
    def resource_request(self, resource_request):
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
    def timeout(self, timeout):
        TimeoutSchema.verify_schema(timeout)
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

        return turing.generated.models.Enricher(
            id=self.id,
            image=self.image,
            resource_request=self.resource_request.to_open_api(),
            endpoint=self.endpoint,
            timeout=self.timeout,
            port=self.port,
            env=[env_var.to_open_api() for env_var in self.env],
            service_account=self.service_account
        )
