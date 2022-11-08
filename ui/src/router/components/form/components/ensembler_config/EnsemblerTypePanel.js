import React from "react";
import { Panel } from "../Panel";
import { EuiCheckbox, EuiForm, EuiFormRow, EuiSuperSelect } from "@elastic/eui";
import { DescribedFormGroup } from "../../../../../components/form/described_form_group/DescribedFormGroup";
import { FormLabelWithToolTip } from "../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { get } from "../../../../../components/form/utils";
import { useOnChangeHandler } from "@gojek/mlp-ui";

export const EnsemblerTypePanel = ({ type, lazyRouting, typeOptions, onChange, errors }) => {
  const selectedOption = typeOptions.find((option) => option.value === type);
  const onChangeHandler = useOnChangeHandler(onChange);

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
            isInvalid={!!get(errors, "type")}
            error={get(errors, "type")}
            display="row">
            <EuiSuperSelect
              fullWidth
              options={typeOptions}
              valueOfSelected={selectedOption ? selectedOption.value : ""}
              onChange={onChangeHandler("type")}
              itemLayoutAlign="top"
              hasDividers
              isInvalid={!!get(errors, "type")}
            />
          </EuiFormRow>
        </DescribedFormGroup>
        <EuiFormRow
          fullWidth
          isInvalid={!!get(errors, "lazy_routing")}
          error={get(errors, "lazy_routing")}
          display="row">
          <EuiCheckbox
            id="lazy-routing-checkbox"
            label={
              <FormLabelWithToolTip
                label="Do Lazy Routing"
                content="Lazy Routing will call the experiment engine first and then invoke only the required route, based on the generated treatment."
              />
            }
            checked={lazyRouting}
            onChange={onChangeHandler("lazy_routing")}
          />
        </EuiFormRow>
      </EuiForm>
    </Panel>
  );
};
