import React, { Fragment } from "react";
import { useToggle } from "@caraml-dev/ui-lib";
import {
  EuiForm,
  EuiFormRow,
  EuiFieldNumber,
  EuiIcon,
  EuiLink,
  EuiSpacer,
  EuiSuperSelect,
} from "@elastic/eui";

import { Panel } from "../Panel";
import { DescribedFormGroup } from "../../../../../components/form/described_form_group/DescribedFormGroup";
import { FormLabelWithToolTip } from "../../../../../components/form/label_with_tooltip/FormLabelWithToolTip";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";

import { AutoscalingPolicyPanelFlyout } from "./AutoscalingPolicyPanelFlyout";
import { autoscalingPolicyOptions } from "./typeOptions";

export const AutoscalingPolicyPanel = ({
  autoscalingPolicyConfig,
  onChangeHandler,
  errors = {},
}) => {
  const { onChange } = useOnChangeHandler(onChangeHandler);
  // Parse the float portion of the value
  const onChangeTarget = (value) => {
    const parsedFloat = parseFloat(value);
    onChange("target")(isNaN(value) ? "" : parsedFloat.toString());
  };
  // Update default target if the metric is changing
  const onChangeMetric = (value) => {
    const newSelection = autoscalingPolicyOptions.find((e) => e.value === value);
    onChange("metric")(value);
    onChangeTarget(newSelection.defaultValue);
  };

  const selectedMetric = autoscalingPolicyOptions.find((e) => e.value === autoscalingPolicyConfig.metric);
  const [isFlyoutVisible, toggleIsFlyoutVisible] = useToggle();

  return (
    <Panel title={
      <Fragment>
        Autoscaling Policy{" "}
        <EuiLink color="ghost" onClick={() => toggleIsFlyoutVisible()}>
          <EuiIcon
            type="questionInCircle"
            color="subdued"
            className="eui-alignBaseline"
          />
        </EuiLink>
      </Fragment>
    }>
      <EuiForm>
        <DescribedFormGroup description={selectedMetric?.description || ""}>
          <EuiFormRow
            label={
              <FormLabelWithToolTip
                label="Metric Type*"
                content="Specify the autoscaling metric to monitor"
              />
            }
            isInvalid={!!errors.metric}
            error={errors.metric}
            fullWidth>
            <EuiSuperSelect
              fullWidth
              options={autoscalingPolicyOptions}
              valueOfSelected={selectedMetric?.value || ""}
              onChange={onChangeMetric}
              itemLayoutAlign="top"
              hasDividers
              isInvalid={!!errors.metric}
            />
          </EuiFormRow>

          <EuiSpacer size="l" />

          <EuiFormRow
            label={
              <FormLabelWithToolTip
                label="Target*"
                content="Specify the target value after which autoscaling should be triggered"
              />
            }
            isInvalid={!!errors.target}
            error={errors.target}
            fullWidth>
            <EuiFieldNumber
              placeholder="Enter a number. Eg: 20"
              value={autoscalingPolicyConfig.target}
              onChange={(e) => onChangeTarget(e.target.value)}
              isInvalid={!!errors.target}
              name="memory"
              // The min value is set as 0.005 because it's the smallest value, when rounded to 2 decimal places, gives
              // 0.01, the smallest value accepted as an autoscaling target (concurrency).
              min={0.005}
              step={"any"}
              append={selectedMetric.unit}
            />
          </EuiFormRow>
        </DescribedFormGroup>
      </EuiForm>

      {isFlyoutVisible && (
        <AutoscalingPolicyPanelFlyout onClose={() => toggleIsFlyoutVisible()} />
      )}
    </Panel>
  );
};
