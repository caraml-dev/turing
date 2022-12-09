-- remove not null constraint from enrichers
ALTER TABLE enrichers ALTER COLUMN resource_request DROP NOT NULL;
-- remove not null constraint from routers
ALTER TABLE router_versions ALTER COLUMN resource_request DROP NOT NULL;
-- resource requests are stored as jsonb in ensembler_configs and do not have a null constraint to be removed