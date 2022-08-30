UPDATE router_versions
SET traffic_rules = s.new_traffic_rules
FROM (
    SELECT
        jsonb_agg(
            jsonb_build_object(
                'conditions', elem -> 'conditions',
                'routes', elem -> 'routes'
            )
        ) new_traffic_rules
    FROM
        router_versions, jsonb_array_elements(traffic_rules) as elem
) s
