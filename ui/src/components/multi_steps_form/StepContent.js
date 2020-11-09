import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import React from "react";

export const StepContent = ({ children, width = "75%" }) => (
  <EuiFlexGroup direction="row" justifyContent="center">
    <EuiFlexItem grow={false} style={{ width }}>
      {children}
    </EuiFlexItem>
  </EuiFlexGroup>
);
