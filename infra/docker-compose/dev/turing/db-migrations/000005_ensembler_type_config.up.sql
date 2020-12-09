ALTER TABLE ensemblers ADD type text;
ALTER TABLE ensemblers ADD standard_config jsonb;
ALTER TABLE ensemblers ADD docker_config jsonb;

-- Migrate Ensembler rows data from old schema to new schema (with type: docker and docker_config populated
-- from the columns in the old schema)
UPDATE ensemblers
SET type          = 'docker',
    docker_config = (SELECT json_build_object(
        'image', image,
        'resource_request', resource_request,
        'endpoint', endpoint,
        'timeout', timeout,
        'port', port,
        'env', env,
        'service_account', service_account
    ))
WHERE id IS NOT NULL;

-- Drop columns that are no longer used in the new schema
ALTER TABLE ensemblers DROP COLUMN image;
ALTER TABLE ensemblers DROP COLUMN resource_request;
ALTER TABLE ensemblers DROP COLUMN endpoint;
ALTER TABLE ensemblers DROP COLUMN timeout;
ALTER TABLE ensemblers DROP COLUMN port;
ALTER TABLE ensemblers DROP COLUMN env;
ALTER TABLE ensemblers DROP COLUMN service_account;