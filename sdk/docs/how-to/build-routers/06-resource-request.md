# ResourceRequest

A `ResourceRequest` class carries information related to the resources that should be allocated to a particular 
component, e.g. router, ensembler, enricher, etc., and is defined by 4 attributes, `min_replica`, `max_replica`, 
`cpu_request`, `memory_request`:

```python
@dataclass
class ResourceRequest:
    min_allowed_replica: ClassVar[int] = 0
    max_allowed_replica: ClassVar[int] = 20
    min_replica: int
    max_replica: int
    cpu_request: str
    memory_request: str
```

Note that the units for CPU and memory requests are measured in cpu units and bytes respectively. You may wish to 
read more about how these are measured [here](https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/). 