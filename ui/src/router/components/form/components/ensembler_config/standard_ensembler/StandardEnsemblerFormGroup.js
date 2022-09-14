import React, { useEffect } from "react";
import { EuiFlexItem, EuiText } from "@elastic/eui";

import { get } from "../../../../../../components/form/utils";
import { StandardEnsembler } from "../../../../../../services/ensembler";
import { StandardEnsemblerPanel } from "./StandardEnsemblerPanel";
import { RouteSelectionPanel } from "../RouteSelectionPanel";
import { FormLabelWithToolTip } from "../../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { useOnChangeHandler } from "../../../../../../components/form/hooks/useOnChangeHandler";

export const StandardEnsemblerFormGroup = ({
  experimentConfig = {},
  routes,
  rules,
  default_traffic_rule,
  standardConfig,
  onChangeHandler,
  errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  useEffect(() => {
    !standardConfig && onChangeHandler(StandardEnsembler.newConfig());
  }, [standardConfig, onChangeHandler]);

  const routeOptions = [
    {
      value: "nop",
      inputDisplay: "Select a route...",
      disabled: true,
    },
    ...routes.map((route) => ({
      icon: "graphApp",
      value: route.id,
      inputDisplay: route.id,
      dropdownDisplay: (
        <>
          <strong>{route.id}</strong>
          <EuiText color="subdued" size="s">
            {route.endpoint}
          </EuiText>
        </>
      ),
    })),
  ];

  return (
    !!standardConfig && (
      <>
        <EuiFlexItem>
          <StandardEnsemblerPanel
            experiments={experimentConfig.experiments}
            mappings={standardConfig.experiment_mappings}
            routeOptions={routeOptions}
            onChangeHandler={onChange("experiment_mappings")}
            errors={get(errors, "experiment_mappings")}
          />
        </EuiFlexItem>
        <EuiFlexItem>
          <RouteSelectionPanel
            routeId={standardConfig.fallback_response_route_id}
            routes={routes}
            rules={rules}
            default_traffic_rule={default_traffic_rule}
            onChange={onChange("fallback_response_route_id")}
            errors={get(errors, "fallback_response_route_id")}
            panelTitle="Fallback"
            inputLabel={
              <FormLabelWithToolTip
                label="Fallback Response *"
                content="Select the route to fallback to, if the call to the experiment engine fails or if a matching route cannot be found for the treatment."
              />
            }
          />
        </EuiFlexItem>
      </>
    )
  );
};
