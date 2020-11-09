import React from "react";
import { Panel } from "../Panel";
import { DescribedFormGroup } from "../../../../../components/form/described_form_group/DescribedFormGroup";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import { FormLabelWithToolTip } from "../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";

export const ResultLoggingTypePanel = ({ type, options, onChange, errors }) => {
  const selectedOption = options.find(option => option.value === type);

  return (
    <Panel title="General">
      <EuiForm>
        <DescribedFormGroup description={(selectedOption || {}).description}>
          <EuiFormRow
            fullWidth
            label={
              <FormLabelWithToolTip
                label="Results Logging *"
                content="Select the destination for request/response log data"
              />
            }
            isInvalid={!!errors}
            error={errors}
            display="row">
            <EuiSuperSelect
              fullWidth
              options={options}
              valueOfSelected={(selectedOption || {}).value}
              onChange={onChange}
              isInvalid={!!errors}
              itemLayoutAlign="top"
              hasDividers
            />
          </EuiFormRow>
        </DescribedFormGroup>
      </EuiForm>
    </Panel>
  );
};
