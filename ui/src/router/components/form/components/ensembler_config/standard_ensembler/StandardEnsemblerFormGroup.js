import React, { useEffect } from "react";
import { EuiFlexItem, EuiText } from "@elastic/eui";
import { get } from "../../../../../../components/form/utils";
import { StandardEnsembler } from "../../../../../../services/ensembler";
import { StandardEnsemblerPanel } from "./StandardEnsemblerPanel";
import { useOnChangeHandler } from "../../../../../../components/form/hooks/useOnChangeHandler";

export const StandardEnsemblerFormGroup = ({
  experimentConfig = {},
  routes,
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
      <EuiFlexItem>
        <StandardEnsemblerPanel
          experiments={experimentConfig.experiments}
          mappings={standardConfig.experiment_mappings}
          routeOptions={routeOptions}
          onChangeHandler={onChange("experiment_mappings")}
          errors={get(errors, "experiment_mappings")}
        />
      </EuiFlexItem>
    )
  );
};
