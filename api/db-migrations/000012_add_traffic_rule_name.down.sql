UPDATE router_versions
SET traffic_rules = combined_rule FROM (
    SELECT
        t1.id
        , json_agg(t2.unnamed_rule) AS combined_rule
    FROM
    (
        (
            SELECT
                id,
                position
            FROM
                router_versions,
                jsonb_array_elements(traffic_rules) with ordinality arr(elem, position)
        ) AS t1
        LEFT JOIN
        (
        SELECT
            rv.id,
            position,
            elem - 'name' AS unnamed_rule
        FROM 
            router_versions rv,
            jsonb_array_elements(rv.traffic_rules) with ordinality arr(elem, position)
        ) AS t2
        ON t1.id = t2.id AND t1.position = t2.position
    )
    GROUP BY t1.id
) AS t3
WHERE router_versions.id = t3.id;
