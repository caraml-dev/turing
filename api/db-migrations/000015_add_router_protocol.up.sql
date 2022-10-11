-- Create type for Router protocol
CREATE TYPE protocol as ENUM ('HTTP_JSON', 'UPI_V1');

-- Add Router protocol
ALTER TABLE router_versions ADD protocol protocol NOT NULL DEFAULT 'HTTP_JSON';