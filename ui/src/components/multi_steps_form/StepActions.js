import React from "react";
import {
  EuiButton,
  EuiButtonEmpty,
  EuiFlexGroup,
  EuiFlexItem
} from "@elastic/eui";

export const StepActions = ({
  currentStep = 0,
  onCancel,
  onPrevious,
  onSubmit,
  submitLabel,
  isSubmitting
}) => {
  return (
    <EuiFlexGroup direction="row" justifyContent="flexEnd">
      <EuiFlexItem grow={false}>
        {currentStep === 0 ? (
          <EuiButtonEmpty size="s" color="primary" onClick={onCancel}>
            Cancel
          </EuiButtonEmpty>
        ) : (
          <EuiButton size="s" color="primary" onClick={onPrevious}>
            Previous
          </EuiButton>
        )}
      </EuiFlexItem>

      <EuiFlexItem grow={false}>
        <EuiButton
          size="s"
          color="primary"
          fill
          isLoading={isSubmitting}
          onClick={onSubmit}>
          {submitLabel}
        </EuiButton>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
