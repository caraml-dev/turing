import React, { useEffect, useState } from "react";
import { ConfigSection } from "../../../components/config_section";
import { LogEntry } from "../../../services/logs/LogEntry";
import { get } from "../../../components/form/utils";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { PodLogsViewer } from "../../../components/pod_logs_viewer/PodLogsViewer";
import { EuiPanel } from "@elastic/eui";
import { appConfig } from "../../../config";
import { useLogsEmitter } from "../../../components/pod_logs_viewer/hooks/useLogsEmitter";

const components = [
  {
    value: "router",
    name: "Router",
  },
  {
    value: "enricher",
    name: "Enricher",
  },
  {
    value: "ensembler",
    name: "Ensembler",
  },
];

export const RouterLogsView = ({ router }) => {
  const { podLogs: configOptions } = appConfig;

  useEffect(() => {
    replaceBreadcrumbs([
      {
        text: "Routers",
        href: "../",
      },
      {
        text: router.name,
        href: "./",
      },
      {
        text: "Logs",
      },
    ]);
  }, [router.name]);

  const [query, setQuery] = useState({
    component_type: "router",
    tail_lines: configOptions.defaultTailLines,
  });

  const { emitter } = useLogsEmitter(
    `/projects/${router.project_id}/routers/${router.id}/logs`,
    query,
    (entries) =>
      !!entries.length ? entries[entries.length - 1].timestamp : undefined,
    (entries) => entries.map((entry) => LogEntry.fromJson(entry).toString())
  );

  const availableComponents = components.filter(
    (c) => c.value === "router" || !!get(router, `config.${c.value}`)
  );

  return (
    <ConfigSection title="Logs">
      <EuiPanel>
        <PodLogsViewer
          components={availableComponents}
          emitter={emitter}
          query={query}
          onQueryChange={setQuery}
          batchSize={configOptions.batchSize}
        />
      </EuiPanel>
    </ConfigSection>
  );
};
