-- Create secrets column for enrichers
ALTER TABLE enrichers ADD COLUMN secrets jsonb NOT NULL DEFAULT '[]'::jsonb;

-- Create secrets field in docker_config and pyfunc_config columns for ensemblers
UPDATE ensembler_configs SET docker_config = jsonb_set(docker_config, '{secrets}', '[]'::jsonb) WHERE docker_config IS NOT NULL AND docker_config->'secrets' IS NULL;

UPDATE ensembler_configs SET pyfunc_config = jsonb_set(pyfunc_config, '{secrets}', '[]'::jsonb) WHERE pyfunc_config IS NOT NULL AND pyfunc_config->'secrets' IS NULL;
