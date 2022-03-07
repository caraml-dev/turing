# Deploy Router Version

To deploy a specific router version, you would need the version number of the router you would like to deploy. Call 
the `deploy_version` method of your `Router` instance with the version number as an argument:

```python
# 6. Deploy a specific router config version (the first one we created)
response = my_router.deploy_version(first_ver_no)
```

The return value would be a `dict` containing the `router_id` and `version` of the newly deployed version. 