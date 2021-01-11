-- Update log_config field value in router_versions table for 'kafka' logger type
-- by adding 'serialization_format' in the 'kafka_config' field value. This is 
-- to support multiple serialization format 'json' / 'protobuf' in newer
-- version of Turing.
--
-- In previous SQL schema, there is no 'serialization_format' field and it is assumed
-- to always be JSON. Hence, this migration script will set the 'serialization_format'
-- for all existing rows to 'json'. 

UPDATE router_versions
SET log_config = (SELECT json_build_object(
    'log_level', log_config -> 'log_level',
    'jaeger_enabled', log_config -> 'jaeger_enabled',
    'result_logger_type', log_config -> 'result_logger_type',
    'custom_metrics_enabled', log_config -> 'custom_metrics_enabled',
    'fiber_debug_log_enabled', log_config -> 'fiber_debug_log_enabled',
    'kafka_config', json_build_object(
        'serialization_format', 'json',
        'topic', log_config -> 'kafka_config' -> 'topic',
        'brokers', log_config -> 'kafka_config' -> 'brokers')
))
WHERE log_config->>'result_logger_type' = 'kafka' AND
      log_config->'kafka_config'->>'serialization_format' IS NULL;
