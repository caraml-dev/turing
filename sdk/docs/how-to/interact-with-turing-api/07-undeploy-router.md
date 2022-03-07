# Undeploy Router

To undeploy a router, call the `undeploy` method on a `Router` instance:

```python
# 7. Undeploy the current active router configuration
response = my_router.undeploy()
```

The return value would be a `dict` containing the `router_id` of the router.