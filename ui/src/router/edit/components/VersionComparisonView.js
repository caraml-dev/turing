import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiPanel } from "@elastic/eui";
import { ConfigSection } from "../../../components/config_section";
import { VersionComparisonPanel } from "../../versions/components/version_diff/VersionComparisonPanel";
import { RouterVersion } from "../../../services/version/RouterVersion";
import { StepActions } from "../../../components/multi_steps_form/StepActions";
import { ExperimentEngineContextProvider } from "../../../providers/experiments/ExperimentEngineContextProvider";
import ExperimentEngineContext from "../../../providers/experiments/context";

const VersionComparisonView = ({
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
          submitLabel="Deploy"
          onPrevious={onPrevious}
          onSubmit={onSubmit}
          isSubmitting={isSubmitting}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};

const VersionComparisonViewWrapper = (props) => (
  <ExperimentEngineContextProvider>
    <VersionComparisonView {...props} />
  </ExperimentEngineContextProvider>
);

export { VersionComparisonViewWrapper as VersionComparisonView };
