import React, { useState } from "react";
import { Panel } from "../Panel";
import { DescribedFormGroup } from "../../../../../components/form/described_form_group/DescribedFormGroup";
import { EuiForm, EuiFormRow, EuiSuperSelect, EuiSwitch } from "@elastic/eui";
import { FormLabelWithToolTip } from "../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";

export const ResultLoggingTypePanel = ({
  type,
  options,
  onChange,
  errors,
  protocol,
}) => {
  const selectedOption = options.find((option) => option.value === type);

  const [isUPILogging, setIsUPILogging] = useState(type === "upi");

  const setUPILogger = (e) => {
    setIsUPILogging(e.target.checked);
    if (!isUPILogging) {
      onChange("upi");
    } else {
      onChange("nop");
    }
    setIsUPILogging(!isUPILogging);
  };

  return (
    <Panel title="General">
      {protocol === "UPI_V1" && (
        <EuiSwitch
          label="Enable Logging"
          checked={isUPILogging}
          onChange={setUPILogger}
        />
      )}
      {protocol !== "UPI_V1" && (
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
              display="row"
            >
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
      )}
    </Panel>
  );
};
