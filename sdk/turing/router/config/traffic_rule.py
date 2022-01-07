import abc
from enum import Enum
from collections import Counter

from typing import Iterable, MutableMapping, Optional, Dict, List
import turing.generated.models
from turing.router.config.route import Route
from turing.generated.model_utils import OpenApiModel


class FieldSource(Enum):
    HEADER = "header"
    PAYLOAD = "payload"

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.FieldSource(self.value)


class TrafficRuleCondition:
    def __init__(self,
                 field_source: str,
                 field: str,
                 operator: str,
                 values: List[str]):
        try:
            assert operator == "in"
        except AssertionError:
            raise InvalidOperatorException(f"Invalid operator passed: {operator}")

        self._field_source = FieldSource(field_source)
        self._field = field
        self._operator = operator
        self._values = values

    @property
    def field_source(self) -> FieldSource:
        return self._field_source

    @property
    def field(self) -> str:
        return self._field

    @property
    def operator(self) -> str:
        return self._operator

    @property
    def values(self) -> List[str]:
        return self._values

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.TrafficRuleCondition(
            field_source=self.field_source.to_open_api(),
            field=self.field,
            operator=self.operator,
            values=self.values
        )


class InvalidOperatorException(Exception):
    pass


class HeaderTrafficRuleCondition(TrafficRuleCondition):
    def __init__(self,
                 field: str,
                 values: List[str]):
        super().__init__(field_source="header", field=field, operator="in", values=values)


class PayloadTrafficRuleCondition(TrafficRuleCondition):
    def __init__(self,
                 field: str,
                 values: List[str]):
        super().__init__(field_source="payload", field=field, operator="in", values=values)


class TrafficRule:
    def __init__(self,
                 conditions: List[TrafficRuleCondition],
                 routes: List[Route]):
        self._conditions = conditions
        self._routes = routes

    @property
    def conditions(self) -> List[TrafficRuleCondition]:
        return self._conditions

    @property
    def routes(self) -> List[Route]:
        return self._routes

    def to_open_api(self) -> OpenApiModel:
        self._verify_no_duplicate_routes()

        return turing.generated.models.TrafficRule(
            conditions=[traffic_rule_condition.to_open_api() for traffic_rule_condition in self.conditions],
            routes=[route.id for route in self.routes]
        )

    def _verify_no_duplicate_routes(self):
        route_id_counter = Counter(route.id for route in self.routes)
        most_common_route_id, max_frequency = route_id_counter.most_common(n=1)[0]
        if max_frequency > 1:
            raise turing.router.config.route.DuplicateRouteException(
                f"Routes with duplicate ids are specified for this traffic rule. Duplicate id: {most_common_route_id}"
            )
