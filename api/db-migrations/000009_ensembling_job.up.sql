CREATE TYPE ensembling_job_status as ENUM (
    'pending',
    'running',
    'completed',
    'failed',
    'terminating',
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
    retry_count      integer NOT NULL default 0,
    run_id           integer NOT NULL,
    status           ensembling_job_status NOT NULL default 'pending',
    error            text,
    created_at       timestamp   NOT NULL default current_timestamp,
    updated_at       timestamp   NOT NULL default current_timestamp
);

CREATE INDEX ensembling_jobs_status_updated_at_idx on ensembling_jobs (status, updated_at);
CREATE INDEX ensembling_job_ensembler_id_idx on ensembling_jobs (ensembler_id);
