import React from "react";
import {
  EuiBadge,
  EuiFlexGroup,
  EuiFlexItem,
  EuiText,
  EuiTextColor,
  EuiToolTip
} from "@elastic/eui";

export const RuleCardRouteDropDownOption = ({ id, endpoint, isDefault }) => {
  const option = (
    <EuiFlexGroup direction="row" gutterSize="s">
      <EuiFlexItem grow={false}>
        <EuiBadge color={isDefault ? "hollow" : "default"}>{id}</EuiBadge>
      </EuiFlexItem>
      <EuiFlexItem className="eui-textTruncate">
        <EuiTextColor color="subdued">
          <EuiText size="s" className="eui-textTruncate">
            {endpoint}
          </EuiText>
        </EuiTextColor>
      </EuiFlexItem>
    </EuiFlexGroup>
  );

  return isDefault ? (
    <EuiToolTip content="Default route can't be a part of any rules">
      {option}
    </EuiToolTip>
  ) : (
    option
  );
};
