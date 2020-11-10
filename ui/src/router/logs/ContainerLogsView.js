import React, { useEffect, useMemo, useState } from "react";
import { ConfigSection } from "../components/configuration/components/section";
import { EuiFlexGroup, EuiFlexItem, EuiPanel, EuiSpacer } from "@elastic/eui";
import { LazyLog, ScrollFollow } from "react-lazylog";
import { LogEntry } from "../../services/logs/LogEntry";
import { LogsSearchBar } from "./search_bar/LogsSearchBar";
import { get } from "../../components/form/utils";
import { replaceBreadcrumbs, slugify } from "@gojek/mlp-ui";
import useLogsApiEventEmitter from "./hooks/useEventEmitterLogsApi";
import "./ContainerLogsView.scss";

const formatMessage = data => LogEntry.fromJson(data).toString();

export const ContainerLogsView = ({ projectId, routerId, router }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      {
        text: "Routers",
        href: "../"
      },
      {
        text: router.name,
        href: "./"
      },
      {
        text: "Logs"
      }
    ]);
  }, [projectId, routerId, router.name]);

  const components = useMemo(() => {
    return [
      "router",
      ...["enricher", "ensembler"].filter(
        component => !!get(router, `config.${component}`)
      )
    ];
  }, [router]);

  const [params, setParams] = useState({
    component_type: "router",
    tail_lines: "1000"
  });

  const { emitter } = useLogsApiEventEmitter(
    projectId,
    routerId,
    params,
    formatMessage
  );

  return (
    <ConfigSection title="Logs">
      <EuiPanel className="euiPanel--logsContainer">
        <EuiFlexGroup direction="column" gutterSize="none">
          <EuiFlexItem grow={false}>
            <LogsSearchBar {...{ components, params, setParams }} />
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
    </ConfigSection>
  );
};
