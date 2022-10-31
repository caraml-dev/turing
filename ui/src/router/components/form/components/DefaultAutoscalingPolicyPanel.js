import React from "react";
import { Panel } from "./Panel";
import {
  EuiFieldText,
  EuiForm,
  EuiFormRow,
  EuiSpacer,
  EuiSwitch,
} from "@elastic/eui";
import { FormLabelWithToolTip } from "../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { useOnChangeHandler } from "../../../../components/form/hooks/useOnChangeHandler";
import { DescribedFormGroup } from "@gojek/mlp-ui";

export const DefaultAutoscalingPolicyPanel = ({
  useDefaultAutoscalingPolicy,
  toggleUseDefaultAutoscalingPolicy,
  autoscalingPolicyConfig,
  onChangeHandler,
  errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);

  return (
    <Panel title="Resources Requirements">
      <EuiForm>
        <EuiSpacer size={"m"}/>

        <EuiSwitch
          label={useDefaultAutoscalingPolicy ?
            "Use default autoscaling policy"
            : "Customise resources requirements and autoscaling manually"}
          checked={useDefaultAutoscalingPolicy}
          onChange={toggleUseDefaultAutoscalingPolicy}
        />

        {useDefaultAutoscalingPolicy &&
          <>
            <EuiSpacer size={"m"}/>

            <DescribedFormGroup description={
              "The value specified here will be used to perform autoscaling automatically. There is no need to specify " +
              "and resource requests values or custom autoscaling values."
            }>
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
                  fullWidth
                />
              </EuiFormRow>
            </DescribedFormGroup>
          </>
        }
      </EuiForm>
    </Panel>
  );
};
