import React from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { RouterActivityLogView } from "../activity_log/RouterActivityLogView";
import { ListRouterVersionsView } from "../versions/list/ListRouterVersionsView";

export const HistoryView = ({ router, ...props }) => {
  return (
    <EuiFlexGroup direction="row">
      <EuiFlexItem grow={5}>
        <ListRouterVersionsView router={router} {...props} />
      </EuiFlexItem>

      <EuiFlexItem grow={3}>
        <RouterActivityLogView router={router} {...props} />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
