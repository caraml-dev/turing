import React from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { ConfigSectionPanel } from "../../../../../components/config_section";
import { ContainerConfigTable } from "../docker_config_section/ContainerConfigTable";
import { ResourcesConfigTable } from "../ResourcesConfigTable";
import { ConfigMultiSectionPanel } from "../../../../../components/config_multi_section_panel/ConfigMultiSectionPanel";
import { PyFuncRefConfigTable } from "./PyFuncRefConfigTable";
import { EnvVariablesConfigTable } from "../docker_config_section/EnvVariablesConfigTable";

export const PyFuncConfigViewGroup = ({
  componentName,
  pyfuncConfig,
  dockerConfig,
}) => {
  const items = [
    {
      title: "Pyfunc Ensembler Details",
      children: <PyFuncRefConfigTable config={pyfuncConfig} />,
    },
  ];

  if (!!dockerConfig) {
    items.splice(
      items.length,
      0,
      {
        title: "Container",
        children: <ContainerConfigTable config={dockerConfig} />,
      },
      {
        title: "Environment Variables",
        children: <EnvVariablesConfigTable variables={dockerConfig.env} />,
      }
    );
  } else {
    items.push({
      title: "Environment Variables",
      children: <EnvVariablesConfigTable variables={pyfuncConfig.env} />,
    });
  }

  return (
    <EuiFlexGroup direction="row" wrap>
      <EuiFlexItem grow={3} className="euiFlexItem--childFlexPanel">
        <ConfigMultiSectionPanel items={items} />
      </EuiFlexItem>
      <div>
        {!!dockerConfig ? (
          <EuiFlexItem grow={1} className="euiFlexItem--smallPanel">
            <ConfigSectionPanel title={`${componentName} Resources`}>
              <ResourcesConfigTable
                resourceRequest={dockerConfig.resource_request}
              />
            </ConfigSectionPanel>
          </EuiFlexItem>
        ) : null}
      </div>
    </EuiFlexGroup>
  );
};
