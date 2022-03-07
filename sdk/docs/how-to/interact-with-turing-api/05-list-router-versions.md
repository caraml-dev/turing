# List Router Versions

To list all the versions of a router, simply run the following:

```python
# 5. List all the router config versions of your router
my_router_versions = my_router.list_versions()
```

The return value would be a list of `RouterVersion` objects representing the router versions that have been 
created for the router.