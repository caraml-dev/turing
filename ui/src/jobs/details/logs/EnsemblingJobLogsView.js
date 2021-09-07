import React, { useState } from "react";
import { useEffect } from "react";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { ConfigSection } from "../../../components/config_section";
import { PodLogsViewer } from "../../../pod_logs/components/logs_viewer/PodLogsViewer";
import useLogsApiEventEmitter from "../../../pod_logs/hooks/useEventEmitterLogsApi";
import { LogEntry } from "../../../services/logs/LogEntry";

const formatMessage = (data) => LogEntry.fromJson(data).toString();

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

  const [params, setParams] = useState({
    component_type: "driver",
    tail_lines: "1000",
  });

  const { emitter } = useLogsApiEventEmitter(
    `/projects/${job.project_id}/jobs/${job.id}/logs`,
    params,
    formatMessage
  );

  return (
    <ConfigSection title="Logs">
      <PodLogsViewer
        components={components}
        emitter={emitter}
        params={params}
        onParamsChange={setParams}
      />
    </ConfigSection>
  );
};
