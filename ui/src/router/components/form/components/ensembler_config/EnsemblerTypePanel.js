import React from "react";
import { Panel } from "../Panel";
import { EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import { DescribedFormGroup } from "../../../../../components/form/described_form_group/DescribedFormGroup";
import { FormLabelWithToolTip } from "../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";

export const EnsemblerTypePanel = ({ type, options, onChange, errors }) => {
  const selectedOption = options.find((option) => option.value === type);

  return (
    <Panel title="Ensembler">
      <EuiForm>
        <DescribedFormGroup description={(selectedOption || {}).description}>
          <EuiFormRow
            fullWidth
            label={
              <FormLabelWithToolTip
                label="Type *"
                content="Select the type of ensembler to be used in your deployment"
              />
            }
            isInvalid={!!errors}
            error={errors}
            display="row">
            <EuiSuperSelect
              fullWidth
              options={options}
              valueOfSelected={selectedOption ? selectedOption.value : ""}
              onChange={onChange}
              itemLayoutAlign="top"
              hasDividers
              isInvalid={!!errors}
            />
          </EuiFormRow>
        </DescribedFormGroup>
      </EuiForm>
    </Panel>
  );
};
