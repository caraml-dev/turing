import React from "react";
import { EuiFlexGroup, EuiFlexItem, EuiPanel } from "@elastic/eui";
import { ConfigSectionPanelTitle } from "../config_section";

export const ConfigMultiSectionPanel = React.forwardRef(
  ({ items, className }, ref) => {
    return (
      <EuiPanel className={`euiPanel--configSection ${className}`}>
        <div ref={ref}>
          <EuiFlexGroup direction="column" gutterSize="m">
            {items.map(({ title, children }, idx) => (
              <EuiFlexItem key={idx}>
                {title && <ConfigSectionPanelTitle title={title} />}
                {children}
              </EuiFlexItem>
            ))}
          </EuiFlexGroup>
        </div>
      </EuiPanel>
    );
  }
);
