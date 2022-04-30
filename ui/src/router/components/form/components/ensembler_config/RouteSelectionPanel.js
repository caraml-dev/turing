import React, { useEffect } from "react";
import { EuiFlexItem, EuiForm, EuiFormRow } from "@elastic/eui";

import { Panel } from "../Panel";
import { EuiComboBoxSelect } from "../../../../../components/form/combo_box/EuiComboBoxSelect";

const SelectRouteIdCombobox = ({ value, routeOptions, onChange, ...props }) => {
  // Note: routeNames changes on every render, which we need to capture name changes to routes
  const routeNames = routeOptions.map((e) => (!!e.value ? e.value : e.label));

  useEffect(() => {
    // Clear route if it doesn't exist in the options
    if (!!value && !routeNames.includes(value)) {
      onChange(undefined);
    }
    // Auto-select route if there is exactly one
    if (!value && routeNames.length === 1) {
      onChange(routeNames[0]);
    }
  }, [routeNames, value, onChange]);

  return (
    <EuiComboBoxSelect
      fullWidth={props.fullWidth}
      placeholder={props.placeholder}
      isInvalid={props.isInvalid}
      options={routeOptions}
      value={value || ""}
      onChange={onChange}
      isDisabled={routeOptions.length <= 1}
    />
  );
};

export const RouteSelectionPanel = ({
  routeId,
  routes,
  onChange,
  errors,
  panelTitle,
  inputLabel,
}) => {
  const routeOptions = routes
    .filter((e) => !!e.id)
    .map((e) => ({ label: e.id }));

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
              placeholder="Select route"
              onChange={onChange}
              isInvalid={!!errors}
            />
          </EuiFormRow>
        </EuiForm>
      </Panel>
    </EuiFlexItem>
  );
};
