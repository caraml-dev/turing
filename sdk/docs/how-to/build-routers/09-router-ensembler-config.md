# RouterEnsemblerConfig

Ensembling for Turing Routers is done through Standard, Docker or Pyfunc ensemblers, and its configuration is 
managed by the `RouterEnsemblerConfig` class. Three helper classes (child classes of `RouterEnsemblerConfig`) have been 
created to assist you in constructing these objects:

```python
@dataclass
class StandardRouterEnsemblerConfig(RouterEnsemblerConfig):
    def __init__(self,
                 experiment_mappings: List[Dict[str, str]]):
        """
        Method to create a new standard ensembler

        :param experiment_mappings: configured mappings between routes and treatments
        """
        self.experiment_mappings = experiment_mappings
        super().__init__(type="standard")
```

```python
@dataclass
class DockerRouterEnsemblerConfig(RouterEnsemblerConfig):
    def __init__(self,
                 image: str,
                 resource_request: ResourceRequest,
                 endpoint: str,
                 timeout: str,
                 port: int,
                 env: List['EnvVar'],
                 service_account: str = None):
        """
        Method to create a new Docker ensembler

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
        super().__init__(type="docker")
```

```python
@dataclass
class PyfuncRouterEnsemblerConfig(RouterEnsemblerConfig):
    def __init__(self,
                 project_id: int,
                 ensembler_id: int,
                 timeout: str,
                 resource_request: ResourceRequest):
        """
        Method to create a new Pyfunc ensembler

        :param project_id: project id of the current project
        :param ensembler_id: ensembler_id of the ensembler
        :param resource_request: ResourceRequest instance containing configs related to the resources required
        :param timeout: request timeout which when exceeded, the request to the ensembler will be terminated
        """
        self.project_id = project_id
        self.ensembler_id = ensembler_id
        self.resource_request = resource_request
        self.timeout = timeout
        super().__init__(type="pyfunc")
```
