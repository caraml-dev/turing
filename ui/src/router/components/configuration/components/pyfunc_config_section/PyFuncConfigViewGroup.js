import React from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { ContainerConfigTable } from "../docker_config_section/ContainerConfigTable";
import { ResourcesConfigTable } from "../ResourcesConfigTable";
import { ConfigMultiSectionPanel } from "../../../../../components/config_multi_section_panel/ConfigMultiSectionPanel";
import { PyFuncConfigTable } from "./PyFuncConfigTable";
import { EnvVariablesConfigTable } from "../docker_config_section/EnvVariablesConfigTable";

export const PyFuncConfigViewGroup = ({
  componentName,
  pyfuncConfig,
  dockerConfig,
}) => {
  const items = [
    {
      title: "Pyfunc Ensembler Details",
      children: <PyFuncConfigTable config={pyfuncConfig} />,
    },
  ];

  if (!!dockerConfig) {
    items.push({
      title: "Container",
      children: <ContainerConfigTable config={dockerConfig} />,
    });
  }

  items.push({
    title: "Environment Variables",
    children: <EnvVariablesConfigTable variables={pyfuncConfig.env} />,
  });

  return (
    <EuiFlexGroup direction="row" wrap>
      <EuiFlexItem grow={3} className="euiFlexItem--childFlexPanel">
        <ConfigMultiSectionPanel items={items} />
      </EuiFlexItem>
      <EuiFlexItem grow={1} className="euiFlexItem--smallPanel">
        <ResourcesConfigTable
          componentName="Ensembler"
          autoscalingPolicy={dockerConfig.autoscaling_policy}
          resourceRequest={pyfuncConfig.resource_request}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
