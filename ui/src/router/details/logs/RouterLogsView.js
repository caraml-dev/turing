import React, { useEffect, useState } from "react";
import { ConfigSection } from "../../../components/config_section";
import { LogEntry } from "../../../services/logs/LogEntry";
import { get } from "../../../components/form/utils";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import useLogsApiEventEmitter from "../../../pod_logs/hooks/useEventEmitterLogsApi";
import { PodLogsViewer } from "../../../pod_logs/components/logs_viewer/PodLogsViewer";

const processLogs = (data) => {
  const chunk = data
    .map((entry) => LogEntry.fromJson(entry).toString())
    .join("\n");

  const timestamp = !!data.length ? data[data.length - 1].timestamp : undefined;

  return { chunk, timestamp };
};

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

export const RouterLogsView = ({ projectId, routerId, router }) => {
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
  }, [projectId, routerId, router.name]);

  const [params, setParams] = useState({
    component_type: "router",
    tail_lines: "1000",
  });

  const { emitter } = useLogsApiEventEmitter(
    `/projects/${projectId}/routers/${routerId}/logs`,
    params,
    processLogs
  );

  return (
    <ConfigSection title="Logs">
      <PodLogsViewer
        components={components.filter(
          (c) => c.value === "router" || !!get(router, `config.${c.value}`)
        )}
        emitter={emitter}
        params={params}
        onParamsChange={setParams}
      />
    </ConfigSection>
  );
};
