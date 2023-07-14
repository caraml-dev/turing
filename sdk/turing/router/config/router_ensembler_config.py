from dataclasses import dataclass, field

from turing._base_types import DataObject
import turing.generated.models
from typing import List, Dict, Union, Optional
from turing.generated.model_utils import OpenApiModel
from turing.router.config.autoscaling_policy import (
    AutoscalingPolicy,
    DEFAULT_AUTOSCALING_POLICY,
)
from turing.router.config.resource_request import ResourceRequest
from turing.router.config.common.env_var import EnvVar


@dataclass
class EnsemblerNopConfig:
    final_response_route_id: str

    _final_response_route_id: str = field(init=False, repr=False)

    @property
    def final_response_route_id(self) -> str:
        return self._final_response_route_id

    @final_response_route_id.setter
    def final_response_route_id(self, final_response_route_id: str):
        self._final_response_route_id = final_response_route_id

    def to_open_api(self) -> OpenApiModel:
        return None


@dataclass
class EnsemblerStandardConfig:
    fallback_response_route_id: str
    experiment_mappings: Optional[
        List[turing.generated.models.EnsemblerStandardConfigExperimentMappings]
    ]
    route_name_path: Optional[str]
    lazy_routing: bool

    _fallback_response_route_id: str = field(init=False, repr=False)
    _experiment_mappings: Optional[
        List[turing.generated.models.EnsemblerStandardConfigExperimentMappings]
    ] = field(init=False, repr=False)
    _route_name_path: Optional[str] = field(init=False, repr=False)
    _lazy_routing: bool = field(init=False, repr=False)

    @property
    def experiment_mappings(
        self,
    ) -> List[turing.generated.models.EnsemblerStandardConfigExperimentMappings]:
        return self._experiment_mappings

    @experiment_mappings.setter
    def experiment_mappings(
        self,
        experiment_mappings: List[
            turing.generated.models.EnsemblerStandardConfigExperimentMappings
        ],
    ):
        self._experiment_mappings = experiment_mappings

    @property
    def route_name_path(self) -> str:
        return self._route_name_path

    @route_name_path.setter
    def route_name_path(self, route_name_path: str):
        self._route_name_path = route_name_path

    @property
    def fallback_response_route_id(self) -> str:
        return self._fallback_response_route_id

    @fallback_response_route_id.setter
    def fallback_response_route_id(self, fallback_response_route_id: str):
        self._fallback_response_route_id = fallback_response_route_id

    @property
    def lazy_routing(self) -> bool:
        return self._lazy_routing

    @lazy_routing.setter
    def lazy_routing(self, lazy_routing: bool):
        self._lazy_routing = lazy_routing

    def to_open_api(self) -> OpenApiModel:
        kwargs = {}
        if self.experiment_mappings is not None:
            kwargs["experiment_mappings"] = self.experiment_mappings
        if self.route_name_path is not None:
            kwargs["route_name_path"] = self.route_name_path

        return turing.generated.models.EnsemblerStandardConfig(
            lazy_routing=self.lazy_routing,
            **kwargs,
        )


