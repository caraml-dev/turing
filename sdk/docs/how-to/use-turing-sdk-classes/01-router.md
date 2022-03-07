# Router

A `Router` object represents a router that is created on Turing API. It does not (and should not) ever be created 
manually by using its constructor directly. Instead, you should only be manipulating with `Router` instances that 
get returned as a result of using the various `Router` class and instance methods that interact with Turing API.

A `Router` object has attributes such as `id`, `name`, `project_id`, `environment_name`, `monitoring_url`, `status` 
and `endpoint`. It also has a `config` attribute, which is a `RouterConfig` containing the current configuration for 
the router. 

When trying to replicate configuration from an existing router, always retrieve the underlying `RouterConfig` from 
the `Router` instance by accessing its `config` attribute.