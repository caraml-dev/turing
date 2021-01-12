-- Update log_config field value in router_versions table for 'kafka' logger type
-- by adding 'serialization_format' in the 'kafka_config' field value. This is 
-- to support multiple serialization format 'json' / 'protobuf' in newer
-- version of Turing.
--
-- In previous SQL schema, there is no 'serialization_format' field and it is assumed
-- to always be JSON. Hence, this migration script will set the 'serialization_format'
-- for all existing rows to 'json'. 

UPDATE router_versions
SET log_config = (
    SELECT jsonb_set(log_config, '{kafka_config,serialization_format}', '"json"'::jsonb)
)
WHERE log_config->>'result_logger_type' = 'kafka' AND
      log_config->'kafka_config'->>'serialization_format' IS NULL;