@dataclass
class RouterEnsemblerConfig(DataObject):
    """
    Class to create a new RouterEnsemblerConfig

    :param type: type of the ensembler; must be one of {'nop', 'standard', 'docker', 'pyfunc'}
    :param id: id of the ensembler
    :param standard_config: EnsemblerStandardConfig instance containing mappings between routes and treatments
    :param docker_config: EnsemblerDockerConfig instance containing configs for the docker ensembler
    """

    type: str
    id: int = None
    nop_config: EnsemblerNopConfig = None
    standard_config: EnsemblerStandardConfig = None
    docker_config: turing.generated.models.EnsemblerDockerConfig = None
    pyfunc_config: turing.generated.models.EnsemblerPyfuncConfig = None

    def __init__(
        self,
        type: str,
        id: int = None,
        nop_config: EnsemblerNopConfig = None,
        standard_config: EnsemblerStandardConfig = None,
        docker_config: turing.generated.models.EnsemblerDockerConfig = None,
        pyfunc_config: turing.generated.models.EnsemblerPyfuncConfig = None,
        **kwargs,
    ):
        self.id = id
        self.type = type
        self.nop_config = nop_config
        self.standard_config = standard_config
        self.docker_config = docker_config
        self.pyfunc_config = pyfunc_config

    @property
    def id(self) -> int:
        return self._id

    @id.setter
    def id(self, id: int):
        self._id = id

    @property
    def type(self) -> str:
        return self._type

    @type.setter
    def type(self, type: str):
        assert type in {"nop", "standard", "docker", "pyfunc"}
        self._type = type

    @property
    def standard_config(self) -> EnsemblerStandardConfig:
        return self._standard_config

    @standard_config.setter
    def standard_config(self, standard_config: Union[EnsemblerStandardConfig, Dict]):
        if isinstance(standard_config, EnsemblerStandardConfig):
            self._standard_config = standard_config
        elif isinstance(standard_config, dict):
            openapi_standard_config = standard_config.copy()
            openapi_standard_config["experiment_mappings"] = None
            if "experiment_mappings" in standard_config and standard_config["experiment_mappings"] is not None:
                openapi_standard_config["experiment_mappings"] = [
                    turing.generated.models.EnsemblerStandardConfigExperimentMappings(
                        **mapping
                    )
                    for mapping in standard_config["experiment_mappings"]
                ]
            self._standard_config = EnsemblerStandardConfig(**openapi_standard_config)
        else:
            self._standard_config = standard_config

    @property
    def docker_config(self) -> turing.generated.models.EnsemblerDockerConfig:
        return self._docker_config

    @docker_config.setter
    def docker_config(
        self, docker_config: turing.generated.models.EnsemblerDockerConfig
    ):
        if isinstance(docker_config, turing.generated.models.EnsemblerDockerConfig):
            self._docker_config = docker_config
        elif isinstance(docker_config, dict):
            openapi_docker_config = docker_config.copy()
            openapi_docker_config[
                "resource_request"
            ] = turing.generated.models.ResourceRequest(
                **openapi_docker_config["resource_request"]
            )
            openapi_docker_config["autoscaling_policy"] = (
                turing.generated.models.AutoscalingPolicy(
                    **openapi_docker_config["autoscaling_policy"]
                )
                if "autoscaling_policy" in openapi_docker_config
                else DEFAULT_AUTOSCALING_POLICY.to_open_api()
            )
            openapi_docker_config["env"] = [
                turing.generated.models.EnvVar(**env_var)
                for env_var in docker_config["env"]
            ]
            self._docker_config = turing.generated.models.EnsemblerDockerConfig(
                **openapi_docker_config
            )
        else:
            self._docker_config = docker_config

    @property
    def pyfunc_config(self) -> turing.generated.models.EnsemblerPyfuncConfig:
        return self._pyfunc_config

    @pyfunc_config.setter
    def pyfunc_config(
        self, pyfunc_config: turing.generated.models.EnsemblerPyfuncConfig
    ):
        if isinstance(pyfunc_config, turing.generated.models.EnsemblerPyfuncConfig):
            self._pyfunc_config = pyfunc_config
        elif isinstance(pyfunc_config, dict):
            openapi_pyfunc_config = pyfunc_config.copy()
            openapi_pyfunc_config[
                "resource_request"
            ] = turing.generated.models.ResourceRequest(
                **pyfunc_config["resource_request"]
            )
            openapi_pyfunc_config["autoscaling_policy"] = (
                turing.generated.models.AutoscalingPolicy(
                    **openapi_pyfunc_config["autoscaling_policy"]
                )
                if "autoscaling_policy" in openapi_pyfunc_config
                else DEFAULT_AUTOSCALING_POLICY.to_open_api()
            )
            openapi_pyfunc_config["env"] = [
                turing.generated.models.EnvVar(**env_var)
                for env_var in pyfunc_config["env"]
            ]
            self._pyfunc_config = turing.generated.models.EnsemblerPyfuncConfig(
                **openapi_pyfunc_config
            )
        else:
            self._pyfunc_config = pyfunc_config

    @property
    def nop_config(self) -> EnsemblerNopConfig:
        return self._nop_config

    @nop_config.setter
    def nop_config(self, nop_config: EnsemblerNopConfig):
        if isinstance(nop_config, EnsemblerNopConfig):
            self._nop_config = nop_config
        elif isinstance(nop_config, dict):
            self._nop_config = EnsemblerNopConfig(**nop_config)
        else:
            self._nop_config = nop_config

    def to_open_api(self) -> OpenApiModel:
        kwargs = {}

        if self.standard_config is not None:
            kwargs["standard_config"] = self.standard_config.to_open_api()
        if self.docker_config is not None:
            kwargs["docker_config"] = self.docker_config
        if self.pyfunc_config is not None:
            kwargs["pyfunc_config"] = self.pyfunc_config

        return turing.generated.models.RouterEnsemblerConfig(type=self.type, **kwargs)


