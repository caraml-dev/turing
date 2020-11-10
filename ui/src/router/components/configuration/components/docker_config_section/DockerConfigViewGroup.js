import React from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { ConfigSectionPanel, ConfigSectionPanelTitle } from "../section";
import { ContainerConfigTable } from "./ContainerConfigTable";
import { EnvVariablesConfigTable } from "./EnvVariablesConfigTable";
import { ResourcesConfigTable } from "../ResourcesConfigTable";

export const DockerConfigViewGroup = ({ componentName, dockerConfig }) => {
  return (
    <EuiFlexGroup direction="row" wrap>
      <EuiFlexItem grow={3}>
        <EuiFlexGroup direction="row" wrap>
          <EuiFlexItem grow={2}>
            <ConfigSectionPanel>
              <EuiFlexGroup direction="column" gutterSize="m">
                <EuiFlexItem>
                  <ConfigSectionPanelTitle title="Container" />
                  <ContainerConfigTable config={dockerConfig} />
                </EuiFlexItem>

                <EuiFlexItem>
                  <ConfigSectionPanelTitle title="Environment Variables" />
                  <EnvVariablesConfigTable variables={dockerConfig.env} />
                </EuiFlexItem>
              </EuiFlexGroup>
            </ConfigSectionPanel>
          </EuiFlexItem>
        </EuiFlexGroup>
      </EuiFlexItem>
      <EuiFlexItem grow={1} className="euiFlexItem--smallPanel">
        <ConfigSectionPanel title={`${componentName} Resources`}>
          <ResourcesConfigTable
            resourceRequest={dockerConfig.resource_request}
          />
        </ConfigSectionPanel>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
