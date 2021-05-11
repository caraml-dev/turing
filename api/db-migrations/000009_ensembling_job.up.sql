CREATE TYPE ensembling_job_status as ENUM (
    'pending',
    'running',
    'terminating',
    'completed',
    'failed',
    'terminated',
    'failed_submission',
    'failed_building'
);

CREATE TABLE IF NOT EXISTS ensembling_jobs
(
    id               serial PRIMARY KEY,
    name             varchar(50) NOT NULL,
    version_id       integer,
    ensembler_id     integer NOT NULL,
    project_id       integer NOT NULL,
    environment_name varchar(50) NOT NULL,
    infra_config     jsonb NOT NULL,
    ensembler_config jsonb NOT NULL,
    status           ensembling_job_status NOT NULL default 'pending',
    error            text,
    created_at       timestamp   NOT NULL default current_timestamp,
    updated_at       timestamp   NOT NULL default current_timestamp
);
