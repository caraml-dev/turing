import React, { Fragment } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiHorizontalRule, EuiTitle } from "@elastic/eui";

export const ConfigSectionPanelTitleWithToggle = ({ title, toggle }) => (
  <Fragment>
    <EuiFlexGroup justifyContent="spaceBetween">
      <EuiFlexItem grow={false}>
        <EuiTitle size="xs">
          <span>{title}</span>
        </EuiTitle>
      </EuiFlexItem>
      <EuiFlexItem grow={false}>
        {toggle}
      </EuiFlexItem>
    </EuiFlexGroup>
    <EuiHorizontalRule size="full" margin="xs" />
  </Fragment>
);
