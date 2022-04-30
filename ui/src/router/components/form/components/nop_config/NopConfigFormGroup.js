import React, { useEffect } from "react";

import { NopEnsembler } from "../../../../../services/ensembler";
import { FormLabelWithToolTip } from "../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";
import { RouteSelectionPanel } from "../ensembler_config/RouteSelectionPanel";

export const NopConfigFormGroup = ({
  routes,
  nopConfig,
  onChangeHandler,
  errors = {},
}) => {
  useEffect(() => {
    !nopConfig && onChangeHandler(NopEnsembler.newConfig());
  }, [nopConfig, onChangeHandler]);
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    !!nopConfig && (
      <RouteSelectionPanel
        routeId={nopConfig.final_response_route_id}
        routes={routes}
        onChange={onChange("final_response_route_id")}
        errors={errors.final_response_route_id}
        panelTitle="Response"
        inputLabel={
          <FormLabelWithToolTip
            label="Final Response *"
            content="Select the route to respond with"
          />
        }
      />
    )
  );
};
