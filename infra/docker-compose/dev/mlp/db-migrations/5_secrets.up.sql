CREATE TABLE IF NOT EXISTS secrets
(
    id          serial PRIMARY KEY,
    project_id  integer REFERENCES projects(id) NOT NULL,
    name        varchar(100)  NOT NULL,
    data        text          NOT NULL,
    created_at  timestamp     NOT NULL default current_timestamp,
    updated_at  timestamp     NOT NULL default current_timestamp,
    UNIQUE (project_id, name)
);
