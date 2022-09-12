WITH default_traffic_rule_routes AS (
    SELECT id , default_traffic_rule -> 'routes' AS routes FROM router_versions
),
route_tbl AS (
    SELECT id, jsonb_array_elements(routes) AS route FROM router_versions
),
original_route_tbl AS (
    SELECT id, route -> 'id' AS route_id FROM route_tbl
),
-- all routes for each router version
all_routes AS (
    SELECT
        id, jsonb_agg(DISTINCT route_id) routes
    FROM original_route_tbl
    GROUP BY id
),
dangling_routes AS (
    SELECT
        id, routes
    FROM (
        SELECT * FROM default_traffic_rule_routes
        EXCEPT
        SELECT * FROM all_routes
    ) t
    -- WHERE clause filters away rows with no traffic rules
    WHERE t.routes IS NOT NULL    
),
exploded_traffic_rules AS (
    SELECT
        id
        , position
        , elem
    FROM router_versions, jsonb_array_elements(traffic_rules) WITH ordinality arr(elem, position)
    WHERE id IN (SELECT id FROM dangling_routes)
),
original_traffic_rules AS (
    SELECT
        exploded_traffic_rules.id
        , position
        , CASE
            WHEN elem IS NOT NULL THEN jsonb_build_object(
                'conditions', elem -> 'conditions',
                'routes', array_to_json(ARRAY(SELECT jsonb_array_elements_text(elem -> 'routes') EXCEPT
                    SELECT jsonb_array_elements_text(dangling_routes.routes)))::jsonb,
                'name', elem -> 'name'
            )
            ELSE NULL
            END AS traffic_rule
    FROM
        exploded_traffic_rules
        LEFT JOIN dangling_routes ON exploded_traffic_rules.id = dangling_routes.id
),
joined_original_traffic_rules AS (
    SELECT
        id,
        json_agg(traffic_rule ORDER BY position) traffic_rules_updated
    FROM original_traffic_rules GROUP BY id
)

UPDATE router_versions AS t
SET traffic_rules = traffic_rules_updated
FROM joined_original_traffic_rules
WHERE joined_original_traffic_rules.id = t.id;

ALTER TABLE router_versions DROP column default_traffic_rule;
