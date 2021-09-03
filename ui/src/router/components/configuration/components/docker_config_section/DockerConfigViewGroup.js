import React from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { ConfigSectionPanel } from "../../../../../components/config_section";
import { ContainerConfigTable } from "./ContainerConfigTable";
import { EnvVariablesConfigTable } from "./EnvVariablesConfigTable";
import { ResourcesConfigTable } from "../ResourcesConfigTable";
import { ConfigMultiSectionPanel } from "../../../../../components/config_multi_section_panel/ConfigMultiSectionPanel";

export const DockerConfigViewGroup = ({ componentName, dockerConfig }) => {
  const items = [
    {
      title: "Container",
      children: <ContainerConfigTable config={dockerConfig} />,
    },
    {
      title: "Environment Variables",
      children: <EnvVariablesConfigTable variables={dockerConfig.env} />,
    },
  ];

  return (
    <EuiFlexGroup direction="row" wrap>
      <EuiFlexItem grow={3} className="euiFlexItem--childFlexPanel">
        <ConfigMultiSectionPanel items={items} />
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
