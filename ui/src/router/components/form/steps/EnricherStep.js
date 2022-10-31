import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import React, { Fragment, useContext } from "react";
import { FormContext, FormValidationContext } from "@gojek/mlp-ui";
import { EnvVariablesPanel } from "../components/docker_config/EnvVariablesPanel";
import { EnricherTypePanel } from "../components/enricher_config/EnricherTypePanel";
import { DockerDeploymentPanel } from "../components/docker_config/DockerDeploymentPanel";
import { DockerRegistriesContextProvider } from "../../../../providers/docker/context";
import { useConfig } from "../../../../config";
import { get } from "../../../../components/form/utils";
import { useOnChangeHandler } from "../../../../components/form/hooks/useOnChangeHandler";
import { enricherTypeOptions } from "../components/enricher_config/typeOptions";
import { SecretsContextProvider } from "../../../../providers/secrets/context";
import { ResourcesRequirementsPanel } from "../components/ResourcesRequirementsPanel";

export const EnricherStep = ({ projectId }) => {
  const {
    appConfig: {
      scaling: { maxAllowedReplica },
    },
  } = useConfig();
  const {
    data: {
      config: { enricher, protocol },
    },
    onChangeHandler,
  } = useContext(FormContext);

  const { errors } = useContext(FormValidationContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <EuiFlexGroup direction="column" gutterSize="m">
      <EuiFlexItem>
        <EnricherTypePanel
          type={enricher.type}
          options={enricherTypeOptions[protocol]}
          onChange={onChange("config.enricher.type")}
          errors={get(errors, "config.enricher.type")}
        />
      </EuiFlexItem>

      {enricher.type === "docker" && (
        <Fragment>
          <EuiFlexItem>
            <SecretsContextProvider projectId={projectId}>
              <DockerRegistriesContextProvider>
                <DockerDeploymentPanel
                  projectId={projectId}
                  values={enricher}
                  onChangeHandler={onChange("config.enricher")}
                  errors={get(errors, "config.enricher")}
                />
              </DockerRegistriesContextProvider>
            </SecretsContextProvider>
          </EuiFlexItem>

          <EuiFlexItem>
            <EnvVariablesPanel
              variables={enricher.env}
              onChangeHandler={onChange("config.enricher.env")}
              errors={get(errors, "config.enricher.env")}
            />
          </EuiFlexItem>

          <ResourcesRequirementsPanel
            resources={enricher.resource_request}
            resourcesOnChangeHandler={onChange("config.enricher.resource_request")}
            resourcesErrors={get(errors, "config.enricher.resource_request")}
            autoscalingPolicy={enricher.autoscaling_policy}
            autoscalingPolicyOnChangeHandler={onChange("config.enricher.autoscaling_policy")}
            autoscalingPolicyErrors={get(errors, "config.enricher.autoscaling_policy")}
            maxAllowedReplica={maxAllowedReplica}
          />
        </Fragment>
      )}
    </EuiFlexGroup>
  );
};
