-- Remove 'serialization_format' field in 'kafka_config'

UPDATE router_versions
SET log_config = (SELECT json_build_object(
    'log_level', log_config -> 'log_level',
        'jaeger_enabled', log_config -> 'jaeger_enabled',
        'result_logger_type', log_config -> 'result_logger_type',
        'custom_metrics_enabled', log_config -> 'custom_metrics_enabled',
        'fiber_debug_log_enabled', log_config -> 'fiber_debug_log_enabled',
        'kafka_config', json_build_object(
            'topic', log_config -> 'kafka_config' -> 'topic',
            'brokers', log_config -> 'kafka_config' -> 'brokers')
))
WHERE log_config->>'result_logger_type' = 'kafka';