@dataclass
class PyfuncRouterEnsemblerConfig(RouterEnsemblerConfig):
    def __init__(
        self,
        project_id: int,
        ensembler_id: int,
        timeout: str,
        resource_request: ResourceRequest,
        env: List["EnvVar"],
        autoscaling_policy: AutoscalingPolicy = DEFAULT_AUTOSCALING_POLICY,
    ):
        """
        Method to create a new Pyfunc ensembler

        :param project_id: project id of the current project
        :param ensembler_id: ensembler_id of the ensembler
        :param resource_request: ResourceRequest instance containing configs related to the resources required
        :param autoscaling_policy: AutoscalingPolicy instance containing configs for the deployment autoscaling
        :param timeout: request timeout which when exceeded, the request to the ensembler will be terminated
        :param env: environment variables required by the container
        """
        self.project_id = project_id
        self.ensembler_id = ensembler_id
        self.resource_request = resource_request
        self.autoscaling_policy = autoscaling_policy
        self.timeout = timeout
        self.env = env
        super().__init__(type="pyfunc")

    @property
    def project_id(self) -> int:
        return self._project_id

    @project_id.setter
    def project_id(self, project_id: int):
        self._project_id = project_id

    @property
    def ensembler_id(self) -> int:
        return self._ensembler_id

    @ensembler_id.setter
    def ensembler_id(self, ensembler_id: int):
        self._ensembler_id = ensembler_id

    @property
    def resource_request(self) -> ResourceRequest:
        return self._resource_request

    @resource_request.setter
    def resource_request(self, resource_request: ResourceRequest):
        self._resource_request = resource_request

    @property
    def autoscaling_policy(self) -> AutoscalingPolicy:
        return self._autoscaling_policy

    @autoscaling_policy.setter
    def autoscaling_policy(self, autoscaling_policy: AutoscalingPolicy):
        self._autoscaling_policy = autoscaling_policy

    @property
    def timeout(self) -> str:
        return self._timeout

    @timeout.setter
    def timeout(self, timeout: str):
        self._timeout = timeout

    @property
    def env(self) -> List["EnvVar"]:
        return self._env

    @env.setter
    def env(self, env: List["EnvVar"]):
        self._env = env

    @classmethod
    def from_config(
        cls, config: turing.generated.models.EnsemblerPyfuncConfig
    ) -> "PyfuncRouterEnsemblerConfig":
        return cls(
            project_id=config.project_id,
            ensembler_id=config.ensembler_id,
            timeout=config.timeout,
            resource_request=ResourceRequest(
                min_replica=config.resource_request.min_replica,
                max_replica=config.resource_request.max_replica,
                cpu_request=config.resource_request.cpu_request,
                memory_request=config.resource_request.memory_request,
            ),
            autoscaling_policy=AutoscalingPolicy(
                metric=config.autoscaling_policy.metric,
                target=config.autoscaling_policy.target,
            )
            if config.autoscaling_policy is not None
            else DEFAULT_AUTOSCALING_POLICY,
            env=[EnvVar(name=env.name, value=env.value) for env in config.env],
        )

    def to_open_api(self) -> OpenApiModel:
        assert all(isinstance(env_var, EnvVar) for env_var in self.env)

        self.pyfunc_config = turing.generated.models.EnsemblerPyfuncConfig(
            project_id=self.project_id,
            ensembler_id=self.ensembler_id,
            resource_request=self.resource_request.to_open_api(),
            autoscaling_policy=self.autoscaling_policy.to_open_api(),
            timeout=self.timeout,
            env=[env_var.to_open_api() for env_var in self.env],
        )
        return super().to_open_api()


