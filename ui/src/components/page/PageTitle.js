import { EuiFlexGroup, EuiFlexItem, EuiIcon, EuiTitle } from "@elastic/eui";
import { appConfig } from "../../config";
import React from "react";

export const PageTitle = ({ title, size, icon, iconSize, prepend }) => (
  <EuiFlexGroup direction="row" gutterSize="s" alignItems="center">
    <EuiFlexItem grow={false}>
      <EuiIcon type={icon || appConfig.appIcon} size={iconSize || "xl"} />
    </EuiFlexItem>

    <EuiFlexItem grow={false}>
      <EuiTitle size={size || "l"}>
        <span>{title}</span>
      </EuiTitle>
    </EuiFlexItem>

    {!!prepend && <EuiFlexItem grow={false}>{prepend}</EuiFlexItem>}
  </EuiFlexGroup>
);
