import React, { useEffect, useContext, useRef } from "react";
import { EuiFlexItem, EuiLoadingChart } from "@elastic/eui";
import { initConfig } from "./config";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";
import { OverlayMask } from "../../../../../components/overlay_mask/OverlayMask";
import { ExperimentContextProvider } from "./providers/ExperimentContextProvider";
import ExperimentEngineContext from "../../../../../providers/experiments/context";
import { ClientConfigPanel } from "./components/client_config/ClientConfigPanel";
import { ExperimentsConfigPanel } from "./components/experiments_config/ExperimentsConfigPanel";
import { VariablesConfigPanel } from "./components/variables_config/VariablesConfigPanel";

export const StandardExperimentConfigGroup = ({
  engineType,
  experimentConfig,
  onChangeHandler,
  errors = {},
}) => {
  const experimentSectionRef = useRef();

  // Get engine's properties
  const { getEngineProperties } = useContext(ExperimentEngineContext);
  const engineProps = getEngineProperties(engineType);

  const { onChange } = useOnChangeHandler(onChangeHandler);

  // Set engineProps to the experiment config (for validation and comparison of changes to engine type),
  // if not already exists. Set experimentConfig if empty or reset it if the type has changed.
  useEffect(() => {
    if (!!engineProps.name) {
      if (
        !experimentConfig ||
        (experimentConfig.engine &&
          experimentConfig.engine.name &&
          experimentConfig.engine.name !== engineProps.name)
      ) {
        onChangeHandler({
          ...initConfig(),
          engine: engineProps,
        });
      } else if (!experimentConfig.engine || !experimentConfig.engine.name) {
        // Set the engine props to the experiment config
        onChange("engine")(engineProps);
      }
    }
  }, [experimentConfig, engineProps, onChange, onChangeHandler]);

  /* Not memoizing this because useMemo would check for the referential equality
     while the id property of each experiment will be set after the experiment
     element is created, and we want the change to be detected correctly. */
  const selectedExpIds = ((experimentConfig || {}).experiments || [])
    .filter((exp) => !!exp.id)
    .map((exp) => exp.id)
    .sort((a, b) => (a > b ? 1 : -1));

  return !!experimentConfig &&
    !!experimentConfig.engine &&
    !!experimentConfig.engine.name ? (
    <ExperimentContextProvider
      engineProps={experimentConfig.engine}
      clientId={experimentConfig.client.id || ""}
      experimentIds={selectedExpIds.join()}>
      {experimentConfig.engine.standard_experiment_manager_config
        .client_selection_enabled && (
        <EuiFlexItem grow={false}>
          <ClientConfigPanel
            client={experimentConfig.client}
            onChangeHandler={onChange("client")}
            errors={errors.client}
          />
        </EuiFlexItem>
      )}

      {experimentConfig.engine.standard_experiment_manager_config
        .experiment_selection_enabled && (
        <EuiFlexItem grow={false}>
          <ExperimentsConfigPanel
            experiments={experimentConfig.experiments}
            onChangeHandler={onChange("experiments")}
            errors={errors.experiments}
          />
        </EuiFlexItem>
      )}

      <EuiFlexItem grow={false}>
        <VariablesConfigPanel
          variables={experimentConfig.variables}
          onChangeHandler={onChange("variables")}
          errors={errors.variables}
        />
      </EuiFlexItem>
    </ExperimentContextProvider>
  ) : (
    <div ref={experimentSectionRef}>
      <OverlayMask parentRef={experimentSectionRef} opacity={0.4}>
        <EuiLoadingChart size="xl" mono />
      </OverlayMask>
    </div>
  );
};
