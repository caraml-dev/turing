# RouterVersion

A `RouterVersion` represents a single version (and configuration) of a Turing Router. Just as `Router` objects, they 
should almost never be created manually by using their constructor.

Besides assessing attributes of a `RouterVersion` object directly, which will allow you to access attributes such as 
`id`, `version`, `created_at`, `updated_at`, `environment_name`, `status`, `name`, `monitoring_url`, `log_config`, 
you may also consider retrieving the entire router configuration from a specific `RouterVersion` object as a 
`RouterConfig` for further manipulation:

```python
my_config = router_version.get_config()
```
