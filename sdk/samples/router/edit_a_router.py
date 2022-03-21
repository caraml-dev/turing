import turing
import turing.batch
import turing.batch.config
import turing.router.config.router_config
from turing.router.config.common.env_var import EnvVar


def main(turing_api: str, project: str):
    # Initialize Turing client
    turing.set_url(turing_api)
    turing.set_project(project)

    # Imagine we only have a router's id, and would like to retrieve it
    router_1 = turing.Router.get(1)

    # Now we'd like to update the config of router_1, but with some fields modified
    # Reminder: When trying to replicate configuration from an existing router, always retrieve the underlying
    # `RouterConfig` from the `Router` instance by accessing its `config` attribute.

    # Get the router config from router_1
    new_router_config_to_deploy = router_1.config

    # Make your desired changes to the config
    # Note that router_config.enricher.env is a regular Python list; so you can use methods such as append or extend
    new_router_config_to_deploy.enricher.env.append(
        EnvVar(
            name="WORKERS",
            value="2"
        )
    )

    new_router_config_to_deploy.resource_request.max_replica = 5

    # When editing a router, you can either 1. UPDATE the router, which would create a new router version and deploy it
    # immediately, or 2. SAVE the router version, which would only create a new router version without deploying it

    # 1. When you UPDATE a router, Turing will save the new version and attempt to deploy it immediately
    router_1.update(new_router_config_to_deploy)

    # Notice that the latest router version is pending deployment while the current router version is still active
    versions = router_1.list_versions()
    for v in versions:
        print(v)

    # 2. When you SAVE a router, Turing will save the new version, but not deploy it.
    new_router_config_to_save = router_1.config
    new_router_config_to_save.resource_request.min_replica = 0

    router_1.save_version(new_router_config_to_deploy)

    # Notice that the latest router version is undeployed (Turing has created the new version without deploying it)
    versions = router_1.list_versions()
    for v in versions:
        print(v)


if __name__ == '__main__':
    import fire
    fire.Fire(main)
