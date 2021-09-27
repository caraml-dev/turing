import React, { useContext } from "react";
import { Panel } from "../../Panel";
import {
  EuiFieldText,
  EuiForm,
  EuiFormRow,
  EuiLink,
  EuiSpacer,
  EuiText,
} from "@elastic/eui";
import SecretsContext from "../../../../../../providers/secrets/context";
import { FormLabelWithToolTip } from "../../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { useOnChangeHandler } from "../../../../../../components/form/hooks/useOnChangeHandler";
import { ServiceAccountComboBox } from "../../../../../../components/form/service_account_combo_box/ServiceAccoutComboBox";

export const BigQueryConfigPanel = ({
  projectId,
  bigQueryConfig,
  onChangeHandler,
  errors = {},
}) => {
  const secrets = useContext(SecretsContext);

  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <Panel title="BigQuery Configuration">
      <EuiSpacer size="m" />
      <EuiForm>
        <EuiFormRow
          fullWidth
          label={
            <FormLabelWithToolTip
              label="BigQuery Table *"
              content="Specify a BQ table, that will be used for storing request/response logs"
            />
          }
          isInvalid={!!errors.table}
          error={errors.table}
          display="row"
        >
          <EuiFieldText
            fullWidth
            placeholder="project_name.dataset.table"
            value={bigQueryConfig.table || ""}
            onChange={(e) => onChange("table")(e.target.value)}
            isInvalid={!!errors.table}
            name="bigQuery-table"
          />
        </EuiFormRow>

        <EuiSpacer size="m" />

        <EuiFormRow
          fullWidth
          label="Service Account *"
          helpText={
            <EuiText size="s">
              <p>
                Choose a service account that has a write access to the
                configured BQ table.
                <br />
                You can add a new service account at Project's{" "}
                <EuiLink
                  href={`/projects/${projectId}/settings/secrets-management`}
                  target="_blank"
                  external
                >
                  Secrets Management
                </EuiLink>{" "}
                page.
              </p>
            </EuiText>
          }
          isInvalid={!!errors.service_account_secret}
          error={errors.service_account_secret}
        >
          <ServiceAccountComboBox
            fullwidth
            value={bigQueryConfig.service_account_secret}
            secrets={secrets}
            onChange={onChange("service_account_secret")}
            isInvalid={!!errors.service_account_secret}
          />
        </EuiFormRow>
      </EuiForm>
    </Panel>
  );
};
