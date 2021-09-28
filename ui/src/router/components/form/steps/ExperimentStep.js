import React, { useContext, useEffect, useMemo } from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { EngineTypePanel } from "../components/experiment_config/EngineTypePanel";
import { FormContext } from "../../../../components/form/context";
import { get } from "../../../../components/form/utils";
import FormValidationContext from "../../../../components/form/validation";
import { useOnChangeHandler } from "../../../../components/form/hooks/useOnChangeHandler";
import ExperimentEngineContext from "../../../../providers/experiments/context";
import { ExperimentConfigGroup } from "../components/experiment_config/ExperimentConfigGroup";
import { ensemblerTypeOptions } from "../components/ensembler_config/typeOptions";
import { getExperimentEngineOptions } from "../components/experiment_config/typeOptions";

export const ExperimentStep = ({ projectId }) => {
  const {
    data: {
      config: {
        experiment_engine,
        ensembler: { type: ensemblerType },
      },
    },
    onChangeHandler,
  } = useContext(FormContext);

  const { errors } = useContext(FormValidationContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const { experimentEngines, getEngineProperties } = useContext(
    ExperimentEngineContext
  );
  const engineProps = getEngineProperties(experiment_engine.type);
  const experimentEngineOptions = useMemo(
    () => getExperimentEngineOptions(experimentEngines),
    [experimentEngines]
  );

  useEffect(() => {
    const ensemblerOptions = ensemblerTypeOptions(engineProps).filter(
      (o) => !o.disabled
    );

    const ensemblerTypeOption = ensemblerOptions.find(
      (o) => o.value === ensemblerType
    );

    if (ensemblerOptions.length && !ensemblerTypeOption) {
      onChange("config.ensembler.type")(ensemblerOptions[0].value);
    }
  }, [experiment_engine.type, engineProps, onChange, ensemblerType]);

  /* Check for xp should be replaced by a check for the default engine property. */
  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={false}>
        <EngineTypePanel
          type={experiment_engine.type}
          options={experimentEngineOptions}
          onChange={onChange("config.experiment_engine.type")}
          errors={get(errors, "config.experiment_engine.type")}
        />
      </EuiFlexItem>

      {experiment_engine.type !== "nop" && (
        <ExperimentConfigGroup
          engineType={experiment_engine.type}
          experimentConfig={experiment_engine.config}
          onChangeHandler={onChange("config.experiment_engine.config")}
          errors={get(errors, "config.experiment_engine.config")}
        />
      )}
    </EuiFlexGroup>
  );
};
