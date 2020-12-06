INSERT INTO applications
(
    name,
    href,
    description,
    icon,
    use_projects,
    is_in_beta,
    is_disabled
)
VALUES
    ('Turing', '/turing', 'Platform for setting up ML experiments', 'graphApp', TRUE, TRUE, FALSE);

UPDATE applications SET is_disabled = TRUE WHERE name = 'Clockwork';
