import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { DockerConfigFormGroup } from "../components/docker_config/DockerConfigFormGroup";
import { ensemblerTypeOptions } from "../components/ensembler_config/typeOptions";
import { EnsemblerTypePanel } from "../components/ensembler_config/EnsemblerTypePanel";
import { FormContext, FormValidationContext } from "@gojek/mlp-ui";
import { get } from "../../../../components/form/utils";
import { StandardEnsemblerFormGroup } from "../components/ensembler_config/standard_ensembler/StandardEnsemblerFormGroup";
import { useOnChangeHandler } from "../../../../components/form/hooks/useOnChangeHandler";
import ExperimentEngineContext from "../../../../providers/experiments/context";

export const EnsemblerStep = ({ projectId }) => {
  const {
    data: {
      config: { experiment_engine, ensembler, routes },
    },
    onChangeHandler,
  } = useContext(FormContext);
  const { errors } = useContext(FormValidationContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const { getEngineProperties } = useContext(ExperimentEngineContext);
  const engineProps =
    experiment_engine.type === "nop"
      ? {}
      : getEngineProperties(experiment_engine.type);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem>
        <EnsemblerTypePanel
          type={ensembler.type}
          options={ensemblerTypeOptions(engineProps)}
          onChange={onChange("config.ensembler.type")}
          errors={get(errors, "config.ensembler.type")}
        />
      </EuiFlexItem>

      {ensembler.type === "docker" && (
        <DockerConfigFormGroup
          projectId={projectId}
          dockerConfig={ensembler.docker_config}
          onChangeHandler={onChange("config.ensembler.docker_config")}
          errors={get(errors, "config.ensembler.docker_config")}
        />
      )}

      {ensembler.type === "standard" && (
        <StandardEnsemblerFormGroup
          experimentConfig={experiment_engine.config}
          routes={routes}
          standardConfig={ensembler.standard_config}
          onChangeHandler={onChange("config.ensembler.standard_config")}
          errors={get(errors, "config.ensembler.standard_config")}
        />
      )}
    </EuiFlexGroup>
  );
};
