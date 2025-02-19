import React from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { ContainerConfigTable } from "./ContainerConfigTable";
import { EnvVariablesConfigTable } from "./EnvVariablesConfigTable";
import { SecretsConfigTable } from "./SecretsConfigTable";
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
    {
      title: "Secrets",
      children: <SecretsConfigTable variables={dockerConfig.secrets} />,
    },
  ];

  return (
    <EuiFlexGroup direction="row" wrap>
      <EuiFlexItem grow={3} className="euiFlexItem--childFlexPanel">
        <ConfigMultiSectionPanel items={items} />
      </EuiFlexItem>
      <EuiFlexItem grow={1} className="euiFlexItem--smallPanel">
        <ResourcesConfigTable
          componentName={componentName}
          autoscalingPolicy={dockerConfig.autoscaling_policy}
          resourceRequest={dockerConfig.resource_request}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
