import React, { Fragment, useEffect } from "react";
import { EuiFlexItem } from "@elastic/eui";
import { useConfig } from "../../../../../config";
import { ResourcesPanel } from "../ResourcesPanel";
import { SecretsContextProvider } from "../../../../../providers/secrets/context";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";
import { PyFuncEnsembler } from "../../../../../services/ensembler";
import { PyFuncDeploymentPanel } from "./PyFuncDeploymentPanel";
import { EnsemblersContextContextProvider } from "../../../../../providers/ensemblers/context";
import { EnvVariablesPanel } from "../docker_config/EnvVariablesPanel";
import { AutoscalingPolicyPanel } from "../autoscaling_policy/AutoscalingPolicyPanel";

export const PyFuncConfigFormGroup = ({
  projectId,
  pyfuncConfig,
  onChangeHandler,
  errors = {},
}) => {
  const {
    appConfig: {
      scaling: { maxAllowedReplica },
    },
  } = useConfig();
  const { onChange } = useOnChangeHandler(onChangeHandler);

  useEffect(() => {
    !pyfuncConfig &&
      onChangeHandler(PyFuncEnsembler.newConfig(parseInt(projectId)));
  }, [pyfuncConfig, onChangeHandler, projectId]);

  return (
    !!pyfuncConfig && (
      <Fragment>
        <EuiFlexItem>
          <SecretsContextProvider projectId={projectId}>
            <EnsemblersContextContextProvider
              projectId={projectId}
              ensemblerType={"pyfunc"}>
              <PyFuncDeploymentPanel
                values={pyfuncConfig}
                onChangeHandler={onChangeHandler}
                errors={errors}
              />
            </EnsemblersContextContextProvider>
          </SecretsContextProvider>
        </EuiFlexItem>

        <EuiFlexItem>
          <EnvVariablesPanel
            variables={pyfuncConfig.env}
            onChangeHandler={onChange("env")}
            errors={errors.env}
          />
        </EuiFlexItem>

        <EuiFlexItem>
          <ResourcesPanel
            resourcesConfig={pyfuncConfig.resource_request}
            onChangeHandler={onChange("resource_request")}
            errors={errors.resource_request}
            maxAllowedReplica={maxAllowedReplica}
          />
        </EuiFlexItem>

        <EuiFlexItem>
          <AutoscalingPolicyPanel
            autoscalingPolicyConfig={pyfuncConfig.autoscaling_policy}
            onChangeHandler={onChange("autoscaling_policy")}
            errors={errors.autoscaling_policy}
          />
        </EuiFlexItem>
      </Fragment>
    )
  );
};
