import React, { useCallback, useState } from "react";
import { useEffect } from "react";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { ConfigSection } from "../../../components/config_section";
import { PodLogsViewer } from "../../../components/pod_logs_viewer/PodLogsViewer";
import { useTuringPollingApiEmitter } from "../../../hooks/useTuringPollingApiEmitter";
import { useLogsApiEmitter } from "../../../components/pod_logs_viewer/hooks/useLogsApiEmitter";
import { LogEntry } from "../../../services/logs/LogEntry";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiLink,
  EuiPanel,
  EuiText,
} from "@elastic/eui";
import { appConfig } from "../../../config";

const processLogs = (data) =>
  (data.logs || []).map((entry) => LogEntry.fromJson(entry).toString());

const components = [
  {
    value: "image_builder",
    name: "Image Builder",
  },
  {
    value: "driver",
    name: "Driver",
  },
  {
    value: "executor",
    name: "Executors",
  },
];

export const EnsemblingJobLogsView = ({ job }) => {
  const { podLogs: configOptions } = appConfig;
  const [loggingUrl, setLoggingUrl] = useState();

  useEffect(() => {
    replaceBreadcrumbs([
      {
        text: "Jobs",
        href: "../",
      },
      {
        text: `Job ${job.id}`,
        href: "./",
      },
      {
        text: "Logs",
      },
    ]);
  }, [job.id]);

  const [apiOptions, setApiOptions] = useState({
    query: {
      component_type: "driver",
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
    `/projects/${job.project_id}/jobs/${job.id}/logs`,
    apiOptions,
    configOptions.pollInterval
  );

  useEffect(() => {
    apiEmitter.on("data", (data) => {
      setLoggingUrl(data.logging_url);

      const entries = data.logs || [];
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
  }, [apiEmitter, setLoggingUrl, configOptions.batchSize, setQuery]);

  const { emitter } = useLogsApiEmitter(apiEmitter, processLogs);

  return (
    <ConfigSection title="Logs">
      <EuiPanel>
        <EuiFlexGroup direction="column" gutterSize="xs">
          <EuiFlexItem>
            <PodLogsViewer
              components={components}
              emitter={emitter}
              query={apiOptions.query}
              onQueryChange={setQuery}
              batchSize={configOptions.batchSize}
            />
          </EuiFlexItem>

          {!!loggingUrl && (
            <EuiFlexItem>
              <EuiFlexGroup direction="row" justifyContent="flexEnd">
                <EuiFlexItem grow={false}>
                  <EuiText size="s">
                    <EuiLink href={loggingUrl} target="_blank">
                      External Logs
                    </EuiLink>
                  </EuiText>
                </EuiFlexItem>
              </EuiFlexGroup>
            </EuiFlexItem>
          )}
        </EuiFlexGroup>
      </EuiPanel>
    </ConfigSection>
  );
};
