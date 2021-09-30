import React, { useContext } from "react";
import {
  EuiAccordion,
  EuiFieldText,
  EuiFieldNumber,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer,
  EuiText,
  EuiLink,
} from "@elastic/eui";
import { Panel } from "../Panel";
import { SelectDockerImageComboBox } from "../../../../../components/form/docker_image_combo_box/SelectDockerImageComboBox";
import { EuiFieldDuration } from "../../../../../components/form/field_duration/EuiFieldDuration";
import SecretsContext from "../../../../../providers/secrets/context";
import DockerRegistriesContext from "../../../../../providers/docker/context";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";
import { ServiceAccountComboBox } from "../../../../../components/form/service_account_combo_box/ServiceAccoutComboBox";
import { appConfig } from "../../../../../config";

const imageOptions = [];

export const DockerDeploymentPanel = ({
  projectId,
  values: { image, port = 0, endpoint, timeout, service_account },
  onChangeHandler,
  errors = {},
}) => {
  const secrets = useContext(SecretsContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);
  const registries = useContext(DockerRegistriesContext);

  return (
    <Panel title="Deployment">
      <EuiForm>
        <EuiFormRow
          label="Docker Image *"
          isInvalid={!!errors.image}
          error={errors.image}
          fullWidth
          display="row">
          <SelectDockerImageComboBox
            fullWidth
            value={image || `${appConfig.defaultDockerRegistry}/`}
            placeholder="echo:1.0.2"
            registryOptions={registries}
            imageOptions={imageOptions}
            onChange={onChange("image")}
            isInvalid={!!errors.image}
          />
        </EuiFormRow>

        <EuiSpacer size="m" />

        <EuiFlexGroup direction="row">
          <EuiFlexItem grow={2}>
            <EuiFormRow
              label="Endpoint *"
              isInvalid={!!errors.endpoint}
              error={errors.endpoint}>
              <EuiFieldText
                fullWidth
                placeholder="/preprocess"
                value={endpoint}
                onChange={(e) => onChange("endpoint")(e.target.value)}
                isInvalid={!!errors.endpoint}
              />
            </EuiFormRow>
          </EuiFlexItem>

          <EuiFlexItem grow={1}>
            <EuiFormRow
              label="Port *"
              isInvalid={!!errors.port}
              error={errors.port}>
              <EuiFieldNumber
                fullWidth
                min={0}
                max={65535}
                placeholder="8080"
                value={port}
                onChange={(e) => {
                  let port = parseInt(e.target.value);
                  onChange("port")(isNaN(port) ? undefined : port);
                }}
                isInvalid={!!errors.port}
              />
            </EuiFormRow>
          </EuiFlexItem>
          <EuiFlexItem grow={1}>
            <EuiFormRow
              label="Timeout *"
              isInvalid={!!errors.timeout}
              error={errors.timeout}>
              <EuiFieldDuration
                fullWidth
                placeholder="100"
                value={timeout}
                onChange={onChange("timeout")}
                isInvalid={!!errors.timeout}
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>

        <EuiSpacer size="l" />

        <EuiAccordion
          id="euiAccordion--serviceAccount"
          initialIsOpen={!!service_account}
          buttonContent="+ Service Account"
          paddingSize="m">
          <EuiFormRow
            fullWidth
            label="Service Account"
            helpText={
              <EuiText size="s">
                <p>
                  Choose a service account that you want to use inside of the
                  Docker container.
                  <br />
                  You can add a new service account at Project's{" "}
                  <EuiLink
                    href={`/projects/${projectId}/settings/secrets-management`}
                    target="_blank"
                    external>
                    Secrets Management
                  </EuiLink>{" "}
                  page.
                </p>
              </EuiText>
            }
            isInvalid={!!errors.service_account}
            error={errors.service_account}>
            <ServiceAccountComboBox
              fullwidth
              value={service_account}
              secrets={secrets}
              onChange={onChange("service_account")}
              isInvalid={!!errors.service_account}
            />
          </EuiFormRow>
        </EuiAccordion>
      </EuiForm>
    </Panel>
  );
};
