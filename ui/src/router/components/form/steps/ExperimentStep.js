import React, { useContext, useEffect, useMemo } from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { EngineTypePanel } from "../components/experiment_config/EngineTypePanel";
import { FormContext, FormValidationContext } from "@gojek/mlp-ui";
import { get } from "../../../../components/form/utils";
import { useOnChangeHandler } from "../../../../components/form/hooks/useOnChangeHandler";
import ExperimentEngineContext from "../../../../providers/experiments/context";
import { ExperimentConfigPanel } from "../components/experiment_config/ExperimentConfigPanel";
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
  const onChangeEngineType = (newType) => {
    // Reset the experiment config if the engine type changed
    onChange("config.experiment_engine")({ type: newType });
    // Reset the standard ensembler config if the engine type is changed
    if (ensemblerType === "standard") {
      // Standard Ensemblers are not available when no experiment engine is configured
      if (newType === "nop") {
        onChange("config.ensembler.type")("nop");
      }
      onChange("config.ensembler.standard_config.experiment_mappings")([]);
      onChange("config.ensembler.standard_config.route_name_path")("");
    }
  };

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

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem grow={false}>
        <EngineTypePanel
          type={experiment_engine.type}
          options={experimentEngineOptions}
          onChange={onChangeEngineType}
          errors={get(errors, "config.experiment_engine.type")}
        />
      </EuiFlexItem>

      {experiment_engine.type !== "nop" && (
        <ExperimentConfigPanel
          projectId={projectId}
          engine={experiment_engine}
          onChangeHandler={onChange("config.experiment_engine.config")}
          errors={get(errors, "config.experiment_engine.config")}
        />
      )}
    </EuiFlexGroup>
  );
};
