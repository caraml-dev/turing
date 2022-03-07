# Delete Router Version

To delete a version of a specific router, call the `delete_version` method on the `Router` instance with the version 
ID as the argument:

```python
# 10. Delete a specific router version of the router
response = my_router.delete_version(latest_ver_no)
```

The return value would be a `dict` containing the `router_id` and the `id` of the deleted version.