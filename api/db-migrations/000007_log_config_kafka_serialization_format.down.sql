-- Remove 'serialization_format' field in 'kafka_config'

UPDATE router_versions
SET log_config = (SELECT log_config #- '{kafka_config,serialization_format}')
WHERE log_config->>'result_logger_type' = 'kafka';
