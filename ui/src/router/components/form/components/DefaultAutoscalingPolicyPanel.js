import React from "react";
import { Panel } from "./Panel";
import {
  EuiFieldText,
  EuiFlexGroup,
  EuiFlexItem,
  EuiForm,
  EuiFormRow,
} from "@elastic/eui";
import { FormLabelWithToolTip } from "../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { useOnChangeHandler } from "../../../../components/form/hooks/useOnChangeHandler";


export const DefaultAutoscalingPolicyPanel = ({
  autoscalingPolicyConfig,
  onChangeHandler,
  errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <Panel title="Default Autoscaling Policy">
      <EuiForm>
        <EuiFlexGroup direction="row">
          <EuiFlexItem>
            <EuiFormRow
              label={
                <FormLabelWithToolTip
                  label="Payload Size*"
                  content="Specify the estimated payload size to be received by this component"
                />
              }
              isInvalid={!!errors.payload_size}
              error={errors.payload_size}
              fullWidth>
              <EuiFieldText
                placeholder="500Mi"
                value={autoscalingPolicyConfig.payload_size}
                onChange={(e) => onChange("payload_size")(e.target.value)}
                isInvalid={!!errors.payload_size}
                name="payload size"
              />
            </EuiFormRow>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiForm>
    </Panel>
  );
};
