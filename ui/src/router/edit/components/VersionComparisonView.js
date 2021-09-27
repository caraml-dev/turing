import React from "react";
import { EuiFlexGroup, EuiFlexItem, EuiPanel } from "@elastic/eui";
import { ConfigSection } from "../../../components/config_section";
import { VersionComparisonPanel } from "../../versions/components/version_diff/VersionComparisonPanel";
import { RouterVersion } from "../../../services/version/RouterVersion";
import { StepActions } from "../../../components/multi_steps_form/StepActions";

export const VersionComparisonView = ({
  currentRouter,
  updatedRouter,
  onPrevious,
  onSubmit,
  isSubmitting,
}) => (
  <EuiFlexGroup direction="column">
    <EuiFlexItem>
      <ConfigSection title="Comparing Versions">
        <EuiPanel>
          <VersionComparisonPanel
            leftValue={RouterVersion.fromJson(
              currentRouter.config
            ).toPrettyYaml()}
            rightValue={RouterVersion.fromJson(
              updatedRouter.config
            ).toPrettyYaml()}
            leftTitle={"Current Version"}
            rightTitle={"New Version"}
          />
        </EuiPanel>
      </ConfigSection>
    </EuiFlexItem>

    <EuiFlexItem>
      <StepActions
        submitLabel="Deploy"
        onPrevious={onPrevious}
        onSubmit={onSubmit}
        isSubmitting={isSubmitting}
      />
    </EuiFlexItem>
  </EuiFlexGroup>
);
