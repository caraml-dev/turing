import React from "react";
import { EuiFlexGroup, EuiFlexItem, EuiPanel } from "@elastic/eui";
import { ConfigSectionPanelTitle } from "../config_section";

export const ConfigMultiSectionPanel = React.forwardRef(
  ({ items, className }, ref) => {
    return (
      <EuiPanel className={`euiPanel--configSection ${className}`}>
        <div ref={ref}>
          <EuiFlexGroup direction="column" gutterSize="m">
            {items.map(({ title, appendTitle, children }, idx) => (
              <EuiFlexItem key={idx}>
                {
                  title && appendTitle ? <ConfigSectionPanelTitle title={title} append={appendTitle} />
                  : title && <ConfigSectionPanelTitle title={title} />
                }
                {children}
              </EuiFlexItem>
            ))}
          </EuiFlexGroup>
        </div>
      </EuiPanel>
    );
  }
);
