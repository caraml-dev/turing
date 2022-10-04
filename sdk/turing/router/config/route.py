import turing.generated.models
from turing.generated.model_utils import OpenApiModel
from urllib.parse import urlparse
from dataclasses import dataclass
from enum import Enum
from typing import Union

class RouteProtocol(Enum):
    """
    Protocol of Route
    """

    GRPC = "GRPC"
    HTTP = "HTTP"

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.RouteProtocol(self.value)

@dataclass
class Route:
    """
    Class to create a new Route object

    :param id: route's name
    :param endpoint: endpoint of the route. Must be a valid URL
    :param timeout: timeout indicating the duration past which the request execution will end
    """

    id: str
    endpoint: str
    timeout: str
    protocol: RouteProtocol
    service_method: str

    def __init__(self, id: str, endpoint: str, timeout: str, protocol:Union[RouteProtocol, str]=RouteProtocol.HTTP, service_method=None, **kwargs):
        self.id = id
        self.timeout = timeout
        self.protocol = RouteProtocol(protocol)
        self.service_method = service_method
        self.endpoint = endpoint

    @property
    def id(self) -> str:
        return self._id

    @id.setter
    def id(self, id: str):
        self._id = id

    @property
    def endpoint(self) -> str:
        return self._endpoint

    @endpoint.setter
    def endpoint(self, endpoint: str):
        if self._protocol == RouteProtocol.HTTP: 
            Route._verify_endpoint(endpoint)
        self._endpoint = endpoint

    @property
    def timeout(self) -> str:
        return self._timeout

    @timeout.setter
    def timeout(self, timeout: str):
        self._timeout = timeout

    @property
    def protocol(self) -> RouteProtocol:
        return self._protocol

    @protocol.setter
    def protocol(self, protocol: Union[RouteProtocol, str]):
        if isinstance(protocol, str):
            self._protocol = RouteProtocol(protocol)
        else:
            self._protocol = protocol

    @property
    def service_method(self) -> str:
        return self._service_method

    @service_method.setter
    def service_method(self, service_method: str):
        if self._protocol == RouteProtocol.GRPC: 
            Route._verify_service_method(service_method)
        self._service_method = service_method

    @classmethod
    def _verify_endpoint(cls, endpoint: str):
        """Rudimentary url checker"""
        parse_result = urlparse(endpoint)

        if parse_result.scheme and parse_result.netloc:
            return
        else:
            raise InvalidUrlException(f"Invalid url entered: {endpoint}")

    @classmethod
    def _verify_service_method(cls, service_method: str):
        """Rudimentary grpc service method checker"""
        if service_method is None or len(service_method) == 0:
            raise MissingServiceMethodException(f"Missing service_method for grpc route")

    def to_open_api(self) -> OpenApiModel:
        kwargs = {}
        if self.service_method is not None:
            kwargs["service_method"] = self.service_method

        return turing.generated.models.Route(
            id=self.id,
            type="PROXY",
            endpoint=self.endpoint,
            timeout=self.timeout,
            protocol=self.protocol.to_open_api(),
            **kwargs
        )


class InvalidUrlException(Exception):
    pass

class MissingServiceMethodException(Exception):
    pass

class DuplicateRouteException(Exception):
    pass

class InvalidRouteException(Exception):
    pass
