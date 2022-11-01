ALTER TABLE router_versions
    ALTER COLUMN resource_request SET NOT NULL,
    ALTER COLUMN resource_request SET DEFAULT '{
    "min_replica": 1,
    "max_replica": 1,
    "cpu_request": "200m",
    "memory_request": "256Mi"
  }'::jsonb;