import React from "react";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";
import { RouterActivityLogView } from "../activity_log/RouterActivityLogView";
import { ListRouterVersionsView } from "../versions/list/ListRouterVersionsView";

export const HistoryView = ({ projectId, router }) => {
  return (
    <EuiFlexGroup direction="row">
      <EuiFlexItem grow={4} className="euiFlexItem--mediumPanel">
        <ListRouterVersionsView router={router} />
      </EuiFlexItem>

      <EuiFlexItem grow={3} className="euiFlexItem--mediumPanel">
        <RouterActivityLogView projectId={projectId} router={router} />
      </EuiFlexItem>
    </EuiFlexGroup>
  );
};
