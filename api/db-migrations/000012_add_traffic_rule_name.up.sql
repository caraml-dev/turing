WITH cleaned_router_versions AS (
    SELECT id, CASE WHEN traffic_rules = 'null' THEN null ELSE traffic_rules END AS traffic_rules
    FROM router_versions
),
tbl1 AS (
    SELECT id, position, CONCAT('rule_', position:: text) AS name, elem
    FROM cleaned_router_versions, jsonb_array_elements(traffic_rules) WITH ordinality arr(elem, position)
),
tbl2 AS (
    SELECT id, position, jsonb_build_object(
        'conditions', elem -> 'conditions',
        'routes', elem -> 'routes',
        'name', name
    ) AS traffic_rule FROM tbl1
),
tbl3 AS (
    SELECT id, json_agg(traffic_rule ORDER BY position) AS traffic_rules_updated
    FROM tbl2 GROUP BY id
)
UPDATE router_versions AS t SET traffic_rules = traffic_rules_updated
FROM tbl3
WHERE tbl3.id = t.id;
