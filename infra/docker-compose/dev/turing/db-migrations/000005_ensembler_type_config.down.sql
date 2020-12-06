ALTER TABLE ensemblers ADD image varchar(128) NOT NULL DEFAULT '';
ALTER TABLE ensemblers ADD resource_request jsonb NOT NULL DEFAULT json_object('{}');
ALTER TABLE ensemblers ADD endpoint varchar(128) NOT NULL DEFAULT '';
ALTER TABLE ensemblers ADD timeout varchar(20) NOT NULL DEFAULT '';
ALTER TABLE ensemblers ADD port integer NOT NULL DEFAULT '0';
ALTER TABLE ensemblers ADD env jsonb;
ALTER TABLE ensemblers ADD service_account text;

-- Migrate Ensembler rows data from new schema to old schema.
-- Note that in the old schema the columns are referring to only docker config.
-- There is no config for "standard" type.
-- Hence, this migration involves data loss for ensembler with type that is not "docker"
UPDATE ensemblers
SET image            = docker_config ->> 'image',
    resource_request = (docker_config ->> 'resource_request')::jsonb,
    endpoint         = docker_config ->> 'endpoint',
    timeout          = docker_config ->> 'timeout',
    port             = (docker_config ->> 'port')::int,
    env              = (docker_config ->> 'env')::jsonb,
    service_account  = docker_config ->> 'service_account'
WHERE type = 'docker';

ALTER TABLE ensemblers DROP COLUMN type;
ALTER TABLE ensemblers DROP COLUMN standard_config;
ALTER TABLE ensemblers DROP COLUMN docker_config;

