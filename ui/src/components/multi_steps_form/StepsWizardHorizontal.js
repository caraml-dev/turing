import React, { useState } from "react";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiSpacer,
  EuiStepsHorizontal
} from "@elastic/eui";
import { StepContent } from "./StepContent";
import { StepActions } from "./StepActions";
import FormValidationContext, {
  FormValidationContextProvider
} from "../form/validation";

export const StepsWizardHorizontal = ({
  steps,
  onCancel,
  onSubmit,
  submitLabel = "Done"
}) => {
  const [currentStep, setCurrentStep] = useState(0);
  const isLastStep = currentStep === steps.length - 1;

  const onPrevious = () => setCurrentStep(step => step - 1);
  const onNext = () => setCurrentStep(step => step + 1);

  return (
    <FormValidationContextProvider
      schema={steps[currentStep].validationSchema}
      context={steps[currentStep].validationContext}
      onSubmit={isLastStep ? onSubmit : onNext}>
      <EuiFlexGroup direction="column" gutterSize="none">
        <EuiFlexItem>
          <EuiStepsHorizontal
            steps={steps.map((step, idx) => ({
              title: step.title,
              isSelected: idx === currentStep,
              isComplete: idx < currentStep,
              onClick: () => {
                idx < currentStep && setCurrentStep(idx);
              }
            }))}
          />
        </EuiFlexItem>

        <EuiSpacer size="l" />

        <EuiFlexItem>
          <StepContent>{steps[currentStep].children}</StepContent>
        </EuiFlexItem>

        <EuiSpacer size="l" />

        <EuiFlexItem>
          <StepContent>
            <FormValidationContext.Consumer>
              {({ onSubmit, isSubmitting }) => (
                <StepActions
                  currentStep={currentStep}
                  submitLabel={isLastStep ? submitLabel : "Next"}
                  onCancel={onCancel}
                  onPrevious={onPrevious}
                  onSubmit={onSubmit}
                  isSubmitting={isSubmitting}
                />
              )}
            </FormValidationContext.Consumer>
          </StepContent>
        </EuiFlexItem>
      </EuiFlexGroup>
    </FormValidationContextProvider>
  );
};
