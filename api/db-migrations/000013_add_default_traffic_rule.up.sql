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
),
-- Need to join from cleaned_router_versions since exploded_traffic_rules will exclude router versions which have no traffic rules
exploded_traffic_rules_with_dangling_routes AS (
    SELECT
        cleaned_router_versions.id
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
        cleaned_router_versions
        LEFT JOIN exploded_traffic_rules ON cleaned_router_versions.id = exploded_traffic_rules.id
        LEFT JOIN dangling_routes ON cleaned_router_versions.id = dangling_routes.id
),
-- FILTER clause added to prevent aggregating null values into [null] jsonb
joined_updated_traffic_rules AS (
    SELECT
        id,
        json_agg(traffic_rule ORDER BY position) FILTER (where traffic_rule IS NOT NULL) traffic_rules_updated
    FROM exploded_traffic_rules_with_dangling_routes GROUP BY id
)

UPDATE router_versions AS t 
SET default_traffic_rule = jsonb_build_object('routes', dangling_routes),
    traffic_rules = (CASE WHEN traffic_rules_updated IS NOT NULL THEN traffic_rules_updated ELSE '[]'::json END)
FROM 
    (
        SELECT
            router_versions.id,
            dangling_routes,
            traffic_rules_updated
        FROM
            router_versions
        LEFT JOIN dangling_routes ON router_versions.id = dangling_routes.id
        LEFT JOIN joined_updated_traffic_rules ON router_versions.id = joined_updated_traffic_rules.id
    ) AS t2
WHERE t.id = t2.id;
