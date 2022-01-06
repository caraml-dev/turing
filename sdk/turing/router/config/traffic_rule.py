import abc
from enum import Enum
from typing import Iterable, MutableMapping, Optional, Dict, List
import turing.generated.models
from turing._base_types import DataObject
from turing.generated.model_utils import OpenApiModel


class FieldSource(Enum):
    HEADER = "header"
    PAYLOAD = "payload"


class TrafficRuleCondition:
    def __init__(self,
                 field_source: str,
                 field: str,
                 operator: str,
                 values: List[str]):
        assert operator == "in"
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
        assert self.operator == "in"
        return turing.generated.models.TrafficRuleCondition(
            field_source=self.field_source.value,
            field=self.field,
            operator=self.operator,
            values=self.values
        )


class TrafficRule:
    def __init__(self,
                 conditions: List[TrafficRuleCondition],
                 routes: List[str]):
        self._conditions = conditions
        self._routes = routes

    @property
    def conditions(self) -> List[TrafficRuleCondition]:
        return self._conditions

    @property
    def routes(self) -> List[str]:
        return self._routes

    def to_open_api(self) -> OpenApiModel:
        return turing.generated.models.TrafficRule(
            conditions=[traffic_rule_condition.to_open_api() for traffic_rule_condition in self.conditions],
            routes=self.routes
        )
