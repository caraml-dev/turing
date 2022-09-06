-- Add column autoscaling_policy to the existing component configs.
-- enrichers
ALTER TABLE enrichers ADD autoscaling_policy jsonb NOT NULL DEFAULT '{"metric": "concurrency", "target": "1"}';
ALTER TABLE enrichers ALTER COLUMN autoscaling_policy DROP DEFAULT;
-- router versions
ALTER TABLE router_versions ADD autoscaling_policy jsonb NOT NULL DEFAULT '{"metric": "concurrency", "target": "1"}';
ALTER TABLE router_versions ALTER COLUMN autoscaling_policy DROP DEFAULT;
-- ensemblers
UPDATE ensembler_configs set docker_config = docker_config || jsonb '{"metric": "concurrency", "target": "1"}'
    WHERE type='docker';
UPDATE ensembler_configs set pyfunc_config = pyfunc_config || jsonb '{"metric": "concurrency", "target": "1"}'
    WHERE type='pyfunc';
