import {
  EuiButtonIcon,
  EuiCallOut,
  EuiButtonEmpty,
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
import { DraggableCardHeader } from "../../../../../../components/card/DraggableCardHeader";
import { RouteDropDownOption } from "../../RouteDropDownOption";
import { TrafficRuleCondition } from "../../../../traffic_rule_condition/TrafficRuleCondition";
import { FormLabelWithToolTip } from "@gojek/mlp-ui";

import "./RuleCard.scss";

const noneRoute = {
  value: "_none_",
  inputDisplay: "Add a route to this rule...",
  disabled: true,
};

export const RuleCard = ({
  isDefault,
  rule,
  routes,
  protocol,
  onChangeHandler,
  onDelete,
  errors,
  ...props
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const newCondition = () => ({
    field_source: protocol === "UPI_V1" ? "prediction_context" : "header",
    field: "",
    operator: "in",
    values: [],
  });

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

  // Hide last row when all routes are selected
  const routeSelectionOptions = (rule.routes.length < routes.length ? [...rule.routes,  "_none_"]: rule.routes);
  
  const conditionLabel = protocol === "UPI_V1" ? 
  <FormLabelWithToolTip
    label="Conditions *"
    content='Prediction Context of the UPI proto.  e.g. "model-a" in "model"'
  /> : "Conditions *"

  return (
    <EuiCard
      className="euiCard--routeCard"
      title={isDefault ? "Default Traffic Rule" : ""}
      description=""
      textAlign="left">
      {!isDefault &&
      <>
      <EuiFlexGroup
        className="euiFlexGroup--removeButton"
        justifyContent="flexEnd"
        gutterSize="s"
        direction="row">
        <EuiFlexItem>
          <DraggableCardHeader
            onDelete={onDelete}
            dragHandleProps={props.dragHandleProps}
          />
        </EuiFlexItem>
      </EuiFlexGroup>
      <EuiSpacer />
      </>
      }

      {!isDefault &&
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
      }

      {!isDefault ? <EuiFormRow
        label={conditionLabel}
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
                  protocol={protocol}
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
      : <EuiFormRow
          label="Conditions *"
          fullWidth>
          <EuiCallOut title="Conditions are disabled for the default rule." color="warning" iconType="help">
            <p>
              The default rule will be triggered if no other rule matches the request.
            </p>
          </EuiCallOut>
        </EuiFormRow>
      }

      <EuiFormRow
        label="Routes *"
        isInvalid={!!get(errors, "routes")}
        error={
          Array.isArray(get(errors, "routes")) ? get(errors, "routes") : []
        }
        fullWidth>
        <Fragment>
        {routeSelectionOptions.map((route, idx) => {
            return (
              <Fragment key={`rule-routes-${idx}`}>
                <EuiFlexGroup
                  className="euiFlexGroup--trafficRulesRow"
                  direction="row"
                  gutterSize="m"
                  alignItems="flexStart">
                  <EuiFlexItem grow={true} className="euiFlexItem--content">
                    <EuiFormRow
                      isInvalid={!!get(errors, `routes.${idx}`)}
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
                    ) : isDefault ? (
                      <EuiButtonEmpty size="xs" color="text" onClick={() => onChange("routes")(routes.map(route => route.id))}>
                       Select All
                      </EuiButtonEmpty>
                    )
                    : (
                      <EuiIcon type="empty" size="l" />
                    )}
                  </EuiFlexItem>
                </EuiFlexGroup>
                <EuiSpacer size="s" />
              </Fragment>
            )
        })}
        </Fragment>
      </EuiFormRow>
    </EuiCard>
  );
};
