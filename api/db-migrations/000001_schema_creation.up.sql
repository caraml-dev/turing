CREATE TYPE router_status as ENUM ('pending', 'failed', 'deployed', 'undeployed');
CREATE TYPE router_version_status as ENUM ('pending', 'failed', 'deployed', 'undeployed');

CREATE TABLE IF NOT EXISTS routers
(
    id                     serial        PRIMARY KEY,
    project_id             integer       NOT NULL,
    environment_name       varchar(50)   NOT NULL,
    name                   varchar(50)   NOT NULL,
    status                 router_status NOT NULL default 'pending',
    endpoint               varchar(128),

    curr_router_version_id integer,

    created_at             timestamp     NOT NULL default current_timestamp,
    updated_at             timestamp     NOT NULL default current_timestamp,

    CONSTRAINT uc_router_project_environment_name UNIQUE (project_id, name)
);

CREATE TABLE IF NOT EXISTS enrichers
(
    id               serial PRIMARY KEY,
    image            varchar(128) NOT NULL,
    resource_request jsonb        NOT NULL,
    endpoint         varchar(128) NOT NULL,
    timeout          varchar(20)  NOT NULL,
    port             integer      NOT NULL,
    env              jsonb,
    created_at       timestamp    NOT NULL default current_timestamp,
    updated_at       timestamp    NOT NULL default current_timestamp
);

CREATE TABLE IF NOT EXISTS ensemblers
(
    id               serial PRIMARY KEY,
    image            varchar(128) NOT NULL,
    resource_request jsonb        NOT NULL,
    endpoint         varchar(128) NOT NULL,
    timeout          varchar(20)  NOT NULL,
    port             integer      NOT NULL,
    env              jsonb,
    created_at       timestamp    NOT NULL default current_timestamp,
    updated_at       timestamp    NOT NULL default current_timestamp
);

CREATE TABLE IF NOT EXISTS router_versions
(
    id                   serial PRIMARY KEY,
    router_id            integer references routers (id)    NOT NULL,
    version              integer                            NOT NULL,
    status               router_version_status              NOT NULL default 'pending',
    image                varchar(128)                       NOT NULL,
    routes               jsonb                              NOT NULL,
    default_route_id     varchar(40)                        NOT NULL,
    experiment_engine    jsonb                              NOT NULL,
    resource_request     jsonb                              NOT NULL,
    timeout              varchar(20)                        NOT NULL,
    log_config           jsonb                              NOT NULL,
    enricher_id          integer references enrichers (id),
    ensembler_id         integer references ensemblers (id),
    error                text,
    created_at           timestamp                          NOT NULL default current_timestamp,
    updated_at           timestamp                          NOT NULL default current_timestamp
);

ALTER TABLE routers
    ADD CONSTRAINT fk_curr_deployed_version FOREIGN KEY (curr_router_version_id) REFERENCES router_versions (id) ON DELETE CASCADE;
