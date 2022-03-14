import React, { Fragment, useEffect } from "react";
import { EuiFlexItem } from "@elastic/eui";
import { useConfig } from "../../../../../config";
import { ResourcesPanel } from "../ResourcesPanel";
import { SecretsContextProvider } from "../../../../../providers/secrets/context";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";
import { PyFuncEnsembler } from "../../../../../services/ensembler";
import { PyFuncDeploymentPanel } from "./PyFuncDeploymentPanel";
import { EnsemblersContextContextProvider } from "../../../../../providers/ensemblers/context";

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
          <ResourcesPanel
            resourcesConfig={pyfuncConfig.resource_request}
            onChangeHandler={onChange("resource_request")}
            errors={errors.resource_request}
            maxAllowedReplica={maxAllowedReplica}
          />
        </EuiFlexItem>
      </Fragment>
    )
  );
};
