import turing.generated.models
from turing.generated.model_utils import OpenApiModel
from turing.router.config.common.schemas import TimeoutSchema
from urllib.parse import urlparse


class Route:
    def __init__(self,
                 id: str,
                 endpoint: str,
                 timeout: str,
                 **kwargs):
        """
        Method to create a new Route object

        :param id: route's name
        :param endpoint: endpoint of the route. Must be a valid URL
        :param timeout: timeout indicating the duration past which the request execution will end
        """
        Route._verify_endpoint(endpoint)

        self.id = id
        self.endpoint = endpoint
        self.timeout = timeout

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
    def endpoint(self, endpoint):
        Route._verify_endpoint(endpoint)
        self._endpoint = endpoint

    @property
    def timeout(self) -> str:
        return self._timeout

    @timeout.setter
    def timeout(self, timeout: str):
        TimeoutSchema.verify_schema(timeout)
        self._timeout = timeout

    @classmethod
    def _verify_endpoint(cls, endpoint):
        """Rudimentary url checker """
        parse_result = urlparse(endpoint)

        if parse_result.scheme and parse_result.netloc:
            return
        else:
            raise InvalidUrlException(f"Invalid url entered: {endpoint}")

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.Route(
            id=self.id,
            type='PROXY',
            endpoint=self.endpoint,
            timeout=self.timeout,
        )


class InvalidUrlException(Exception):
    pass


class DuplicateRouteException(Exception):
    pass
