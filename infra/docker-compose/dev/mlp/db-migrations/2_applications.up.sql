CREATE TABLE IF NOT EXISTS applications
(
    id                 serial PRIMARY KEY,
    name               varchar(50)                      UNIQUE NOT NULL,
    href               varchar(50)                      NOT NULL,
    description        varchar(200),
    icon               varchar(50)                      NOT NULL,
    use_projects       boolean                          NOT NULL,
    is_in_beta         boolean                          NOT NULL DEFAULT FALSE,
    is_disabled        boolean                          NOT NULL DEFAULT FALSE
);

INSERT INTO applications(
    name,
    href,
    description,
    icon,
    use_projects,
    is_in_beta,
    is_disabled
) VALUES
    ('Merlin', '/merlin', 'Platform for deploying machine learning models', 'machineLearningApp', TRUE, FALSE, FALSE),
    ('Clockwork', '/clockwork', 'Declarative scheduler built on top of Apache Airflow', 'upgradeAssistantApp', FALSE, FALSE, FALSE),
    ('Excalibur', '/excalibur', 'Platform for Stream-to-Stream ML Inference', 'pipelineApp', TRUE, TRUE, FALSE)
;