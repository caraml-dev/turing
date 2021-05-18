ALTER TABLE ensemblers RENAME TO ensembler_configs;

CREATE TABLE IF NOT EXISTS ensemblers
(
    id         serial PRIMARY KEY,
    project_id integer     NOT NULL,
    name       varchar(50) ,
    type       varchar(20) ,

    created_at timestamp   NOT NULL default current_timestamp,
    updated_at timestamp   NOT NULL default current_timestamp,

    CONSTRAINT uc_ensemblers_project_name UNIQUE (project_id, name)
);

CREATE TABLE IF NOT EXISTS pyfunc_ensemblers
(
    mlflow_experiment_id integer      NOT NULL,
    mlflow_run_id        varchar(32)  NOT NULL,
    artifact_uri         varchar(128) NOT NULL,

    CONSTRAINT uc_pyfunc_ensemblers_project_name UNIQUE (project_id, name)
) inherits (ensemblers);
