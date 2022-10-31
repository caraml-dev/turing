import React, { Fragment, useEffect } from "react";
import { EuiFlexItem } from "@elastic/eui";
import { useConfig } from "../../../../../config";
import { DockerDeploymentPanel } from "./DockerDeploymentPanel";
import { DockerEnsembler } from "../../../../../services/ensembler";
import { DockerRegistriesContextProvider } from "../../../../../providers/docker/context";
import { EnvVariablesPanel } from "./EnvVariablesPanel";
import { SecretsContextProvider } from "../../../../../providers/secrets/context";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";
import { ResourcesRequirementsPanel } from "../ResourcesRequirementsPanel";

export const DockerConfigFormGroup = ({
  projectId,
  dockerConfig,
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
    !dockerConfig && onChangeHandler(DockerEnsembler.newConfig());
  }, [dockerConfig, onChangeHandler]);

  return (
    !!dockerConfig && (
      <Fragment>
        <EuiFlexItem>
          <SecretsContextProvider projectId={projectId}>
            <DockerRegistriesContextProvider>
              <DockerDeploymentPanel
                projectId={projectId}
                values={dockerConfig}
                onChangeHandler={onChangeHandler}
                errors={errors}
              />
            </DockerRegistriesContextProvider>
          </SecretsContextProvider>
        </EuiFlexItem>

        <EuiFlexItem>
          <EnvVariablesPanel
            variables={dockerConfig.env}
            onChangeHandler={onChange("env")}
            errors={errors.env}
          />
        </EuiFlexItem>

        <ResourcesRequirementsPanel
          resources={dockerConfig.resource_request}
          resourcesOnChangeHandler={onChange("resource_request")}
          resourcesErrors={errors.resource_request}
          autoscalingPolicy={dockerConfig.autoscaling_policy}
          autoscalingPolicyOnChangeHandler={onChange("autoscaling_policy")}
          autoscalingPolicyErrors={errors.autoscaling_policy}
          maxAllowedReplica={maxAllowedReplica}
        />
      </Fragment>
    )
  );
};
