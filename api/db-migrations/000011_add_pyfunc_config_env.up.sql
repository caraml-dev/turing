-- Migrate pyfunc_config data from old schema to new schema (with new 'env' field set as an empty array by default)
UPDATE ensembler_configs
SET pyfunc_config = (SELECT json_build_object(
        'env', '[]'::jsonb,
        'timeout', pyfunc_config ->> 'timeout',
        'project_id', (pyfunc_config ->> 'project_id')::int,
        'ensembler_id', (pyfunc_config ->> 'ensembler_id')::int,
        'resource_request', (pyfunc_config ->> 'resource_request')::jsonb
    ))
WHERE type ='pyfunc';
