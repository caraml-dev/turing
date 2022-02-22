-- Migrate Ensembler rows data from new schema to old schema.
-- Note that in the old schema the docker_config column has resource_request and timeout as separate columns.
-- There is no container_runtime_config for "standard" type so ensemblers with that type are unaffected.
-- Hence, this migration involves data loss for ensemblers with type that is "pyfunc"
UPDATE ensemblers
SET docker_config = (SELECT json_build_object(
        'image', image,
        'resource_request', (docker_config -> 'container_runtime_config' ->> 'resource_request')::jsonb,
        'endpoint', endpoint,
        'timeout', (docker_config -> 'container_runtime_config' ->> 'timeout')::jsonb,
        'port', port,
        'env', env,
        'service_account', service_account
    ))

ALTER TABLE ensemblers DROP COLUMN py_func_ref_config;

