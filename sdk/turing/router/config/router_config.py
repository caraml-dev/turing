import inspect
from typing import List, Dict, Union
from collections import Counter
from dataclasses import dataclass

import turing.generated.models
from turing.generated.model_utils import OpenApiModel
from turing.router.config.route import Route
from turing.router.config.traffic_rule import TrafficRule
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.log_config import LogConfig, ResultLoggerType
from turing.router.config.enricher import Enricher
from turing.router.config.router_ensembler_config import RouterEnsemblerConfig
from turing.router.config.experiment_config import ExperimentConfig


NAME_INDEX = 0
VALUE_INDEX = 1


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
    :param experiment_engine: experiment engine config file
    :param resource_request: resources to be provisioned for the router
    :param timeout: request timeout which when exceeded, the request to the router will be terminated
    :param log_config: logging config settings to be used with the router
    :param enricher: enricher config settings to be used with the router
    :param ensembler: ensembler config settings to be used with the router
    """
    environment_name: str
    name: str
    routes: Union[List[Route], List[Dict[str, str]]] = None
    rules: Union[List[TrafficRule], List[Dict]] = None
    default_route_id: str = None
    experiment_engine: Union[ExperimentConfig, Dict] = None
    resource_request: Union[ResourceRequest, Dict[str, Union[str, int]]] = None
    timeout: str = None
    log_config: Union[LogConfig, Dict[str, Union[str, bool, int]]] = None
    enricher: Union[Enricher, Dict] = None
    ensembler: Union[RouterEnsemblerConfig, Dict] = None

    def __init__(self,
                 environment_name: str,
                 name: str,
                 routes: Union[List[Route], List[Dict[str, str]]] = None,
                 rules: Union[List[TrafficRule], List[Dict]] = None,
                 default_route_id: str = None,
                 experiment_engine: Union[ExperimentConfig, Dict] = None,
                 resource_request: Union[ResourceRequest, Dict[str, Union[str, int]]] = None,
                 timeout: str = None,
                 log_config: Union[LogConfig, Dict[str, Union[str, bool, int]]] = LogConfig(
                     result_logger_type=ResultLoggerType.NOP
                 ),
                 enricher: Union[Enricher, Dict] = None,
                 ensembler: Union[RouterEnsemblerConfig, Dict] = None,
                 **kwargs):
        self.environment_name = environment_name
        self.name = name
        self.routes = routes
        self.rules = rules
        self.default_route_id = default_route_id
        self.experiment_engine = experiment_engine
        self.resource_request = resource_request
        self.timeout = timeout
        self.log_config = log_config
        self.enricher = enricher
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
        if isinstance(routes, list):
            if all(isinstance(route, Route) for route in routes):
                self._routes = routes
            elif all(isinstance(route, dict) for route in routes):
                self._routes = [Route(**route) for route in routes]
            else:
                self._routes = routes
        else:
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
    def default_route_id(self) -> str:
        return self._default_route_id

    @default_route_id.setter
    def default_route_id(self, default_route_id: str):
        self._default_route_id = default_route_id

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
    def resource_request(self, resource_request: Union[ResourceRequest, Dict[str, Union[str, int]]]):
        if isinstance(resource_request, ResourceRequest):
            self._resource_request = resource_request
        elif isinstance(resource_request, dict):
            self._resource_request = ResourceRequest(**resource_request)
        else:
            self._resource_request = resource_request

    @property
    def timeout(self) -> str:
        return self._timeout

    @timeout.setter
    def timeout(self, timeout: str):
        self._timeout = timeout

    @property
    def log_config(self) -> LogConfig:
        return self._log_config

    @log_config.setter
    def log_config(self, log_config: Union[LogConfig, Dict[str, Union[str, bool, int]]]):
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
        if isinstance(ensembler, RouterEnsemblerConfig):
            self._ensembler = ensembler
        elif isinstance(ensembler, dict):
            self._ensembler = RouterEnsemblerConfig(**ensembler)
        else:
            self._ensembler = ensembler

    def to_open_api(self) -> OpenApiModel:
        kwargs = {}
        self._verify_no_duplicate_routes()

        if self.rules is not None:
            kwargs['rules'] = [rule.to_open_api() for rule in self.rules]
        if self.resource_request is not None:
            kwargs['resource_request'] = self.resource_request.to_open_api()
        if self.enricher is not None:
            kwargs['enricher'] = self.enricher.to_open_api()
        if self.ensembler is not None:
            kwargs['ensembler'] = self.ensembler.to_open_api()

        return turing.generated.models.RouterConfig(
            environment_name=self.environment_name,
            name=self.name,
            config=turing.generated.models.RouterConfigConfig(
                routes=[route.to_open_api() for route in self.routes],
                default_route_id=self.default_route_id,
                experiment_engine=self.experiment_engine.to_open_api(),
                timeout=self.timeout,
                log_config=self.log_config.to_open_api(),
                **kwargs
            )
        )

    def _verify_no_duplicate_routes(self):
        route_id_counter = Counter(route.id for route in self.routes)
        most_common_route_id, max_frequency = route_id_counter.most_common(n=1)[0]
        if max_frequency > 1:
            raise turing.router.config.route.DuplicateRouteException(
                f"Routes with duplicate ids are specified for this traffic rule. Duplicate id: {most_common_route_id}"
            )

    def to_dict(self):
        att_dict = {}
        for m in inspect.getmembers(self):
            if not inspect.ismethod(m[VALUE_INDEX]) and not m[NAME_INDEX].startswith('_'):
                att_dict[m[NAME_INDEX]] = m[VALUE_INDEX]
        return att_dict
