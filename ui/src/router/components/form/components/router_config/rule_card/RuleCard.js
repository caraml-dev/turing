import {
  EuiButtonIcon,
  EuiCard,
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiFormRow,
  EuiIcon,
  EuiSpacer,
  EuiSuperSelect,
} from "@elastic/eui";
import React, { Fragment, useCallback } from "react";
import { get } from "../../../../../../components/form/utils";
import { useOnChangeHandler } from "../../../../../../components/form/hooks/useOnChangeHandler";
import { RouteDropDownOption } from "../../RouteDropDownOption";
import { TrafficRuleCondition } from "../../../../traffic_rule_condition/TrafficRuleCondition";

import "./RuleCard.scss";

const noneRoute = {
  value: "_none_",
  inputDisplay: "Add a route to this rule...",
  disabled: true,
};

const newCondition = () => ({
  field_source: "header",
  field: "",
  operator: "in",
  values: [],
});

export const RuleCard = ({
  rule,
  routes,
  onChangeHandler,
  onDelete,
  errors,
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const routesOptions = useCallback(
    (item) => {
      return routes
        .filter((route) => route.id === item || !rule.routes.includes(route.id))
        .filter((route) => !!route.id && !!route.endpoint)
        .map((route) => ({
          value: route.id,
          inputDisplay: <RouteDropDownOption {...route} />,
        }));
    },
    [rule.routes, routes]
  );

  const onDeleteCondition = (idx) => () => {
    rule.conditions.splice(idx, 1);
    onChange("conditions")([...rule.conditions]);
  };

  const onDeleteRoute = (idx) => () => {
    rule.routes.splice(idx, 1);
    onChange("routes")([...rule.routes]);
  };

  return (
    <EuiCard
      className="euiCard--routeCard"
      title=""
      description=""
      textAlign="left">
      <EuiFlexGroup
        className="euiFlexGroup--removeButton"
        justifyContent="flexEnd"
        gutterSize="none"
        direction="row">
        <EuiFlexItem grow={false}>
          <EuiButtonIcon
            iconType="cross"
            onClick={onDelete}
            aria-label="delete-route"
          />
        </EuiFlexItem>
      </EuiFlexGroup>

      <EuiFormRow
        label="Name *"
        isInvalid={!!get(errors, "name")}
        error={get(errors, "name")}
        fullWidth>
        <EuiFieldText
          placeholder="rule-name"
          value={rule.name}
          onChange={(e) => onChange("name")(e.target.value)}
          isInvalid={!!get(errors, "name")}
          aria-label="rule-name"
          fullWidth
        />
      </EuiFormRow>

      <EuiFormRow
        label="Conditions *"
        isInvalid={!!get(errors, "conditions")}
        error={
          Array.isArray(get(errors, "conditions"))
            ? get(errors, "conditions")
            : []
        }
        fullWidth>
        <Fragment>
          {[...rule.conditions, newCondition()].map((condition, idx) => (
            <EuiFlexGroup
              className="euiFlexGroup--trafficRulesRow"
              key={`rule-conditions-${idx}`}
              direction="row"
              gutterSize="m"
              alignItems="flexStart">
              <EuiFlexItem grow={true}>
                <TrafficRuleCondition
                  condition={condition}
                  onChangeHandler={onChange(`conditions.${idx}`)}
                  errors={get(errors, `conditions.${idx}`)}
                />
              </EuiFlexItem>

              <EuiFlexItem
                grow={false}
                className="euiFlexItem--hasActions conditions">
                {idx < rule.conditions.length ? (
                  <EuiButtonIcon
                    size="s"
                    color="danger"
                    iconType="trash"
                    onClick={onDeleteCondition(idx)}
                    aria-label="Remove rule condition"
                  />
                ) : (
                  <EuiIcon type="empty" size="l" />
                )}
              </EuiFlexItem>
            </EuiFlexGroup>
          ))}
        </Fragment>
      </EuiFormRow>

      <EuiFormRow
        label="Routes *"
        isInvalid={!!get(errors, "routes")}
        error={
          Array.isArray(get(errors, "routes")) ? get(errors, "routes") : []
        }
        fullWidth>
        <Fragment>
          {[...rule.routes, "_none_"].map((route, idx) => (
            <Fragment key={`rule-routes-${idx}`}>
              <EuiFlexGroup
                className="euiFlexGroup--trafficRulesRow"
                direction="row"
                gutterSize="m"
                alignItems="flexStart">
                <EuiFlexItem grow={true} className="euiFlexItem--content">
                  <EuiFormRow
                    isInvalid={!!get(errors, `routes.${idx}`)}
                    error={
                      Array.isArray(get(errors, `routes.${idx}`))
                        ? get(errors, `routes.${idx}`)
                        : []
                    }
                    fullWidth>
                    <EuiSuperSelect
                      fullWidth
                      hasDividers
                      options={
                        idx < rule.routes.length
                          ? routesOptions(route)
                          : [noneRoute, ...routesOptions(route)]
                      }
                      valueOfSelected={route}
                      onChange={onChange(`routes.${idx}`)}
                      isInvalid={!!get(errors, `routes.${idx}`)}
                    />
                  </EuiFormRow>
                </EuiFlexItem>

                <EuiFlexItem
                  grow={false}
                  className="euiFlexItem--hasActions routes">
                  {idx < rule.routes.length ? (
                    <EuiButtonIcon
                      size="s"
                      color="danger"
                      iconType="trash"
                      onClick={onDeleteRoute(idx)}
                      aria-label="Remove rule route"
                    />
                  ) : (
                    <EuiIcon type="empty" size="l" />
                  )}
                </EuiFlexItem>
              </EuiFlexGroup>
              <EuiSpacer size="s" />
            </Fragment>
          ))}
        </Fragment>
      </EuiFormRow>
    </EuiCard>
  );
};
