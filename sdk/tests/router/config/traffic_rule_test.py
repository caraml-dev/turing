import turing.generated.models
import turing.router.config.route
import turing.router.config.traffic_rule
import pytest


@pytest.mark.parametrize(
    "field_source,expected", [
        pytest.param(
            "header",
            turing.generated.models.FieldSource("header")
        ),
        pytest.param(
            "payload",
            turing.generated.models.FieldSource("payload")
        )
    ])
def test_create_field_source(field_source, expected):
    actual = turing.router.config.traffic_rule.FieldSource(field_source).to_open_api()
    assert actual == expected


@pytest.mark.parametrize(
    "field_source,field,operator,values,expected", [
        pytest.param(
            turing.router.config.traffic_rule.FieldSource.HEADER,
            "x-region",
            "in",
            ["region-a", "region-b"],
            "generic_header_traffic_rule_condition"
        ),
        pytest.param(
            turing.router.config.traffic_rule.FieldSource.PAYLOAD,
            "service_type.id",
            "in",
            ["MyService", "YourService"],
            "generic_payload_traffic_rule_condition"
        )
    ])
def test_create_traffic_rule_condition(field_source, field, operator, values, expected, request):
    actual = turing.router.config.traffic_rule.TrafficRuleCondition(
        field_source=field_source,
        field=field,
        operator=operator,
        values=values
    ).to_open_api()
    assert actual == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "field_source,field,operator,values,expected", [
        pytest.param(
            turing.router.config.traffic_rule.FieldSource.HEADER,
            "x-region",
            "looks_like",
            ["region-a", "region-b"],
            turing.router.config.traffic_rule.InvalidOperatorException
        )
    ])
def test_create_traffic_rule_condition_with_invalid_operator(field_source, field, operator, values, expected):
    with pytest.raises(expected):
        turing.router.config.traffic_rule.TrafficRuleCondition(
            field_source=field_source,
            field=field,
            operator=operator,
            values=values
        ).to_open_api()


@pytest.mark.parametrize(
    "field,values,expected", [
        pytest.param(
            "x-region",
            ["region-a", "region-b"],
            "generic_header_traffic_rule_condition"
        )
    ])
def test_create_header_traffic_rule_condition(field, values, expected, request):
    actual = turing.router.config.traffic_rule.HeaderTrafficRuleCondition(
        field=field,
        values=values
    ).to_open_api()
    assert actual == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "field,values,expected", [
        pytest.param(
            "service_type.id",
            ["MyService", "YourService"],
            "generic_payload_traffic_rule_condition"
        )
    ])
def test_create_payload_traffic_rule_condition(field, values, expected, request):
    actual = turing.router.config.traffic_rule.PayloadTrafficRuleCondition(
        field=field,
        values=values
    ).to_open_api()
    assert actual == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "conditions,routes,expected", [
        pytest.param(
            [
                turing.router.config.traffic_rule.HeaderTrafficRuleCondition(
                    field="x-region",
                    values= ["region-a", "region-b"],
                ),
                turing.router.config.traffic_rule.PayloadTrafficRuleCondition(
                    field="service_type.id",
                    values=["MyService", "YourService"],
                )
            ],
            [
                turing.router.config.route.Route(
                    id="model-a",
                    endpoint="http://predict_this.io/model-a",
                    timeout=100)
            ],
            "generic_traffic_rule"
        )
    ])
def test_create_traffic_rule(conditions, routes, expected, request):
    actual = turing.router.config.traffic_rule.TrafficRule(
        conditions=conditions,
        routes=routes
    ).to_open_api()
    assert actual == request.getfixturevalue(expected)


@pytest.mark.parametrize(
    "conditions,routes,expected", [
        pytest.param(
            [
                turing.router.config.traffic_rule.HeaderTrafficRuleCondition(
                    field="x-region",
                    values= ["region-a", "region-b"],
                ),
                turing.router.config.traffic_rule.PayloadTrafficRuleCondition(
                    field="service_type.id",
                    values=["MyService", "YourService"],
                )
            ],
            [
                turing.router.config.route.Route(
                    id="model-a",
                    endpoint="http://predict_this.io/model-a",
                    timeout=100),
                turing.router.config.route.Route(
                    id="model-a",
                    endpoint="http://predict_this.io/model-a1",
                    timeout=100)
            ],
            turing.router.config.route.DuplicateRouteException
        )
    ])
def test_create_traffic_rule_with_duplicate_route_id(conditions, routes, expected):
    with pytest.raises(expected):
        turing.router.config.traffic_rule.TrafficRule(
            conditions=conditions,
            routes=routes
        ).to_open_api()
