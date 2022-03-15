import React, { useContext } from "react";
import { EuiButton, EuiFlexGroup, EuiFlexItem, EuiPanel } from "@elastic/eui";
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
  setWithDeployment,
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
        <EuiFlexGroup direction="row" justifyContent="flexEnd">
          <EuiFlexItem grow={false}>
            <EuiButton size="s" color="primary" onClick={onPrevious}>
              Previous
            </EuiButton>
          </EuiFlexItem>
          <EuiFlexItem grow={false}>
            <EuiButton
              size="s"
              color="warning"
              fill={true}
              isLoading={isSubmitting}
              onClick={() => {
                setWithDeployment(false);
                return onSubmit();
              }}>
              Save
            </EuiButton>
          </EuiFlexItem>
          <EuiFlexItem grow={false}>
            <EuiButton
              size="s"
              color="primary"
              fill={true}
              isLoading={isSubmitting}
              onClick={() => {
                setWithDeployment(true);
                return onSubmit();
              }}>
              Deploy
            </EuiButton>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
