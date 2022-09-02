import turing
import turing.batch
import turing.batch.config
import turing.router.config.router_config
from turing.router.config.route import Route
from turing.router.config.router_config import RouterConfig
from turing.router.config.router_version import RouterStatus
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.log_config import LogConfig, ResultLoggerType
from turing.router.config.traffic_rule import (
    TrafficRule,
    HeaderTrafficRuleCondition,
    PayloadTrafficRuleCondition,
)
from turing.router.config.enricher import Enricher
from turing.router.config.router_ensembler_config import DockerRouterEnsemblerConfig
from turing.router.config.common.env_var import EnvVar
from turing.router.config.experiment_config import ExperimentConfig


def main(turing_api: str, project: str):
    # Initialize Turing client
    turing.set_url(turing_api)
    turing.set_project(project)

    # Build a router config in order to create a router
    # Note: When constructing a `RouterConfig` object from scratch, it is **highly recommended** that you construct each
    # individual component using the Turing SDK classes provided instead of using `dict` objects which do not perform
    # any schema validation.

    # Create some routes
    routes = [
        Route(id="meow", endpoint="http://fox-says.meow", timeout="20ms"),
        Route(id="woof", endpoint="http://fox-says.woof", timeout="20ms"),
        Route(id="baaa", endpoint="http://fox-says.baa", timeout="20ms"),
        Route(id="oink", endpoint="http://fox-says.oink", timeout="20ms"),
        Route(
            id="ring-ding-ding",
            endpoint="http://fox-says.ring-ding-ding",
            timeout="20ms",
        ),
        Route(id="control", endpoint="http://fox-says.control", timeout="20ms"),
    ]

    # Create some traffic rules
    # Note: Each traffic rule is defined by at least one `TrafficRuleCondition` and one route. Routes are essentially
    # the `id`s of `Route` objects that you intend to specify for the entire `TrafficRule`.
    #
    # When defining a traffic rule, one would need to decide between using a `HeaderTrafficRuleCondition` or a
    # `PayloadTrafficRuleCondition`. These subclasses can be used to build a `TrafficRuleCondition` without having to
    # manually set attributes such as `field_source` or `operator`.
    rules = [
        TrafficRule(
            name="cat-rule",
            conditions=[HeaderTrafficRuleCondition(field="name", values=["cat"])],
            routes=["meow"],
        ),
        TrafficRule(
            name="dog-rule",
            conditions=[HeaderTrafficRuleCondition(field="name", values=["dog"])],
            routes=["woof"],
        ),
        TrafficRule(
            name="sheep-rule",
            conditions=[HeaderTrafficRuleCondition(field="name", values=["sheep"])],
            routes=["baaa"],
        ),
        TrafficRule(
            name="pig-rule",
            conditions=[HeaderTrafficRuleCondition(field="name", values=["pig"])],
            routes=["oink"],
        ),
        TrafficRule(
            name="sus-rule",
            conditions=[PayloadTrafficRuleCondition(field="body", values=["sus"])],
            routes=["meow", "woof", "baaa", "oink", "ring-ding-ding"],
        ),
    ]

    # Create an experiment config
    # The `ExperimentConfig` class is a simple container to carry configuration related to an experiment to be used by a
    # Turing Router. Note that as Turing does not create experiments automatically, you would need to create your
    # experiments separately prior to specifying their configuration here.
    #
    # Also, notice that `ExperimentConfig` does not contain any fixed schema as it simply carries configuration for
    # experiment engines, which are used as plug-ins for Turing. When building an `ExperimentConfig` from scratch, you
    # would need to consider the underlying schema for the `config` attribute as well as the appropriate `type` that
    # corresponds to your selected experiment engine.
    experiment_config = ExperimentConfig(
        type="test-exp",
        config={
            "variables": [
                {"name": "farm_id", "field": "farm_id", "field_source": "header"},
                {"name": "country_code", "field": "country", "field_source": "header"},
                {"name": "latitude", "field": "farm_lat", "field_source": "header"},
                {"name": "longitude", "field": "farm_long", "field_source": "header"},
            ],
            "project_id": 102,
        },
    )

    # Create a resource request config for the router
    # Note: The units for CPU and memory requests are measured in cpu units and bytes respectively. You may wish to
    # read more about how these are measured here:
    # https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/.
    resource_request = ResourceRequest(
        min_replica=0, max_replica=2, cpu_request="500m", memory_request="512Mi"
    )

    # Create a log config for the router
    # Note: Logging for Turing Routers is done through BigQuery or Kafka, and its configuration is managed by the
    # `LogConfig` class. Two helper classes (child classes of `LogConfig`) have been created to assist you in
    # constructing these objects - `BigQueryLogConfig` and `KafkaLogConfig`.
    #
    # If you do not intend to use any logging, simply create a regular `LogConfig` object with `result_loggger_type` set
    # as `ResultLoggerType.NOP`, without defining the other arguments.
    log_config = LogConfig(result_logger_type=ResultLoggerType.NOP)

    # Create an enricher for the router
    enricher = Enricher(
        image="docker.io/ealen/echo-server:0.5.1",
        resource_request=ResourceRequest(
            min_replica=0, max_replica=2, cpu_request="500m", memory_request="512Mi"
        ),
        endpoint="/echo",
        timeout="60ms",
        port=3000,
        env=[EnvVar(name="humans", value="farmer-joe")],
    )

    # Create an ensembler for the router
    # Note: Ensembling for Turing Routers is done through Nop, Standard, Docker or Pyfunc ensemblers, and its configuration
    # is managed by the `RouterEnsemblerConfig` class. Helper classes (child classes of `RouterEnsemblerConfig`)
    # have been created to assist you in constructing these objects - `NopRouterEnsemblerConfig`,
    # `StandardRouterEnsemblerConfig`, `DockerRouterEnsemblerConfig` and `PyfuncRouterEnsemblerConfig`.
    ensembler = DockerRouterEnsemblerConfig(
        image="docker.io/ealen/echo-server:0.5.1",
        resource_request=ResourceRequest(
            min_replica=1, max_replica=3, cpu_request="500m", memory_request="512Mi"
        ),
        endpoint="/echo",
        timeout="60ms",
        port=3000,
        env=[],
    )

    # Create the RouterConfig instance
    router_config = RouterConfig(
        environment_name="id-dev",
        name="what-does-the-fox-say",
        routes=routes,
        rules=rules,
        experiment_engine=experiment_config,
        resource_request=resource_request,
        timeout="100ms",
        log_config=log_config,
        enricher=enricher,
        ensembler=ensembler,
    )

    # 1. Create a new router using the RouterConfig object
    # Note: A `Router` object represents a router that is created on Turing API. It does not (and should not) ever be
    # created manually by using its constructor directly. Instead, you should only be manipulating with `Router`
    # instances that get returned as a result of using the various `Router` class and instance methods that interact
    # with Turing API, such as the one below.
    new_router = turing.Router.create(router_config)
    print(f"1. You have created a router with id: {new_router.id}")

    # 2. List all routers
    routers = turing.Router.list()
    print(f"2. You have just retrieved a list of {len(routers)} routers:")
    for r in routers:
        if r.name == new_router.name:
            my_router = r
        print(r)

    # Wait for the router to get deployed
    try:
        my_router.wait_for_status(RouterStatus.DEPLOYED)
    except TimeoutError:
        raise Exception(
            f"Turing API is taking too long for router {my_router.id} to get deployed."
        )

    # 3. Get the router you just created using the router_id obtained
    my_router = turing.Router.get(my_router.id)
    print(f"3. You have retrieved the router with name: {my_router.name}")

    # Access the router config from the returned Router object directly
    my_router_config = my_router.config

    # Modify something in the router config
    my_router_config.routes.append(
        Route(
            id="fee-fi-fo-fum", endpoint="http://fox-says.fee-fi-fo-fum", timeout="20ms"
        )
    )

    # 4. Update the router with the new router config
    my_router.update(my_router_config)
    print(f"4. You have just updated your router with a new config.")

    # 5. List all the router config versions of your router
    # Note: A `RouterVersion` represents a single version (and configuration) of a Turing Router. Just as `Router`
    # objects, they should almost never be created manually by using their constructor.
    #
    # Besides accessing attributes of a `RouterVersion` object directly, which will allow you to access basic
    # attributes, you may also consider retrieving the entire router configuration from a specific `RouterVersion`
    # object as a `RouterConfig` for further manipulation by performing something like:
    #
    # `my_config = router_version.get_config()`
    my_router_versions = my_router.list_versions()
    print(
        f"5. You have just retrieved a list of {len(my_router_versions)} versions for your router:"
    )
    for ver in my_router_versions:
        print(ver)

    # Sort the versions returned by version number
    my_router_versions.sort(key=lambda x: x.version)
    # Get the version number of the first version returned
    first_ver_no = my_router_versions[0].version
    # Get the version number of the latest version returned
    latest_ver_no = my_router_versions[-1].version

    # Wait for the latest version to get deployed
    try:
        my_router.wait_for_version_status(RouterStatus.DEPLOYED, latest_ver_no)
    except TimeoutError:
        raise Exception(
            f"Turing API is taking too long for router {my_router.id} with version {latest_ver_no} to get "
            f"deployed."
        )

    # 6. Deploy a specific router config version (the first one we created)
    response = my_router.deploy_version(first_ver_no)
    print(
        f"6. You have deployed version {response['version']} of router {response['router_id']}."
    )

    # Wait for the first version to get deployed
    try:
        my_router.wait_for_version_status(RouterStatus.DEPLOYED, first_ver_no)
    except TimeoutError:
        raise Exception(
            f"Turing API is taking too long for router {my_router.id} with version {first_ver_no} to get "
            f"deployed."
        )

    # 7. Undeploy the current active router configuration
    response = my_router.undeploy()
    print(f"7. You have undeployed router {response['router_id']}.")

    # Wait for the router to get undeployed
    try:
        my_router.wait_for_status(RouterStatus.UNDEPLOYED)
    except TimeoutError:
        raise Exception(
            f"Turing API is taking too long for router {my_router.id} to get undeployed."
        )

    # 8. Deploy the router's *current* configuration (notice how it still deploys the *first* version)
    response = my_router.deploy()
    print(
        f"8. You have deployed version {response['version']} of router {response['router_id']}."
    )

    # Wait for the router to get deployed
    try:
        my_router.wait_for_status(RouterStatus.DEPLOYED)
    except TimeoutError:
        raise Exception(
            f"Turing API is taking too long for router {my_router.id} to get deployed."
        )

    # Undeploy the router
    response = my_router.undeploy()
    print(f"You have undeployed router {response['router_id']}.")

    # 9. Get a specific router version of the router
    my_router_ver = my_router.get_version(first_ver_no)
    print(f"9. You have just retrieved version {my_router_ver.version} of your router.")

    # 10. Delete a specific router version of the router
    response = my_router.delete_version(latest_ver_no)
    print(
        f"10. You have deleted version {response['version']} of router {response['router_id']}."
    )

    # 11. Get all deployment events associated with this router
    events = my_router.get_events()
    print(
        f"11. You have just retrieved a list of {len(events)} events for your router:"
    )
    for e in events:
        print(e)

    # 12. Delete this router (using its router_id)
    deleted_router_id = turing.Router.delete(my_router.id)
    print(f"12. You have just deleted the router with id: {deleted_router_id}")

    # Check if the router still exists
    for r in turing.Router.list():
        if r.name == my_router_config.name:
            raise Exception("Oh my, this router still exists!")


if __name__ == "__main__":
    import fire

    fire.Fire(main)
