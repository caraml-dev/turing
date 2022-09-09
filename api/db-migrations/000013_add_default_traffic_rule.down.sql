WITH dangling_routes AS (
    SELECT
        id
        , default_traffic_rule -> 'routes' AS routes
    FROM
        router_versions
    WHERE default_traffic_rule -> 'routes' IS NOT NULL
),
cleaned_router_versions AS (
    SELECT id, CASE WHEN traffic_rules = 'null' THEN '[]'::jsonb ELSE traffic_rules END AS traffic_rules
    FROM router_versions
),
exploded_traffic_rules AS (
    SELECT
        id
        , position
        , elem
    FROM cleaned_router_versions, jsonb_array_elements(traffic_rules) WITH ordinality arr(elem, position)
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
