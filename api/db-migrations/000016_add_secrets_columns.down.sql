-- Remove secrets column for enrichers
ALTER TABLE enrichers DROP COLUMN secrets;

-- Remove secrets field in docker_config and pyfunc_config columns for ensemblers
UPDATE ensembler_configs set docker_config = docker_config - 'secrets' WHERE docker_config IS NOT NULL;

UPDATE ensembler_configs set pyfunc_config = pyfunc_config - 'secrets' WHERE pyfunc_config IS NOT NULL;