@dataclass
class DockerRouterEnsemblerConfig(RouterEnsemblerConfig):
    def __init__(
        self,
        image: str,
        resource_request: ResourceRequest,
        endpoint: str,
        timeout: str,
        port: int,
        env: List["EnvVar"],
        service_account: str = None,
        autoscaling_policy: AutoscalingPolicy = DEFAULT_AUTOSCALING_POLICY,
    ):
        """
        Method to create a new Docker ensembler

        :param image: registry and name of the image
        :param resource_request: ResourceRequest instance containing configs related to the resources required
        :param autoscaling_policy: AutoscalingPolicy instance containing configs for autoscaling of the deployment
        :param endpoint: endpoint URL of the ensembler
        :param timeout: request timeout which when exceeded, the request to the ensembler will be terminated
        :param port: port number exposed by the container
        :param env: environment variables required by the container
        :param service_account: optional service account for the Docker deployment
        """
        self.image = image
        self.resource_request = resource_request
        self.autoscaling_policy = autoscaling_policy
        self.endpoint = endpoint
        self.timeout = timeout
        self.port = port
        self.env = env
        self.service_account = service_account
        super().__init__(type="docker")

    @property
    def image(self) -> str:
        return self._image

    @image.setter
    def image(self, image: str):
        self._image = image

    @property
    def resource_request(self) -> ResourceRequest:
        return self._resource_request

    @resource_request.setter
    def resource_request(self, resource_request: ResourceRequest):
        self._resource_request = resource_request

    @property
    def autoscaling_policy(self) -> AutoscalingPolicy:
        return self._autoscaling_policy

    @autoscaling_policy.setter
    def autoscaling_policy(self, autoscaling_policy: AutoscalingPolicy):
        self._autoscaling_policy = autoscaling_policy

    @property
    def endpoint(self) -> str:
        return self._endpoint

    @endpoint.setter
    def endpoint(self, endpoint: str):
        self._endpoint = endpoint

    @property
    def timeout(self) -> str:
        return self._timeout

    @timeout.setter
    def timeout(self, timeout: str):
        self._timeout = timeout

    @property
    def port(self) -> int:
        return self._port

    @port.setter
    def port(self, port: int):
        self._port = port

    @property
    def env(self) -> List["EnvVar"]:
        return self._env

    @env.setter
    def env(self, env: List["EnvVar"]):
        self._env = env

    @property
    def service_account(self) -> str:
        return self._service_account

    @service_account.setter
    def service_account(self, service_account: str):
        self._service_account = service_account

    @classmethod
    def from_config(
        cls, config: turing.generated.models.EnsemblerDockerConfig
    ) -> "DockerRouterEnsemblerConfig":
        return cls(
            image=config.image,
            resource_request=ResourceRequest(
                min_replica=config.resource_request.min_replica,
                max_replica=config.resource_request.max_replica,
                cpu_request=config.resource_request.cpu_request,
                memory_request=config.resource_request.memory_request,
            ),
            autoscaling_policy=AutoscalingPolicy(
                metric=config.autoscaling_policy.metric,
                target=config.autoscaling_policy.target,
            )
            if config.autoscaling_policy is not None
            else DEFAULT_AUTOSCALING_POLICY,
            endpoint=config.endpoint,
            timeout=config.timeout,
            port=config.port,
            env=[EnvVar(name=env.name, value=env.value) for env in config.env],
            service_account=config["service_account"],
        )

    def to_open_api(self) -> OpenApiModel:
        assert all(isinstance(env_var, EnvVar) for env_var in self.env)

        kwargs = {}
        if self.service_account is not None:
            kwargs["service_account"] = self.service_account

        self.docker_config = turing.generated.models.EnsemblerDockerConfig(
            image=self.image,
            resource_request=self.resource_request.to_open_api(),
            autoscaling_policy=self.autoscaling_policy.to_open_api(),
            endpoint=self.endpoint,
            timeout=self.timeout,
            port=self.port,
            env=[env_var.to_open_api() for env_var in self.env],
            **kwargs,
        )
        return super().to_open_api()


