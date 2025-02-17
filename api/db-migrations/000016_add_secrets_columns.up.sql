-- Create secrets column for enrichers
ALTER TABLE enrichers ADD COLUMN secrets jsonb;
