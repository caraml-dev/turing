import React, { Fragment } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiHorizontalRule, EuiTitle } from "@elastic/eui";

export const ConfigSectionPanelTitle = ({ title, append }) => (
  <Fragment>
    {append ? 
    <EuiFlexGroup justifyContent="spaceBetween">
      <EuiFlexItem grow={false}>
        <EuiTitle size="xs">
          <span>{title}</span>
        </EuiTitle>
      </EuiFlexItem>
      <EuiFlexItem grow={false}>
        {append}
      </EuiFlexItem>
    </EuiFlexGroup>
    : <EuiTitle size="xs">
      <span>{title}</span>
    </EuiTitle>}
    <EuiHorizontalRule size="full" margin="xs" />
  </Fragment>
);
