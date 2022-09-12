ALTER TABLE router_versions ADD default_traffic_rule jsonb;

WITH cleaned_router_versions AS (
    SELECT id, traffic_rules
    FROM router_versions
    WHERE traffic_rules IS NOT NULL AND traffic_rules != 'null'
),
rule_tbl AS (
    SELECT id, jsonb_array_elements(traffic_rules) AS rule FROM cleaned_router_versions
),
-- explode the route names from each rule
rule_route_tbl AS (
    SELECT id, jsonb_array_elements(rule->'routes') AS routes FROM rule_tbl
),
-- all routes for each router version
all_routes AS (
    SELECT
        id, jsonb_agg(DISTINCT routes) all_routes
    FROM rule_route_tbl
    GROUP BY id
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
exploded_traffic_rules_with_dangling_routes AS (
    SELECT
        exploded_traffic_rules.id
        , position
        , CASE
            WHEN elem -> 'routes' IS NOT NULL THEN jsonb_build_object(
                'conditions', elem -> 'conditions',
                'routes', CASE 
                    WHEN dangling_routes IS NOT NULL THEN (elem -> 'routes' || dangling_routes)
                ELSE (elem -> 'routes')
                END,
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
SET default_traffic_rule = 
    CASE
        WHEN dangling_routes IS NOT NULL THEN jsonb_build_object('routes', dangling_routes)
    ELSE jsonb_build_object('routes', all_routes) END,
traffic_rules = traffic_rules_updated
FROM 
    (
        SELECT
            joined_updated_traffic_rules.id,
            dangling_routes,
            all_routes,
            traffic_rules_updated
        FROM
            joined_updated_traffic_rules
        LEFT JOIN dangling_routes ON joined_updated_traffic_rules.id = dangling_routes.id
        LEFT JOIN all_routes ON joined_updated_traffic_rules.id = all_routes.id
    ) AS t2
WHERE t.id = t2.id;
