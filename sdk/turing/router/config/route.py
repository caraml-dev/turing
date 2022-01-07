import turing.generated.models
from turing.generated.model_utils import OpenApiModel
from urllib.parse import urlparse


class Route:
    def __init__(self,
                 id: str,
                 endpoint: str,
                 timeout: int):
        """
        Method to create a new Route object

        :param id: route's name
        :param endpoint: endpoint of the route. Must be a valid URL
        :param timeout: timeout indicating the duration past which the request execution will end
        """
        if not Route.is_valid_url(endpoint):
            raise InvalidUrlException(f"Invalid url entered: {endpoint}")

        timeout_string = str(timeout) + "ms"

        self._id = id
        self._endpoint = endpoint
        self._timeout = timeout_string

    @property
    def id(self) -> str:
        return self._id

    @property
    def endpoint(self) -> str:
        return self._endpoint

    @property
    def timeout(self) -> str:
        return self._timeout

    @classmethod
    def is_valid_url(cls, url):
        """Rudimentary url checker """
        parse_result = urlparse(url)

        if parse_result.scheme and parse_result.netloc:
            return True
        else:
            return False

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
