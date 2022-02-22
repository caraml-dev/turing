-- Migrate Ensembler rows data from new schema to old schema.
-- Note that in the old schema the docker_config column has resource_request and timeout as separate columns.
-- There is no container_runtime_config for "standard" type so ensemblers with that type are unaffected.
-- Hence, this migration involves data loss for ensemblers with type that is "pyfunc"
UPDATE ensembler_configs
SET docker_config = (SELECT json_build_object(
        'image', docker_config ->> 'image',
        'resource_request', docker_config -> 'container_runtime_config' ->> 'resource_request',
        'endpoint', docker_config ->> 'endpoint',
        'timeout', docker_config -> 'container_runtime_config' ->> 'timeout',
        'port', (docker_config ->> 'port')::int,
        'env', (docker_config ->> 'env')::jsonb,
        'service_account', docker_config ->> 'service_account'
    ))
WHERE type = 'docker';

ALTER TABLE ensembler_configs DROP COLUMN py_func_ref_config;