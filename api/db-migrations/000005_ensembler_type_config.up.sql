ALTER TABLE ensemblers ADD py_func_ref_config jsonb;

-- Migrate docker_config columns from old schema to new schema (with type: docker and docker_config populated
-- with resource_request and timeout as separate columns)
UPDATE ensemblers
SET docker_config = (SELECT json_build_object(
        'image', image,
        'container_runtime_config', json_build_object(
            'resource_request', resource_request,
            'timeout', timeout
            )
        'endpoint', endpoint,
        'port', port,
        'env', env,
        'service_account', service_account,
    ))
WHERE type = 'docker'