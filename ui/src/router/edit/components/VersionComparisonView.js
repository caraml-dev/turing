import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiPanel } from "@elastic/eui";
import { StepActions } from "@gojek/mlp-ui";
import { ConfigSection } from "../../../components/config_section";
import { VersionComparisonPanel } from "../../versions/components/version_diff/VersionComparisonPanel";
import { RouterVersion } from "../../../services/version/RouterVersion";
import ExperimentEngineContext from "../../../providers/experiments/context";

export const VersionComparisonView = ({
  currentRouter,
  updatedRouter,
  onPrevious,
  onSubmit,
  isSubmitting,
}) => {
  const { getEngineProperties } = useContext(ExperimentEngineContext);
  const currentVersionContext = {
    experiment_engine: getEngineProperties(
      currentRouter.config?.experiment_engine?.type
    ),
  };
  const updatedVersionContext = {
    experiment_engine: getEngineProperties(
      updatedRouter.config?.experiment_engine?.type
    ),
  };

  return (
    <EuiFlexGroup direction="column">
      <EuiFlexItem>
        <ConfigSection title="Comparing Versions">
          <EuiPanel>
            <VersionComparisonPanel
              leftValue={RouterVersion.fromJson(
                currentRouter.config
              ).toPrettyYaml(currentVersionContext)}
              rightValue={RouterVersion.fromJson(
                updatedRouter.config
              ).toPrettyYaml(updatedVersionContext)}
              leftTitle={"Current Version"}
              rightTitle={"New Version"}
            />
          </EuiPanel>
        </ConfigSection>
      </EuiFlexItem>

      <EuiFlexItem>
        <StepActions
          currentStep={1}
          submitLabel="Deploy"
          onPrevious={onPrevious}
          onSubmit={onSubmit}
          isSubmitting={isSubmitting}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
