UPDATE router_versions
SET traffic_rules = combined_rule FROM (
    SELECT
        id
        , json_agg(rule_with_name) AS combined_rule
    FROM
        (
            SELECT
                t1.id
                , t1.position
                , t1.name
                , t2.elem
                , jsonb_build_object(
                    'conditions', t2.elem -> 'conditions',
                    'routes', t2.elem -> 'routes',
                    'name', t1.name
                ) AS rule_with_name
            FROM
            (
                (
                    SELECT
                        id,
                        position,
                        CONCAT('rule_', position::text) as name
                    FROM
                        router_versions,
                        jsonb_array_elements(traffic_rules) with ordinality arr(elem, position)
                ) AS t1
                LEFT JOIN
                (
                    SELECT
                        rv.id,
                        position,
                        elem
                    FROM router_versions rv, jsonb_array_elements(rv.traffic_rules) with ordinality arr(elem, position)
                ) AS t2
                ON t1.id = t2.id AND t1.position = t2.position
            )
        ) AS t3
    GROUP BY id
) AS t4
WHERE router_versions.id = t4.id;
