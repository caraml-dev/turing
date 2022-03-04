# TrafficRule

Each traffic rule is defined by at least one `TrafficRuleCondition` and one route. Routes are essentially the `id`s 
of `Route` objects that you intend to specify for the entire `TrafficRule`.

```python
@dataclass
class TrafficRule:
    """
    Class to create a new TrafficRule based on a list of conditions and routes

    :param conditions: list of TrafficRuleConditions that need to ALL be satisfied before routing to the given routes
    :param routes: list of routes to send the request to should all the given conditions be met
    """
    conditions: Union[List[TrafficRuleCondition], List[Dict[str, List[str]]]]
    routes: List[str]
```

## Traffic Rule Condition

When defining a traffic rule, one would need to decide between using a `HeaderTrafficRuleCondition` or a 
`PayloadTrafficRuleCondition`. These subclasses can be used to build a `TrafficRuleCondition` without having to 
manually set attributes such as `field_source` or `operator`:

```python
@dataclass
class HeaderTrafficRuleCondition(TrafficRuleCondition):
    def __init__(self,
                 field: str,
                 values: List[str]):
        """
        Method to create a new TrafficRuleCondition that is defined on a request header

        :param field: name of the field specified
        :param values: values that are supposed to match those found in the field
        """
        super().__init__(field_source=FieldSource.HEADER, field=field, operator="in", values=values)


@dataclass
class PayloadTrafficRuleCondition(TrafficRuleCondition):
    def __init__(self,
                 field: str,
                 values: List[str]):
        """
        Method to create a new TrafficRuleCondition that is defined on a request payload

        :param field: name of the field specified
        :param values: values that are supposed to match those found in the field
        """
        super().__init__(field_source=FieldSource.PAYLOAD, field=field, operator="in", values=values)
```

