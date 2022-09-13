import React from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { ConfigSectionPanel } from "../../../../components/config_section";
import { ResourcesConfigTable } from "./ResourcesConfigTable";
import { RoutesConfigTable } from "./router_config_section/RoutesConfigTable";

export const RouterConfigSection = ({ config }) => (
  <EuiFlexGroup direction="row" wrap>
    <EuiFlexItem grow={3}>
      <ConfigSectionPanel title="Routes">
        <RoutesConfigTable
          routes={config.routes}
          rules={config.rules}
          defaultRouteId={config.default_route_id}
        />
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
);
