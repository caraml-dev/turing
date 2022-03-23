-- Migrate Ensembler rows data from old schema to new schema (with new 'env' field set as an empty array by default)
UPDATE ensembler_configs
SET docker_config = (SELECT json_build_object(
                                    'env', '[]'::jsonb,
                                    'port', docker_config ->> 'port',
                                    'image', docker_config ->> 'image',
                                    'timeout', docker_config ->> 'timeout',
                                    'endpoint', docker_config ->> 'endpoint',
                                    'service_account', docker_config ->> 'service_account',
                                    'resource_request', (docker_config ->> 'resource_request')::jsonb,
                                )),
    pyfunc_config = (SELECT json_build_object(
                                    'env', '[]'::jsonb,
                                    'timeout', pyfunc_config ->> 'timeout',
                                    'project_id', pyfunc_config ->> 'project_id',
                                    'ensembler_id', pyfunc_config ->> 'ensembler_id',
                                    'resource_request', (pyfunc_config ->> 'resource_request')::jsonb,
                                ))
WHERE type ='pyfunc';
