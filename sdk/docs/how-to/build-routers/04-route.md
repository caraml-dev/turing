# Route

Creating a route with Turing SDK is just as simple as doing it on the UI; one only needs to specify the `id`, 
`endpoint`, and `timeout` of the route:

```python
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
```
