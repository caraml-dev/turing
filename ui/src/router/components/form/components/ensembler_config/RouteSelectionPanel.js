import React, { useEffect } from "react";
import { EuiFlexItem, EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";

import { RouteDropDownOption } from "../RouteDropDownOption";
import { Panel } from "../Panel";

const SelectRouteIdCombobox = ({ value, routeOptions, onChange, ...props }) => {
  const noneValue = "_none_";

  useEffect(() => {
    const routeNames = routeOptions
      .filter((e) => !e.disabled)
      .map((e) => e.value);
    // Clear route if it doesn't exist in the enabled options
    if (value !== noneValue && !routeNames.includes(value)) {
      onChange(noneValue);
    }
    // Auto-select route if there is exactly one
    if ((!value || value === noneValue) && routeNames.length === 1) {
      onChange(routeNames[0]);
    }
  }, [routeOptions, value, onChange]);

  const routeOptionsWithNone = [
    // Hack to have a placeholder for EuiSuperSelect
    {
      value: noneValue,
      inputDisplay: props.placeholder,
      disabled: true,
    },
    ...routeOptions,
  ];

  return (
    <EuiSuperSelect
      fullWidth={props.fullWidth}
      itemLayoutAlign="top"
      hasDividers
      isInvalid={props.isInvalid}
      options={routeOptionsWithNone}
      value={value || ""}
      onChange={onChange}
      disabled={routeOptionsWithNone.length <= 1}
    />
  );
};

export const RouteSelectionPanel = ({
  routeId,
  routes,
  rules,
  onChange,
  errors,
  panelTitle,
  inputLabel,
}) => {
  const routeOptions = routes
    .filter((e) => !!e.id)
    .map((e) => {
      const isDisabled =
        !!rules && rules.filter((r) => r.routes.includes(e.id)).length > 0;
      return {
        value: e.id,
        inputDisplay: e.id,
        dropdownDisplay: (
          <RouteDropDownOption
            id={e.id}
            endpoint={e.endpoint}
            isDisabled={isDisabled}
            disabledOptionTooltip="Route with traffic rules cannot be selected"
          />
        ),
        disabled: isDisabled,
      };
    });

  return (
    <EuiFlexItem>
      <Panel title={panelTitle}>
        <EuiForm>
          <EuiFormRow
            label={inputLabel}
            isInvalid={!!errors}
            error={errors}
            fullWidth
            display="row">
            <SelectRouteIdCombobox
              fullWidth
              value={routeId}
              routeOptions={routeOptions}
              placeholder="Select route..."
              onChange={onChange}
              isInvalid={!!errors}
            />
          </EuiFormRow>
        </EuiForm>
      </Panel>
    </EuiFlexItem>
  );
};
