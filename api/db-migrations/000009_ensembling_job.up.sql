CREATE TYPE ensembling_job_status as ENUM (
    'pending',
    'running',
    'completed',
    'failed',
    'terminated',
    'failed_submission',
    'failed_building',
    'building'
);

CREATE TABLE IF NOT EXISTS ensembling_jobs
(
    id               serial PRIMARY KEY,
    name             text NOT NULL,
    ensembler_id     integer NOT NULL,
    project_id       integer NOT NULL,
    environment_name varchar(50) NOT NULL,
    infra_config     jsonb NOT NULL,
    job_config       jsonb NOT NULL,
    is_locked        bool NOT NULL default 'false', -- This is set true when the API picks the record up for processing
    retry_count      integer NOT NULL default 0,
    status           ensembling_job_status NOT NULL default 'pending',
    error            text,
    created_at       timestamp   NOT NULL default current_timestamp,
    updated_at       timestamp   NOT NULL default current_timestamp
);

CREATE INDEX ensembling_jobs_status_updated_at_idx on ensembling_jobs (status, updated_at);

COMMENT ON COLUMN ensembling_jobs.is_locked IS 'This is set to true when the API picks the record up for processing';
