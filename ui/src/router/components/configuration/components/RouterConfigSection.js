import React, { useState } from "react";
import { EuiButtonGroup, EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { ConfigSectionPanel } from "../../../../components/config_section";
import { ResourcesConfigTable } from "./ResourcesConfigTable";
import { RoutesConfigTable } from "./router_config_section/RoutesConfigTable";
import { RulesConfigTable } from "./router_config_section/RulesConfigTable";

export const RouterConfigSection = ({ config }) => {
  // Toggle buttons
  const [routesViewMode, setRoutesViewMode] = useState("Routes");

  const onChangeToggle = (optionId) => {
    setRoutesViewMode(optionId);
  };

  const toggleButtons = [
    {
      id: "Routes",
      label: 'Routes',
    },
    {
      id: "Rules",
      label: 'Rules',
    },
  ];

  return (
    <EuiFlexGroup direction="row" wrap>
      <EuiFlexItem grow={3}>
        <ConfigSectionPanel
          title={routesViewMode}
          appendTitle={
            <EuiButtonGroup
              legend="Group by routes or rules"
              options={toggleButtons}
              idSelected={routesViewMode}
              onChange={(id) => onChangeToggle(id)}
            />
          }
        >
          {routesViewMode === "Routes" ? 
            <RoutesConfigTable
              routes={config.routes}
              rules={config.rules}
              defaultRouteId={config.default_route_id}
              defaultTrafficRule={config.default_traffic_rule}
            /> :
            <RulesConfigTable
              routes={config.routes}
              rules={config.rules}
              defaultTrafficRule={config.default_traffic_rule}
            />
          }
        </ConfigSectionPanel>
      </EuiFlexItem>

    <EuiFlexItem grow={1} className="euiFlexItem--smallPanel">
      <ResourcesConfigTable
        componentName="Router"
        autoscalingPolicy={config.autoscaling_policy}
        resourceRequest={config.resource_request}
      />
    </EuiFlexItem>
  </EuiFlexGroup>
  )
};
