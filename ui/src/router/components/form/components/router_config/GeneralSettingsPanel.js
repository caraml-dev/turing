import React, { useContext } from "react";
import {
  EuiForm,
  EuiFormRow,
  EuiSuperSelect,
  EuiFieldText,
  EuiFlexItem,
  EuiFlexGroup,
  EuiSpacer,
} from "@elastic/eui";
import { Panel } from "../Panel";
import EnvironmentsContext from "../../../../../providers/environments/context";
import { EuiFieldDuration } from "../../../../../components/form/field_duration/EuiFieldDuration";
import { get } from "../../../../../components/form/utils";
import { FormLabelWithToolTip } from "../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import sortBy from "lodash/sortBy";
import { protocolTypeOptions } from "./typeOptions";
import { newRoute } from "../../../../../services/router/TuringRouter"

export const GeneralSettingsPanel = ({
  name,
  environment,
  timeout,
  protocol,
  isEdit,
  onChange,
  errors = {},
}) => {
  const environments = useContext(EnvironmentsContext);

  const environmentOptions = sortBy(environments, "name").map(
    (environment) => ({
      value: environment.name,
      inputDisplay: environment.name,
    })
  );

  return (
    <Panel title="General">
      <EuiForm>
        <EuiFormRow
          fullWidth
          label={
            <FormLabelWithToolTip
              label="Environment *"
              content="Specify the target environment your Turing router be deployed to"
            />
          }
          isInvalid={!!errors.environment_name}
          error={errors.environment_name}
          display="row">
          <EuiSuperSelect
            fullWidth
            options={environmentOptions}
            valueOfSelected={environment}
            onChange={onChange("environment_name")}
            isInvalid={!!errors.environment_name}
            disabled={isEdit}
            itemLayoutAlign="top"
            hasDividers
          />
        </EuiFormRow>

        <EuiSpacer size="m" />

        <EuiFormRow
          fullWidth
          label={
            <FormLabelWithToolTip
              label="Protocol *"
              content="Specify the type of Turing router Protocol"
            />
          }
          display="row">
          <EuiSuperSelect
            fullWidth
            options={protocolTypeOptions}
            valueOfSelected={protocol}
            onChange={(e)=>{
              onChange("config.protocol")(e)
              // reset routes when protocol changes
              onChange("config.routes")([newRoute()])
            }}
            hasDividers
          />
        </EuiFormRow>

        <EuiSpacer size="m" />

        <EuiFlexGroup direction="row">
          <EuiFlexItem>
            <EuiFormRow
              fullWidth
              label={
                <FormLabelWithToolTip
                  label="Name *"
                  content="Specify the name for your Turing deployment"
                />
              }
              isInvalid={!!errors.name}
              error={errors.name}
              display="row">
              <EuiFieldText
                fullWidth
                placeholder="deployment-name"
                value={name}
                onChange={(e) => onChange("name")(e.target.value)}
                isInvalid={!!errors.name}
                disabled={isEdit}
                name="router-name"
              />
            </EuiFormRow>
          </EuiFlexItem>

          <EuiFlexItem>
            <EuiFormRow
              fullWidth
              label={
                <FormLabelWithToolTip
                  label="Timeout *"
                  content="Specify the overall timeout, after exceeding of which request execution by Turing router should be terminated"
                />
              }
              isInvalid={!!get(errors, "config.timeout")}
              error={get(errors, "config.timeout")}
              display="row">
              <EuiFieldDuration
                fullWidth
                placeholder="100"
                value={timeout}
                onChange={onChange("config.timeout")}
                isInvalid={!!get(errors, "config.timeout")}
                name="timeout"
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiForm>
    </Panel>
  );
};
