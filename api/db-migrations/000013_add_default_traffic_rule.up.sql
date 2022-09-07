ALTER TABLE router_versions ADD default_traffic_rule jsonb;

WITH cleaned_router_versions AS (
    SELECT id, CASE WHEN traffic_rules = 'null' THEN '[]'::jsonb ELSE traffic_rules END AS traffic_rules
    FROM router_versions
),
tbl1 AS (
    SELECT id, CASE WHEN r.routes IS NOT NULL THEN r.routes ELSE '[]'::jsonb END AS routes
    FROM cleaned_router_versions
    LEFT JOIN lateral jsonb_to_recordset(traffic_rules) as r("routes" jsonb) ON true
),
tbl2 AS (
    SELECT id, jsonb_agg(DISTINCT elems) AS routes
    FROM tbl1, jsonb_array_elements(routes) as elems
    GROUP BY tbl1.id
),
tbl3 AS (
    SELECT tbl1.id, CASE WHEN tbl2.routes IS NOT NULL THEN tbl2.routes ELSE '[]'::jsonb END AS routes
    FROM tbl1 LEFT JOIN tbl2 ON tbl1.id = tbl2.id
),
tbl4 AS (
    SELECT id, jsonb_agg(DISTINCT routes) AS routes
    FROM (
        SELECT
            rv.id
            , r.id as routes
        FROM router_versions rv
        JOIN lateral jsonb_to_recordset(routes) as r("id" jsonb) ON true
    ) AS t_explode
    GROUP BY t_explode.id
),
-- DISTINCT is required for de-duplication
dangling_routes AS (
    SELECT
        DISTINCT tbl3.id
        , array_to_json(ARRAY(SELECT jsonb_array_elements_text(tbl4.routes) EXCEPT
                SELECT jsonb_array_elements_text(tbl3.routes)))::jsonb AS dangling_routes
    FROM
        tbl3 LEFT JOIN tbl4 ON tbl3.id = tbl4.id
),
-- tbl 5,6,7 are for traffic_rules column update
tbl5 AS (
    SELECT
        id
        , position
        , elem
    FROM cleaned_router_versions, jsonb_array_elements(traffic_rules) WITH ordinality arr(elem, position)
),
tbl6 AS (
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
        LEFT JOIN tbl5 ON cleaned_router_versions.id = tbl5.id
        LEFT JOIN dangling_routes ON cleaned_router_versions.id = dangling_routes.id
),
-- FILTER clause added to prevent aggregating null values into [null] jsonb
tbl7 AS (
    SELECT
        id,
        json_agg(traffic_rule ORDER BY position) FILTER (where traffic_rule IS NOT NULL) traffic_rules_updated
    FROM tbl6 GROUP BY id
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
        LEFT JOIN tbl7 ON router_versions.id = tbl7.id
    ) AS t2
WHERE t.id = t2.id;
