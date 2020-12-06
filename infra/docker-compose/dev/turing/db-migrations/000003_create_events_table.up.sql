CREATE TABLE IF NOT EXISTS events
(
    id                 serial PRIMARY KEY,

    router_id          integer      NOT NULL references routers (id) ON DELETE CASCADE,
    version            integer      NOT NULL,
    event_type         varchar(25)  NOT NULL,
    stage              varchar(128) NOT NULL,
    message            text         NOT NULL,

    created_at         timestamp NOT NULL default current_timestamp,
    updated_at         timestamp NOT NULL default current_timestamp
);

-- Index by router id.
CREATE INDEX events_idx ON events (router_id);
