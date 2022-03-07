# Deploy Router

To deploy a router, call the `deploy` method on a `Router` instance:

```python
# 8. Deploy the router's *current* configuration (notice how it still deploys the *first* version)
response = my_router.deploy()
```

The return value would be a `dict` containing the `router_id` of the router.