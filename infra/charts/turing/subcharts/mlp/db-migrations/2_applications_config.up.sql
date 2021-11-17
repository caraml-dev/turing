INSERT INTO applications(name,
                         href,
                         config,
                         description,
                         icon,
                         use_projects,
                         is_in_beta,
                         is_disabled)
VALUES ('Turing',
        '/turing',
        '{
            "sections": [
                {
                    "name": "Routers",
                    "href": "/routers"
                },
                {
                    "name": "Ensemblers",
                    "href": "/ensemblers"
                },
                {
                    "name": "Ensembling Jobs",
                    "href": "/jobs"
                }
            ]
        }',
        'Platform for setting up ML experiments',
        'graphApp',
        TRUE, FALSE, FALSE);