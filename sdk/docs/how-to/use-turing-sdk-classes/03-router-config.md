# RouterConfig

`RouterConfig` objects are what you would probably interact most frequently with when using Turing SDK. They 
essentially carry a router's configuration and define the ways in which a router should be run. All of the 
methods that interact with Turing API that involve the updating or creating of routers involve the 
use of `RouterConfig` objects as arguments.

As you would have seen before, a `RouterConfig` object is built using multiple parts:

```python
@dataclass
class RouterConfig:
    """
    Class to create a new RouterConfig. Can be built up from its individual components or initialised instantly
    from an appropriate API response

    :param environment_name: name of the environment
    :param name: name of the router
    :param routes: list of routes used by the router
    :param rules: list of rules used by the router
    :param default_route_id: default route id to be used
    :param experiment_engine: experiment engine config file
    :param resource_request: resources to be provisioned for the router
    :param timeout: request timeout which when exceeded, the request to the router will be terminated
    :param log_config: logging config settings to be used with the router
    :param enricher: enricher config settings to be used with the router
    :param ensembler: ensembler config settings to be used with the router
    """
    environment_name: str
    name: str
    routes: Union[List[Route], List[Dict[str, str]]] = None
    rules: Union[List[TrafficRule], List[Dict]] = None
    default_route_id: str = None
    experiment_engine: Union[ExperimentConfig, Dict] = None
    resource_request: Union[ResourceRequest, Dict[str, Union[str, int]]] = None
    timeout: str = None
    log_config: Union[LogConfig, Dict[str, Union[str, bool, int]]] = None
    enricher: Union[Enricher, Dict] = None
    ensembler: Union[RouterEnsemblerConfig, Dict] = None
```

When constructing a `RouterConfig` object from scratch, it is **highly recommended** that you construct each 
individual component using the Turing SDK classes provided instead of using `dict` objects which do not perform any 
schema validation.

In the following pages of this subsection, we will go through the usage of these individual components separately.
