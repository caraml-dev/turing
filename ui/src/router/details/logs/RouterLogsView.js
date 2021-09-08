import React, { useCallback, useEffect, useState } from "react";
import { ConfigSection } from "../../../components/config_section";
import { LogEntry } from "../../../services/logs/LogEntry";
import { get } from "../../../components/form/utils";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { PodLogsViewer } from "../../../components/pod_logs_viewer/PodLogsViewer";
import { EuiPanel } from "@elastic/eui";
import { useTuringPollingApiEmitter } from "../../../hooks/useTuringPollingApiEmitter";
import { useLogsApiEmitter } from "../../../components/pod_logs_viewer/hooks/useLogsApiEmitter";
import { appConfig } from "../../../config";

const processLogs = (data) =>
  data.map((entry) => LogEntry.fromJson(entry).toString());

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

  const [apiOptions, setApiOptions] = useState({
    query: {
      component_type: "router",
      tail_lines: configOptions.defaultTailLines,
    },
  });

  const setQuery = useCallback(
    (setQuery) => {
      setApiOptions((options) => ({
        ...options,
        query: setQuery(options.query),
      }));
    },
    [setApiOptions]
  );

  const { emitter: apiEmitter } = useTuringPollingApiEmitter(
    `/projects/${router.project_id}/routers/${router.id}/logs`,
    apiOptions,
    configOptions.pollInterval
  );

  useEffect(() => {
    apiEmitter.on("data", (entries) => {
      const lastTimestamp = !!entries.length
        ? entries[entries.length - 1].timestamp
        : undefined;

      if (!!lastTimestamp) {
        setQuery((q) => ({
          ...q,
          since_time: lastTimestamp,
          head_lines: configOptions.batchSize,
        }));
      }
    });

    apiEmitter.emit("start");

    return () => {
      apiEmitter.emit("abort");
    };
  }, [apiEmitter, setQuery, configOptions.batchSize]);

  const { emitter } = useLogsApiEmitter(apiEmitter, processLogs);

  return (
    <ConfigSection title="Logs">
      <EuiPanel>
        <PodLogsViewer
          components={components.filter(
            (c) => c.value === "router" || !!get(router, `config.${c.value}`)
          )}
          emitter={emitter}
          query={apiOptions.query}
          onQueryChange={setQuery}
          batchSize={configOptions.batchSize}
        />
      </EuiPanel>
    </ConfigSection>
  );
};
