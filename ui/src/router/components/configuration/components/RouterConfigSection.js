import React from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { ResourcesConfigTable } from "./ResourcesConfigTable";
import { ConfigSectionPanel } from "./section";
import { RoutesConfigTable } from "./router_config_section/RoutesConfigTable";

export const RouterConfigSection = ({ config }) => {
  return (
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
        <ConfigSectionPanel title="Router Resources">
          <ResourcesConfigTable resourceRequest={config.resource_request} />
        </ConfigSectionPanel>
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
