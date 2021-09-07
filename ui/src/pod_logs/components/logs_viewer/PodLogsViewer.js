import React from "react";
import { EuiFlexGroup, EuiFlexItem, EuiPanel, EuiSpacer } from "@elastic/eui";
import { LogsSearchBar } from "../search_bar/LogsSearchBar";
import { LazyLog, ScrollFollow } from "react-lazylog";
import { slugify } from "@gojek/mlp-ui";

import "./PodLogsViewer.scss";

export const PodLogsViewer = ({
  components,
  emitter,
  params,
  onParamsChange,
}) => (
  <EuiPanel className="euiPanel--logsContainer">
    <EuiFlexGroup direction="column" gutterSize="none">
      <EuiFlexItem grow={false}>
        <LogsSearchBar {...{ components, params, onParamsChange }} />
      </EuiFlexItem>
      <EuiFlexItem grow={false}>
        <EuiSpacer size="s" />
      </EuiFlexItem>
      <EuiFlexItem grow={true}>
        <ScrollFollow
          startFollowing={true}
          render={({ onScroll, follow }) => (
            <LazyLog
              key={slugify(JSON.stringify(params))}
              eventSource={emitter}
              extraLines={1}
              onScroll={onScroll}
              follow={follow}
              caseInsensitive
              enableSearch
              selectableLines
            />
          )}
        />
      </EuiFlexItem>
    </EuiFlexGroup>
  </EuiPanel>
);
