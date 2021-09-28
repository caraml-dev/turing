import React from "react";
import { EuiFlexGroup, EuiFlexItem, EuiIcon, EuiTitle } from "@elastic/eui";
import { useConfig } from "../../config";

export const PageTitle = ({ title, size, icon, iconSize, prepend }) => {
  const { appConfig } = useConfig();
  return (
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
};
