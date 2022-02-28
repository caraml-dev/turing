import React, { useContext } from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
  EuiSpacer,
  EuiSuperSelect,
} from "@elastic/eui";
import { Panel } from "../Panel";
import { EuiFieldDuration } from "../../../../../components/form/field_duration/EuiFieldDuration";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";
import { FormLabelWithToolTip } from "../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import EnsemblersContext from "../../../../../providers/ensemblers/context";

export const PyFuncDeploymentPanel = ({
  values: { ensembler_id, timeout },
  onChangeHandler,
  errors = {},
}) => {
  const { ensemblers } = useContext(EnsemblersContext);

  let options = Object.values(ensemblers).reduce((pyfunc_ensemblers, val) => {
    pyfunc_ensemblers.push({
      value: val["id"],
      inputDisplay: val["name"],
    });
    return pyfunc_ensemblers;
  }, []);

  const { onChange } = useOnChangeHandler(onChangeHandler);
  const selectedOption = options.find(
    (option) => option.value === ensembler_id
  );

  return (
    <Panel title="Deployment">
      <EuiForm>
        <EuiFlexGroup direction="row">
          <EuiFlexItem grow={2}>
            <EuiFormRow
              fullWidth
              label={
                <FormLabelWithToolTip
                  label="Pyfunc Ensembler *"
                  content="Select the Pyfunc Ensembler to be used in your deployment"
                />
              }
              // isInvalid={!!errors.pyfunc_config}
              // error={errors}
              display="row">
              <EuiSuperSelect
                fullWidth
                options={options}
                valueOfSelected={selectedOption ? selectedOption.value : ""}
                onChange={onChange("ensembler_id")}
                itemLayoutAlign="top"
                hasDividers
                // isInvalid={!!errors}
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
      </EuiForm>
    </Panel>
  );
};
