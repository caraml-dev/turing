-- Add Router Version deployment start time
ALTER TABLE router_versions
ADD deployment_start_time timestamp;
