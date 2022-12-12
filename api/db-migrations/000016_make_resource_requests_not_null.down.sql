-- remove null resource requests from enrichers
ALTER TABLE enrichers
    ALTER COLUMN resource_request SET NOT NULL,
    ALTER COLUMN resource_request SET DEFAULT '{
    "min_replica": 1,
    "max_replica": 1,
    "cpu_request": "200m",
    "memory_request": "256Mi"
  }'::jsonb;
-- remove null resource requests from routers
ALTER TABLE router_versions
    ALTER COLUMN resource_request SET NOT NULL,
    ALTER COLUMN resource_request SET DEFAULT '{
    "min_replica": 1,
    "max_replica": 1,
    "cpu_request": "200m",
    "memory_request": "256Mi"
  }'::jsonb;
-- remove null resource requests from ensemblers
UPDATE ensembler_configs set docker_config = docker_config || jsonb '{
  "resource_requests": {
    "min_replica": 1,
    "max_replica": 1,SLE
    "cpu_request": "200m",
    "memory_request": "256Mi"
  }
}'
WHERE (type='docker' OR type='pyfunc') AND docker_config IS NOT NULL;
UPDATE ensembler_configs set pyfunc_config = pyfunc_config || jsonb '{
  "resource_requests": {
    "min_replica": 1,
    "max_replica": 1,
    "cpu_request": "200m",
    "memory_request": "256Mi"
  }
}'
WHERE type='pyfunc';
