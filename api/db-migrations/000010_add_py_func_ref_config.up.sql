ALTER TABLE ensembler_configs ADD py_func_ref_config jsonb;

-- Migrate docker_config columns from old schema to new schema (with type: docker and docker_config populated
-- with resource_request and timeout as separate columns)
UPDATE ensembler_configs
SET docker_config = (SELECT json_build_object(
        'image', docker_config ->> 'image',
        'container_runtime_config', json_build_object(
                'resource_request', (docker_config ->> 'resource_request')::jsonb,
                'timeout', docker_config ->> 'timeout'
            ),
        'endpoint', docker_config ->> 'endpoint',
        'port', (docker_config ->> 'port')::int,
        'env', (docker_config ->> 'env')::jsonb,
        'service_account', docker_config ->> 'service_account'
    ))
WHERE type='docker';