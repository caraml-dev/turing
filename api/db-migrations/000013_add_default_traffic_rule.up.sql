ALTER TABLE router_versions ADD default_traffic_rule jsonb;

WITH cleaned_router_versions AS (
    SELECT id, CASE WHEN traffic_rules = 'null' THEN '[]'::jsonb ELSE traffic_rules END AS traffic_rules
    FROM router_versions
),
rule_tbl AS (
    SELECT id, jsonb_array_elements(traffic_rules) AS rule FROM cleaned_router_versions
),
-- explode the route names from each rule
rule_route_tbl AS (
    SELECT id, jsonb_array_elements(rule->'routes') AS routes FROM rule_tbl
),
-- explode the original routes column
route_tbl AS (
    SELECT id, jsonb_array_elements(routes) AS route FROM router_versions
),
original_route_tbl AS (
    SELECT id, route -> 'id' AS route_name FROM route_tbl
),
dangling_routes AS (
    SELECT
        id, jsonb_agg(route_name) dangling_routes
    FROM (
        SELECT * FROM original_route_tbl
        EXCEPT
        SELECT * FROM rule_route_tbl
    ) t
    GROUP BY t.id
),
exploded_traffic_rules AS (
    SELECT
        id
        , position
        , elem
    FROM cleaned_router_versions, jsonb_array_elements(traffic_rules) WITH ordinality arr(elem, position)
    WHERE id IN (SELECT id FROM dangling_routes)
),
exploded_traffic_rules_with_dangling_routes AS (
    SELECT
        exploded_traffic_rules.id
        , position
        , CASE
            WHEN elem -> 'routes' IS NOT NULL THEN jsonb_build_object(
                'conditions', elem -> 'conditions',
                'routes', (elem -> 'routes' || dangling_routes),
                'name', elem -> 'name'
            )
            ELSE NULL
            END AS traffic_rule
    FROM
        exploded_traffic_rules
        LEFT JOIN dangling_routes ON exploded_traffic_rules.id = dangling_routes.id
),
joined_updated_traffic_rules AS (
    SELECT
        id,
        json_agg(traffic_rule ORDER BY position) traffic_rules_updated
    FROM exploded_traffic_rules_with_dangling_routes GROUP BY id
)

UPDATE router_versions AS t 
SET default_traffic_rule = jsonb_build_object('routes', dangling_routes),
    traffic_rules = traffic_rules_updated
FROM 
    (
        SELECT
            joined_updated_traffic_rules.id,
            dangling_routes,
            traffic_rules_updated
        FROM
            joined_updated_traffic_rules
        LEFT JOIN dangling_routes ON joined_updated_traffic_rules.id = dangling_routes.id
    ) AS t2
WHERE t.id = t2.id;
