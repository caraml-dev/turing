import React, { Fragment, useEffect } from "react";
import { EuiFlexItem } from "@elastic/eui";
import { useConfig } from "../../../../../config";
import { DockerDeploymentPanel } from "./DockerDeploymentPanel";
import { DockerEnsembler } from "../../../../../services/ensembler";
import { DockerRegistriesContextProvider } from "../../../../../providers/docker/context";
import { EnvVariablesPanel } from "./EnvVariablesPanel";
import { ResourcesPanel } from "../ResourcesPanel";
import { SecretsContextProvider } from "../../../../../providers/secrets/context";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";
import { AutoscalingPolicyPanel } from "../autoscaling_policy/AutoscalingPolicyPanel";

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

        <EuiFlexItem>
          <ResourcesPanel
            resourcesConfig={dockerConfig.resource_request}
            onChangeHandler={onChange("resource_request")}
            errors={errors.resource_request}
            maxAllowedReplica={maxAllowedReplica}
          />
        </EuiFlexItem>

        <EuiFlexItem>
          <AutoscalingPolicyPanel
            autoscalingPolicyConfig={dockerConfig.autoscaling_policy}
            onChangeHandler={onChange("autoscaling_policy")}
            errors={errors.autoscaling_policy}
          />
        </EuiFlexItem>
      </Fragment>
    )
  );
};
