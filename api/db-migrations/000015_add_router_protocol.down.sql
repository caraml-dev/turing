-- remove protocol from individual routes of router_versions.routes
WITH individual_route AS (
    SELECT id, jsonb_array_elements(routes) - 'protocol' - 'service_method' as route FROM router_versions
),
updated_route AS (
    SELECT id, json_agg(route) as updated_routes from individual_route
    GROUP BY individual_route.id
)
UPDATE router_versions
SET routes = updated_routes
    FROM updated_route;

-- Remove Router protocol
ALTER TABLE router_versions DROP COLUMN protocol;

-- Remove type for Router and Route protocol
DROP TYPE IF EXISTS router_protocol;
DROP TYPE IF EXISTS route_protocol;