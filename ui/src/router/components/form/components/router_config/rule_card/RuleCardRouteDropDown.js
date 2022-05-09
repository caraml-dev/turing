import React from "react";
import {
  EuiBadge,
  EuiFlexGroup,
  EuiFlexItem,
  EuiText,
  EuiTextColor,
} from "@elastic/eui";

export const RuleCardRouteDropDownOption = ({ id, endpoint }) => {
  const option = (
    <EuiFlexGroup direction="row" gutterSize="s">
      <EuiFlexItem grow={false}>
        <EuiBadge>{id}</EuiBadge>
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

  return option;
};
