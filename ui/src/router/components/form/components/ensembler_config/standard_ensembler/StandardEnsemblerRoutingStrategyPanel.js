import React from "react";
import { EuiFlexItem, EuiForm, EuiFormRow, EuiSpacer, EuiSuperSelect } from "@elastic/eui";
import { DescribedFormGroup } from "../../../../../../components/form/described_form_group/DescribedFormGroup";
import { Panel } from "../../Panel";
import { routingStrategyOptions } from "./typeOptions";

export const StandardEnsemblerRoutingStrategyPanel = ({
  isLazyRouting,
  onChange,
  errors,
}) => {
  const selectedOption = routingStrategyOptions.find(e => e.flag === Boolean(isLazyRouting).toString());
  const onChangeRoutingStrategy = value => {
    const opt = routingStrategyOptions.find(e => e.value === value);
    // Save the boolean value
    onChange(opt?.flag === "true");
  };

  return (
    <EuiFlexItem>
      <Panel title="Routing Strategy">
        <EuiForm>
          <EuiFormRow
            isInvalid={!!errors}
            error={errors}
            fullWidth
            display="row">
            <DescribedFormGroup description={selectedOption?.description || ""}>
              <EuiFormRow
                fullWidth
                isInvalid={!!errors}
                error={errors}
                display="row">
                <>
                  <EuiSpacer />
                  <EuiSuperSelect
                    fullWidth
                    options={routingStrategyOptions}
                    valueOfSelected={selectedOption?.value || ""}
                    onChange={onChangeRoutingStrategy}
                    itemLayoutAlign="top"
                    hasDividers
                    isInvalid={!!errors?.isLazyRouting}
                  />
                </>
              </EuiFormRow>
            </DescribedFormGroup>
          </EuiFormRow>
        </EuiForm>
      </Panel>
    </EuiFlexItem>
  );
}
