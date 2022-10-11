-- Remove Router protocol
ALTER TABLE router_versions DROP COLUMN protocol;

-- Remove type for Router and Route protocol
DROP TYPE IF EXISTS protocol;