ALTER TABLE enrichers DROP COLUMN autoscaling_policy;
ALTER TABLE router_versions DROP COLUMN autoscaling_policy;
UPDATE ensembler_configs set docker_config = docker_config - 'autoscaling_policy' WHERE type='docker';
UPDATE ensembler_configs set pyfunc_config = pyfunc_config - 'autoscaling_policy' WHERE type='pyfunc';
