CREATE TABLE IF NOT EXISTS users
(
    id                 serial PRIMARY KEY,
    username           varchar(50)                      NOT NULL,
    email              varchar(50)                      NOT NULL,
    created_at         timestamp                        NOT NULL default current_timestamp,
    updated_at         timestamp                        NOT NULL default current_timestamp,
    UNIQUE (id, email)
);

CREATE TABLE IF NOT EXISTS accounts
(
    id                  serial PRIMARY KEY,
    user_id             integer REFERENCES users (id)   NOT NULL,
    account_type        varchar(100)                    NOT NULL,                          
    access_token        varchar(100)                    NOT NULL,
    token_type          varchar(100)                    NOT NULL,
    refresh_token       varchar(100)                    NOT NULL,
    created_at          timestamp                       NOT NULL default current_timestamp,
    updated_at          timestamp                       NOT NULL default current_timestamp
);
