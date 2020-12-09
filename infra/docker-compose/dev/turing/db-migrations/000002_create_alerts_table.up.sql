CREATE TABLE IF NOT EXISTS alerts
(
    id                 serial PRIMARY KEY,

    environment        varchar(255),
    team               varchar(255),
    service            varchar(255),

    metric             varchar(255),
    warning_threshold  double precision,
    critical_threshold double precision,
    duration           varchar(255),

    created_at         timestamp NOT NULL default current_timestamp,
    updated_at         timestamp NOT NULL default current_timestamp
);

-- Index by the columns that are commonly used to filter alerts.
CREATE INDEX alerts_idx ON alerts (service);

-- An alert is also uniquely identified by environment, team, service, and metric.
CREATE UNIQUE INDEX alerts_idx_unique ON alerts (environment, team, service, metric);