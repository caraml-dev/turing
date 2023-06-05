import deprecation
import inspect
from typing import List, Dict, Union
from collections import Counter
from dataclasses import dataclass
from enum import Enum

import turing.generated.models
from turing.generated.model_utils import OpenApiModel
from turing.router.config.traffic_rule import DefaultTrafficRule
from turing.router.config.route import Route
from turing.router.config.traffic_rule import TrafficRule
from turing.router.config.autoscaling_policy import (
    AutoscalingPolicy,
    DEFAULT_AUTOSCALING_POLICY,
)
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.log_config import LogConfig, ResultLoggerType
from turing.router.config.enricher import Enricher
from turing.router.config.router_ensembler_config import (
    DockerRouterEnsemblerConfig,
    NopRouterEnsemblerConfig,
    PyfuncRouterEnsemblerConfig,
    RouterEnsemblerConfig,
    StandardRouterEnsemblerConfig,
)
from turing.router.config.experiment_config import ExperimentConfig


NAME_INDEX = 0
VALUE_INDEX = 1


class Protocol(Enum):
    """
    Router Protocol type
    """

    UPI = "UPI_V1"
    HTTP = "HTTP_JSON"

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.Protocol(self.value)


@dataclass
class RouterConfig:
    """
    Class to create a new RouterConfig. Can be built up from its individual components or initialised instantly
    from an appropriate API response

    :param environment_name: name of the environment
    :param name: name of the router
    :param routes: list of routes used by the router
    :param rules: list of rules used by the router
    :param default_route_id: default route id to be used
    :param default_traffic_rule: default traffic rule to be used if no conditions are matched
    :param experiment_engine: experiment engine config file
    :param resource_request: resources to be provisioned for the router
    :param timeout: request timeout which when exceeded, the request to the router will be terminated
    :param log_config: logging config settings to be used with the router
    :param enricher: enricher config settings to be used with the router
    :param ensembler: ensembler config settings to be used with the router
    """

    environment_name: str = ""
    name: str = ""
    routes: Union[List[Route], List[Dict[str, str]]] = None
    rules: Union[List[TrafficRule], List[Dict]] = None
    default_route_id: str = None
    default_traffic_rule: DefaultTrafficRule = None
    experiment_engine: Union[ExperimentConfig, Dict] = None
    resource_request: Union[ResourceRequest, Dict[str, Union[str, int]]] = None
    autoscaling_policy: Union[AutoscalingPolicy, Dict[str, str]] = None
    timeout: str = None
    protocol: Protocol = None
    log_config: Union[LogConfig, Dict[str, Union[str, bool, int]]] = None
    enricher: Union[Enricher, Dict] = None
    ensembler: Union[RouterEnsemblerConfig, Dict] = None

    def __init__(
        self,
        environment_name: str = "",
        name: str = "",
        routes: Union[List[Route], List[Dict[str, str]]] = None,
        rules: Union[List[TrafficRule], List[Dict]] = None,
        default_route_id: str = None,
        default_traffic_rule: DefaultTrafficRule = None,
        experiment_engine: Union[ExperimentConfig, Dict] = None,
        resource_request: Union[ResourceRequest, Dict[str, Union[str, int]]] = None,
        autoscaling_policy: Union[
            AutoscalingPolicy, Dict[str, str]
        ] = DEFAULT_AUTOSCALING_POLICY,
        timeout: str = None,
        protocol: Union[Protocol, str] = Protocol.HTTP,
        log_config: Union[LogConfig, Dict[str, Union[str, bool, int]]] = LogConfig(
            result_logger_type=ResultLoggerType.NOP
        ),
        enricher: Union[Enricher, Dict] = None,
        ensembler: Union[RouterEnsemblerConfig, Dict] = None,
        **kwargs,
    ):
        self.environment_name = environment_name
        self.name = name
        self.protocol = protocol
        self.routes = routes
        self.rules = rules
        self.default_route_id = default_route_id
        self.default_traffic_rule = default_traffic_rule
        self.experiment_engine = experiment_engine
        self.resource_request = resource_request
        self.autoscaling_policy = autoscaling_policy
        self.timeout = timeout
        self.log_config = log_config
        self.enricher = enricher
        # Init ensembler after the default route has been initialized
        self.ensembler = ensembler

    @property
    def environment_name(self) -> str:
        return self._environment_name

    @environment_name.setter
    def environment_name(self, environment_name: str):
        self._environment_name = environment_name

    @property
    def name(self) -> str:
        return self._name

    @name.setter
    def name(self, name: str):
        self._name = name

    @property
    def routes(self) -> List[Route]:
        return self._routes

    @routes.setter
    def routes(self, routes: Union[List[Route], List[Dict[str, str]]]):
        if isinstance(routes, list) and all(
            isinstance(route, dict) for route in routes
        ):
            routes = [Route(**route) for route in routes]
        for route in routes:
            if self._protocol == Protocol.HTTP:
                Route._verify_endpoint(route.endpoint)
            elif self._protocol == Protocol.UPI:
                Route._verify_service_method(route.service_method)
        self._routes = routes

    @property
    def rules(self) -> List[TrafficRule]:
        return self._rules

    @rules.setter
    def rules(self, rules: Union[List[TrafficRule], List[Dict]]):
        if isinstance(rules, list):
            if all(isinstance(rule, TrafficRule) for rule in rules):
                self._rules = rules
            elif all(isinstance(rule, dict) for rule in rules):
                self._rules = [TrafficRule(**rule) for rule in rules]
            else:
                self._rules = rules
        else:
            self._rules = rules

    @property
    @deprecation.deprecated(
        deprecated_in="0.4.0",
        details="Please use the ensembler properties to configure the final / fallback route.",
    )
    def default_route_id(self) -> str:
        return self._default_route_id

    @default_route_id.setter
    @deprecation.deprecated(
        deprecated_in="0.4.0",
        details="Please use the ensembler properties to configure the final / fallback route.",
    )
    def default_route_id(self, default_route_id: str):
        self._default_route_id = default_route_id
        # User may directly modify the default_route_id property while it is deprecated.
        # So, copy to the nop / standard ensembler if set.
        if hasattr(self, "ensembler"):
            if isinstance(self.ensembler, NopRouterEnsemblerConfig):
                self.ensembler.final_response_route_id = default_route_id
            elif isinstance(self.ensembler, StandardRouterEnsemblerConfig):
                self.ensembler.fallback_response_route_id = default_route_id

    @property
    def default_traffic_rule(self) -> DefaultTrafficRule:
        return self._default_traffic_rule

    @default_traffic_rule.setter
    def default_traffic_rule(self, rule: Union[DefaultTrafficRule, Dict]):
        if isinstance(rule, DefaultTrafficRule):
            self._default_traffic_rule = rule
        elif isinstance(rule, dict):
            self._default_traffic_rule = DefaultTrafficRule(**rule)
        else:
            self._default_traffic_rule = rule

    @property
    def experiment_engine(self) -> ExperimentConfig:
        return self._experiment_engine

    @experiment_engine.setter
    def experiment_engine(self, experiment_engine: Union[ExperimentConfig, Dict]):
        if isinstance(experiment_engine, ExperimentConfig):
            self._experiment_engine = experiment_engine
        elif isinstance(experiment_engine, dict):
            self._experiment_engine = ExperimentConfig(**experiment_engine)
        else:
            self._experiment_engine = experiment_engine

    @property
    def resource_request(self) -> ResourceRequest:
        return self._resource_request

    @resource_request.setter
    def resource_request(
        self, resource_request: Union[ResourceRequest, Dict[str, Union[str, int]]]
    ):
        if isinstance(resource_request, ResourceRequest):
            self._resource_request = resource_request
        elif isinstance(resource_request, dict):
            self._resource_request = ResourceRequest(**resource_request)
        else:
            self._resource_request = resource_request

    @property
    def autoscaling_policy(self) -> AutoscalingPolicy:
        return self._autoscaling_policy

    @autoscaling_policy.setter
    def autoscaling_policy(
        self, autoscaling_policy: Union[AutoscalingPolicy, Dict[str, str]]
    ):
        if isinstance(autoscaling_policy, AutoscalingPolicy):
            self._autoscaling_policy = autoscaling_policy
        elif isinstance(autoscaling_policy, dict):
            self._autoscaling_policy = AutoscalingPolicy(**autoscaling_policy)
        else:
            self._autoscaling_policy = autoscaling_policy

    @property
    def timeout(self) -> str:
        return self._timeout

    @timeout.setter
    def timeout(self, timeout: str):
        self._timeout = timeout

    @property
    def protocol(self) -> Protocol:
        return self._protocol

    @protocol.setter
    def protocol(self, protocol: Union[Protocol, str]):
        if isinstance(protocol, str):
            self._protocol = Protocol(protocol)
        else:
            self._protocol = protocol

    @property
    def log_config(self) -> LogConfig:
        return self._log_config

    @log_config.setter
    def log_config(
        self, log_config: Union[LogConfig, Dict[str, Union[str, bool, int]]]
    ):
        if isinstance(log_config, LogConfig):
            self._log_config = log_config
        elif isinstance(log_config, dict):
            self._log_config = LogConfig(**log_config)
        else:
            self._log_config = log_config

    @property
    def enricher(self) -> Enricher:
        return self._enricher

    @enricher.setter
    def enricher(self, enricher: Union[Enricher, Dict]):
        if isinstance(enricher, Enricher):
            self._enricher = enricher
        elif isinstance(enricher, dict):
            self._enricher = Enricher(**enricher)
        else:
            self._enricher = enricher

    @property
    def ensembler(self) -> RouterEnsemblerConfig:
        return self._ensembler

    @ensembler.setter
    def ensembler(self, ensembler: Union[RouterEnsemblerConfig, Dict]):
        if ensembler is None:
            # Init nop ensembler config if ensembler is not set
            self._ensembler = NopRouterEnsemblerConfig(
                final_response_route_id=self.default_route_id
            )
        elif isinstance(ensembler, dict):
            # Set fallback_response_route_id into standard ensembler config
            if (
                ensembler["type"] == "standard"
                and "fallback_response_route_id" not in ensembler["standard_config"]
            ):
                ensembler["standard_config"][
                    "fallback_response_route_id"
                ] = self.default_route_id
            self._ensembler = RouterEnsemblerConfig(**ensembler)
        else:
            self._ensembler = ensembler
        # Init child class types
        if isinstance(self._ensembler, RouterEnsemblerConfig):
            if self._ensembler.type == "nop" and not isinstance(
                self._ensembler, NopRouterEnsemblerConfig
            ):
                self._ensembler = NopRouterEnsemblerConfig.from_config(
                    self._ensembler.nop_config
                )
            elif self._ensembler.type == "standard" and not isinstance(
                self._ensembler, StandardRouterEnsemblerConfig
            ):
                self._ensembler = StandardRouterEnsemblerConfig.from_config(
                    self._ensembler.standard_config
                )
            elif self._ensembler.type == "docker" and not isinstance(
                self._ensembler, DockerRouterEnsemblerConfig
            ):
                self._ensembler = DockerRouterEnsemblerConfig.from_config(
                    self._ensembler.docker_config
                )
            elif self._ensembler.type == "pyfunc" and not isinstance(
                self._ensembler, PyfuncRouterEnsemblerConfig
            ):
                self._ensembler = PyfuncRouterEnsemblerConfig.from_config(
                    self._ensembler.pyfunc_config
                )
        # Verify that only nop or standard ensembler is configured for UPI router
        if self._protocol == Protocol.UPI:
            self._verify_upi_router_ensembler()

    def to_open_api(self) -> OpenApiModel:
        kwargs = {}
        self._verify_no_duplicate_routes()

        # Get default route id if exists
        default_route_id = self._get_default_route_id()
        if default_route_id is not None:
            kwargs["default_route_id"] = default_route_id

        if self.default_traffic_rule is not None:
            kwargs["default_traffic_rule"] = self.default_traffic_rule.to_open_api()
        if self.rules is not None:
            kwargs["rules"] = [rule.to_open_api() for rule in self.rules]
        if self.resource_request is not None:
            kwargs["resource_request"] = self.resource_request.to_open_api()
        if self.enricher is not None:
            kwargs["enricher"] = self.enricher.to_open_api()
        if self.ensembler is not None:
            kwargs["ensembler"] = self.ensembler.to_open_api()
            if kwargs["ensembler"] is None:
                # The Turing API does not handle an ensembler type "nop" - it must be left unset.
                del kwargs["ensembler"]

        return turing.generated.models.RouterConfig(
            environment_name=self.environment_name,
            name=self.name,
            config=turing.generated.models.RouterVersionConfig(
                routes=[route.to_open_api() for route in self.routes],
                autoscaling_policy=self.autoscaling_policy.to_open_api(),
                experiment_engine=self.experiment_engine.to_open_api(),
                timeout=self.timeout,
                log_config=self.log_config.to_open_api(),
                protocol=self.protocol.to_open_api(),
                **kwargs,
            ),
        )

    def _get_default_route_id(self):
        default_route_id = None
        # If nop config is set, use the final_response_route_id as the default
        if isinstance(self.ensembler, NopRouterEnsemblerConfig):
            default_route_id = self.ensembler.final_response_route_id
        # Or, if standard config is set, use the fallback_response_route_id as the default
        elif isinstance(self.ensembler, StandardRouterEnsemblerConfig):
            default_route_id = self.ensembler.fallback_response_route_id
        if default_route_id is not None:
            self._verify_default_route_exists(default_route_id)
        return default_route_id

    def _verify_default_route_exists(self, default_route_id: str):
        for route in self.routes:
            if route.id == default_route_id:
                return
        raise turing.router.config.route.InvalidRouteException(
            f"Default route id {default_route_id} is not registered in the routes."
        )

    def _verify_no_duplicate_routes(self):
        route_id_counter = Counter(route.id for route in self.routes)
        most_common_route_id, max_frequency = route_id_counter.most_common(n=1)[0]
        if max_frequency > 1:
            raise turing.router.config.route.DuplicateRouteException(
                f"Routes with duplicate ids are specified for this traffic rule. Duplicate id: {most_common_route_id}"
            )

    def _verify_upi_router_ensembler(self):
        # only nop and standard ensembler is allowed
        if not (
            isinstance(self.ensembler, NopRouterEnsemblerConfig)
            or isinstance(self.ensembler, StandardRouterEnsemblerConfig)
        ):
            raise turing.router.config.router_ensembler_config.InvalidEnsemblerTypeException(
                f"UPI router only supports no ensembler or standard ensembler."
            )

    def to_dict(self):
        att_dict = {}
        for m in inspect.getmembers(self):
            if not inspect.ismethod(m[VALUE_INDEX]) and not m[NAME_INDEX].startswith(
                "_"
            ):
                att_dict[m[NAME_INDEX]] = m[VALUE_INDEX]
        return att_dict
