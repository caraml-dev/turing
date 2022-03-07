# Get Router Version

To get a specific router version of a router, call the `get_version` method on a `Router` instance with the router 
ID as an argument:

```python
# 9. Get a specific router version of the router
my_router_ver = my_router.get_version(first_ver_no)
```

The return value would be a `RouterVersion` object representing the router version with the corresponding ID.