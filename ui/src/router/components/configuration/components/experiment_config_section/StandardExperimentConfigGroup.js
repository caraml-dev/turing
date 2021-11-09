import React, { useContext } from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { ConfigSectionPanel } from "../../../../../components/config_section";
import { CredentialsConfigSection } from "./credentials/CredentialsConfigSection";
import { ExperimentsConfigTable } from "./experiments/ExperimentsConfigTable";
import { VariablesConfigTable } from "./variables/VariablesConfigTable";
import ExperimentEngineContext from "../../../../../providers/experiments/context";

export const StandardExperimentConfigGroup = ({
  projectId,
  engineType,
  engineConfig,
}) => {
  const { getEngineProperties } = useContext(ExperimentEngineContext);
  const engineProps = getEngineProperties(engineType);

  return (
    <EuiFlexGroup direction="row" wrap>
      <EuiFlexItem grow={3}>
        <EuiFlexGroup direction="row" wrap>
          <EuiFlexItem grow={1} className="euiFlexItem--smallPanel">
            <ConfigSectionPanel
              title="Experiment Engine"
              className="experimentCredentials">
              <CredentialsConfigSection
                projectId={projectId}
                engineType={engineType}
                deployment={engineConfig.deployment}
                client={engineConfig.client}
                engineProps={engineProps}
              />
            </ConfigSectionPanel>
          </EuiFlexItem>

          <EuiFlexItem grow={2}>
            <ConfigSectionPanel className="experiments" title="Experiments">
              <ExperimentsConfigTable
                experiments={engineConfig.experiments}
                engineProps={engineProps}
              />
            </ConfigSectionPanel>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlexItem>

      <EuiFlexItem grow={1} className="euiFlexItem--smallPanel">
        <ConfigSectionPanel title="Variables" className="variables">
          <VariablesConfigTable
            variables={engineConfig.variables}
            engineProps={engineProps}
          />
        </ConfigSectionPanel>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
