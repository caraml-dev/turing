import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { FormContext, FormValidationContext } from "@gojek/mlp-ui";

import { DockerConfigFormGroup } from "../components/docker_config/DockerConfigFormGroup";
import { NopConfigFormGroup } from "../components/nop_config/NopConfigFormGroup";
import { PyFuncConfigFormGroup } from "../components/pyfunc_config/PyFuncConfigFormGroup";
import { StandardEnsemblerFormGroup } from "../components/ensembler_config/standard_ensembler/StandardEnsemblerFormGroup";

import { ensemblerTypeOptions } from "../components/ensembler_config/typeOptions";
import { EnsemblerTypePanel } from "../components/ensembler_config/EnsemblerTypePanel";
import { get } from "../../../../components/form/utils";
import { useOnChangeHandler } from "../../../../components/form/hooks/useOnChangeHandler";
import ExperimentEngineContext from "../../../../providers/experiments/context";

export const EnsemblerStep = ({ projectId }) => {
  const {
    data: {
      config: { experiment_engine, ensembler, routes, rules, default_traffic_rule },
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

      {ensembler.type === "nop" && (
        <NopConfigFormGroup
          routes={routes}
          rules={rules}
          default_traffic_rule={default_traffic_rule}
          nopConfig={ensembler.nop_config}
          onChangeHandler={onChange("config.ensembler.nop_config")}
          errors={get(errors, "config.ensembler.nop_config")}
        />
      )}

      {ensembler.type === "docker" && (
        <DockerConfigFormGroup
          projectId={projectId}
          dockerConfig={ensembler.docker_config}
          onChangeHandler={onChange("config.ensembler.docker_config")}
          errors={get(errors, "config.ensembler.docker_config")}
        />
      )}

      {ensembler.type === "pyfunc" && (
        <PyFuncConfigFormGroup
          projectId={projectId}
          pyfuncConfig={ensembler.pyfunc_config}
          onChangeHandler={onChange("config.ensembler.pyfunc_config")}
          errors={get(errors, "config.ensembler.pyfunc_config")}
        />
      )}

      {ensembler.type === "standard" && (
        <StandardEnsemblerFormGroup
          projectId={projectId}
          experimentEngine={experiment_engine}
          routes={routes}
          rules={rules}
          default_traffic_rule={default_traffic_rule}
          standardConfig={ensembler.standard_config}
          onChangeHandler={onChange("config.ensembler.standard_config")}
          errors={get(errors, "config.ensembler.standard_config")}
        />
      )}
    </EuiFlexGroup>
  );
};
