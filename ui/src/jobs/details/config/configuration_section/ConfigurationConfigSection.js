import React from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { ConfigSectionPanel } from "../../../../components/config_section";
import { ResourceRequestConfigTable } from "./resource_request_table/ResourceRequestConfigTable";
import { ConfigMultiSectionPanel } from "../../../../components/config_multi_section_panel/ConfigMultiSectionPanel";
import { EnvVariablesConfigTable } from "../../../../router/components/configuration/components/docker_config_section/EnvVariablesConfigTable";
import { MiscConfigSection } from "./misc_table/MiscConfigSection";

export const ConfigurationConfigSection = ({ job: { infra_config = {} } }) => {
  const items = [
    {
      title: "Miscellaneous Info",
      children: <MiscConfigSection infra_config={infra_config} />,
    },
    {
      title: "Environment Variables",
      children: (
        <EnvVariablesConfigTable variables={infra_config.env_vars || []} />
      ),
    },
  ];

  return (
    <EuiFlexGroup direction="row">
      <EuiFlexItem grow={3} className="euiFlexItem--childFlexPanel">
        <ConfigMultiSectionPanel items={items} />
      </EuiFlexItem>

      <EuiFlexItem grow={1} className="euiFlexItem--smallPanel">
        <ConfigSectionPanel title="Resource Request">
          <ResourceRequestConfigTable resources={infra_config.resources} />
        </ConfigSectionPanel>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