@dataclass
class StandardRouterEnsemblerConfig(RouterEnsemblerConfig):
    def __init__(
        self,
        fallback_response_route_id: str,
        experiment_mappings: Optional[List[Dict[str, str]]] = None,
        route_name_path: Optional[str] = None,
        lazy_routing: Optional[bool] = False,
    ):
        """
        Method to create a new standard ensembler

        :param route_name_path: configured route name path that points to the route name within a given treatment config
        :param experiment_mappings: configured mappings between routes and treatments
        :param fallback_response_route_id: configured final response route to be used as a fallback
        :param lazy_routing: controls whether the experiment engine should be called before the routes (True) or in parallel (False)
        """
        self.route_name_path = route_name_path
        self.experiment_mappings = experiment_mappings
        self.fallback_response_route_id = fallback_response_route_id
        self.lazy_routing = lazy_routing
        super().__init__(type="standard")

    @property
    def experiment_mappings(self) -> List[Dict[str, str]]:
        return self._experiment_mappings

    @experiment_mappings.setter
    def experiment_mappings(self, experiment_mappings: Optional[List[Dict[str, str]]]):
        if experiment_mappings is not None:
            StandardRouterEnsemblerConfig._verify_experiment_mappings(
                experiment_mappings
            )
        self._experiment_mappings = experiment_mappings

    @property
    def route_name_path(self) -> str:
        return self._route_name_path

    @route_name_path.setter
    def route_name_path(self, route_name_path: str):
        self._route_name_path = route_name_path

    @property
    def fallback_response_route_id(self) -> str:
        return self._fallback_response_route_id

    @fallback_response_route_id.setter
    def fallback_response_route_id(self, fallback_response_route_id: str):
        self._fallback_response_route_id = fallback_response_route_id

    @property
    def lazy_routing(self) -> bool:
        return self._lazy_routing

    @lazy_routing.setter
    def lazy_routing(self, lazy_routing: bool):
        self._lazy_routing = lazy_routing

    @classmethod
    def _verify_experiment_mappings(cls, experiment_mappings: List[Dict[str, str]]):
        for experiment_mapping in experiment_mappings:
            try:
                assert isinstance(experiment_mapping["experiment"], str)
                assert isinstance(experiment_mapping["treatment"], str)
                assert isinstance(experiment_mapping["route"], str)
            except (KeyError, AssertionError):
                raise InvalidExperimentMappingException(
                    "Experiment mapping should contain the keys {'experiment', 'treatment', 'route'}; "
                    f"experiment_mapping passed: {experiment_mapping}"
                )

    @classmethod
    def from_config(
        cls, config: EnsemblerStandardConfig
    ) -> "StandardRouterEnsemblerConfig":
        return cls(
            experiment_mappings=[e.to_dict() for e in config.experiment_mappings]
            if config.experiment_mappings
            else None,
            route_name_path=config.route_name_path,
            fallback_response_route_id=config.fallback_response_route_id,
            lazy_routing=config.lazy_routing,
        )

    def to_open_api(self) -> OpenApiModel:
        self.standard_config = EnsemblerStandardConfig(
            experiment_mappings=[
                turing.generated.models.EnsemblerStandardConfigExperimentMappings(
                    **experiment_mapping
                )
                for experiment_mapping in self.experiment_mappings
            ]
            if self.experiment_mappings
            else None,
            route_name_path=self.route_name_path,
            fallback_response_route_id=self.fallback_response_route_id,
            lazy_routing=self.lazy_routing,
        )
        return super().to_open_api()


@dataclass
class NopRouterEnsemblerConfig(RouterEnsemblerConfig):
    def __init__(self, final_response_route_id: str):
        """
        Method to create a new Nop ensembler

        :param final_response_route_id: The route id of the route to be returned as the final response
        """
        self.final_response_route_id = final_response_route_id
        super().__init__(type="nop")

    @property
    def final_response_route_id(self) -> str:
        return self._final_response_route_id

    @final_response_route_id.setter
    def final_response_route_id(self, final_response_route_id: str):
        self._final_response_route_id = final_response_route_id

    @classmethod
    def from_config(cls, config: EnsemblerNopConfig) -> "NopRouterEnsemblerConfig":
        return cls(final_response_route_id=config.final_response_route_id)

    def to_open_api(self) -> OpenApiModel:
        self.nop_config = EnsemblerNopConfig(
            final_response_route_id=self.final_response_route_id
        )
        # Nop config is not passed down to the API. The final_response_route_id property
        # will be parsed in the router config and copied over as appropriate.
        return None


class InvalidExperimentMappingException(Exception):
    pass


class InvalidEnsemblerTypeException(Exception):
    pass
