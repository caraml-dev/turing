CREATE TABLE IF NOT EXISTS users
(
    id         serial PRIMARY KEY,
    username   varchar(50) NOT NULL,
    email      varchar(50) NOT NULL,
    created_at timestamp   NOT NULL default current_timestamp,
    updated_at timestamp   NOT NULL default current_timestamp,
    UNIQUE (id, email)
);

CREATE TABLE IF NOT EXISTS accounts
(
    id            serial PRIMARY KEY,
    user_id       integer REFERENCES users (id) NOT NULL,
    account_type  varchar(100)                  NOT NULL,
    access_token  varchar(100)                  NOT NULL,
    token_type    varchar(100)                  NOT NULL,
    refresh_token varchar(100)                  NOT NULL,
    created_at    timestamp                     NOT NULL default current_timestamp,
    updated_at    timestamp                     NOT NULL default current_timestamp
);

CREATE TABLE IF NOT EXISTS applications
(
    id           serial PRIMARY KEY,
    name         varchar(50) UNIQUE NOT NULL,
    href         varchar(50)        NOT NULL,
    config       jsonb,
    description  varchar(200),
    icon         varchar(50)        NOT NULL,
    use_projects boolean            NOT NULL,
    is_in_beta   boolean            NOT NULL DEFAULT FALSE,
    is_disabled  boolean            NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS projects
(
    id                  serial PRIMARY KEY,
    name                varchar(50)  NOT NULL,
    mlflow_tracking_url varchar(100) NOT NULL,
    k8s_cluster_name    varchar(100) NOT NULL default '',
    administrators      varchar(256)[],
    readers             varchar(256)[],
    team                varchar(64),
    stream              varchar(64),
    labels              jsonb,
    created_at          timestamp    NOT NULL default current_timestamp,
    updated_at          timestamp    NOT NULL default current_timestamp,
    UNIQUE (name)
);

CREATE TABLE IF NOT EXISTS secrets
(
    id         serial PRIMARY KEY,
    project_id integer REFERENCES projects (id) NOT NULL,
    name       varchar(100)                     NOT NULL,
    data       text                             NOT NULL,
    created_at timestamp                        NOT NULL default current_timestamp,
    updated_at timestamp                        NOT NULL default current_timestamp,
    UNIQUE (project_id, name)
);
