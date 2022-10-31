import React, { Fragment, useEffect } from "react";
import { EuiFlexItem } from "@elastic/eui";
import { useConfig } from "../../../../../config";
import { SecretsContextProvider } from "../../../../../providers/secrets/context";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";
import { PyFuncEnsembler } from "../../../../../services/ensembler";
import { PyFuncDeploymentPanel } from "./PyFuncDeploymentPanel";
import { EnsemblersContextProvider } from "../../../../../providers/ensemblers/context";
import { EnvVariablesPanel } from "../docker_config/EnvVariablesPanel";
import { ResourcesRequirementsPanel } from "../ResourcesRequirementsPanel";

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
            <EnsemblersContextProvider
              projectId={projectId}
              ensemblerType={"pyfunc"}>
              <PyFuncDeploymentPanel
                values={pyfuncConfig}
                onChangeHandler={onChangeHandler}
                errors={errors}
              />
            </EnsemblersContextProvider>
          </SecretsContextProvider>
        </EuiFlexItem>

        <EuiFlexItem>
          <EnvVariablesPanel
            variables={pyfuncConfig.env}
            onChangeHandler={onChange("env")}
            errors={errors.env}
          />
        </EuiFlexItem>

        <ResourcesRequirementsPanel
          resources={pyfuncConfig.resource_request}
          resourcesOnChangeHandler={onChange("resource_request")}
          resourcesErrors={errors.resource_request}
          autoscalingPolicy={pyfuncConfig.autoscaling_policy}
          autoscalingPolicyOnChangeHandler={onChange("autoscaling_policy")}
          autoscalingPolicyErrors={errors.autoscaling_policy}
          maxAllowedReplica={maxAllowedReplica}
        />
      </Fragment>
    )
  );
};
