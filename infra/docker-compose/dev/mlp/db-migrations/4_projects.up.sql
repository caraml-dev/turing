CREATE TABLE IF NOT EXISTS projects
(
    id                  serial PRIMARY KEY,
    name                varchar(50)     NOT NULL,
    mlflow_tracking_url varchar(100)    NOT NULL,
    k8s_cluster_name    varchar(100)    NOT NULL default '',
    administrators      varchar(256)[],
    readers             varchar(256)[],
    team                varchar(64),
    stream              varchar(64),
    labels              jsonb,
    created_at          timestamp       NOT NULL default current_timestamp,
    updated_at          timestamp       NOT NULL default current_timestamp,
    UNIQUE (name)
);
