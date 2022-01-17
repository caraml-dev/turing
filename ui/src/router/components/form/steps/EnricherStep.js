import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { ResourcesPanel } from "../components/ResourcesPanel";
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

export const EnricherStep = ({ projectId }) => {
  const {
    appConfig: {
      scaling: { maxAllowedReplica },
    },
  } = useConfig();
  const {
    data: {
      config: { enricher },
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
          options={enricherTypeOptions}
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

          <EuiFlexItem>
            <ResourcesPanel
              resourcesConfig={enricher.resource_request}
              onChangeHandler={onChange("config.enricher.resource_request")}
              maxAllowedReplica={maxAllowedReplica}
              errors={get(errors, "config.enricher.resource_request")}
            />
          </EuiFlexItem>
        </Fragment>
      )}
    </EuiFlexGroup>
  );
};
