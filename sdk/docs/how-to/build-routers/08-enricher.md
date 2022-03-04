# Enricher

An `Enricher` object holds configuration needed to define an enricher:

```python
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
```

## EnvVar

To define environment variables for the `env` attribute in an `Enricher`, you would need to define them using the 
`EnvVar` object:

```python
@dataclass
class EnvVar:
    name: str
    value: str
```