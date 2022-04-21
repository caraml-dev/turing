import React, { useEffect } from "react";
import { EuiFlexItem, EuiForm, EuiFormRow } from "@elastic/eui";

import { Panel } from "../Panel";
import { NopEnsembler } from "../../../../../services/ensembler";
import { EuiComboBoxSelect } from "../../../../../components/form/combo_box/EuiComboBoxSelect";
import { FormLabelWithToolTip } from "../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";

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

const NopConfigPanel = ({ routeId, routes, onChangeHandler, errors }) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const routeOptions = routes
    .filter((e) => !!e.id)
    .map((e) => ({ label: e.id }));

  return (
    <EuiFlexItem>
      <Panel title="Response">
        <EuiForm>
          <EuiFormRow
            label={
              <FormLabelWithToolTip
                label="Final Response *"
                content="Select the route to respond with"
              />
            }
            isInvalid={!!errors}
            error={errors}
            fullWidth
            display="row">
            <SelectRouteIdCombobox
              fullWidth
              value={routeId}
              routeOptions={routeOptions}
              placeholder="Select route"
              onChange={onChange("final_response_route_id")}
              isInvalid={!!errors}
            />
          </EuiFormRow>
        </EuiForm>
      </Panel>
    </EuiFlexItem>
  );
};

export const NopConfigFormGroup = ({
  routes,
  nopConfig,
  onChangeHandler,
  errors = {},
}) => {
  useEffect(() => {
    !nopConfig && onChangeHandler(NopEnsembler.newConfig());
  }, [nopConfig, onChangeHandler]);

  return (
    !!nopConfig && (
      <NopConfigPanel
        routeId={nopConfig.final_response_route_id}
        routes={routes}
        onChangeHandler={onChangeHandler}
        errors={errors.final_response_route_id}
      />
    )
  );
};
