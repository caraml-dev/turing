import React, { useState } from "react";
import { useEffect } from "react";
import { replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import { ConfigSection } from "../../../components/config_section";
import { PodLogsViewer } from "../../../components/pod_logs_viewer/PodLogsViewer";
import { LogEntry } from "../../../services/logs/LogEntry";
import {
  EuiFlexGroup,
  EuiFlexItem,
  EuiLink,
  EuiPanel,
  EuiText,
} from "@elastic/eui";
import { useConfig } from "../../../config";
import { useLogsEmitter } from "../../../components/pod_logs_viewer/hooks/useLogsEmitter";

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
  const {
    appConfig: { podLogs: configOptions },
  } = useConfig();
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

  const [query, setQuery] = useState({
    component_type: "driver",
    tail_lines: configOptions.defaultTailLines,
  });

  const { emitter } = useLogsEmitter(
    `/projects/${job.project_id}/jobs/${job.id}/logs`,
    query,
    (data) => {
      const entries = data.logs || [];
      return !!entries.length
        ? entries[entries.length - 1].timestamp
        : undefined;
    },
    (data) => {
      // Update External Logs URL
      setLoggingUrl(data.logging_url);

      return (data.logs || []).map((entry) =>
        LogEntry.fromJson(entry).toString()
      );
    },
    configOptions
  );

  return (
    <ConfigSection title="Logs">
      <EuiPanel>
        <EuiFlexGroup direction="column" gutterSize="xs">
          <EuiFlexItem>
            <PodLogsViewer
              components={components}
              emitter={emitter}
              query={query}
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
