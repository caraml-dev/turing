import React, { useContext, useEffect, useState } from "react";
import { ConfigSection } from "../../../components/config_section";
import { LogEntry } from "../../../services/logs/LogEntry";
import { get } from "../../../components/form/utils";
import { ProjectsContext, replaceBreadcrumbs } from "@caraml-dev/ui-lib";
import { PodLogsViewer } from "../../../components/pod_logs_viewer/PodLogsViewer";
import { useConfig } from "../../../config";
import { useLogsEmitter } from "../../../components/pod_logs_viewer/hooks/useLogsEmitter";
import { createStackdriverUrl } from "../../../utils/createStackdriverUrl";
import EnsemblersContext from "../../../providers/ensemblers/context";
import EnvironmentsContext from "../../../providers/environments/context";

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
  const { appConfig } = useConfig();

  const { currentProject } = useContext(ProjectsContext);

  const { ensemblers } = useContext(EnsemblersContext);
  const ensembler = Object.values(ensemblers)
      .find((value) => value.id === router?.config?.ensembler?.pyfunc_config?.ensembler_id)

  const environments = useContext(EnvironmentsContext);
  const environment = Object.values(environments)
      .find((value) => value.name === router?.environment_name)

  const [stackdriverUrls, setStackdriverUrls] = useState({});

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
    tail_lines: appConfig.podLogs.defaultTailLines,
  });

  const { emitter } = useLogsEmitter(
    `/projects/${router.project_id}/routers/${router.id}/logs`,
    query,
    (entries) =>
      !!entries.length ? entries[entries.length - 1].timestamp : undefined,
    (entries) => entries.map((entry) => LogEntry.fromJson(entry).toString()),
      appConfig.podLogs,
  );

  const availableComponents = components.filter(
    (c) => c.value === "router" || !!get(router, `config.${c.value}`)
  );

  useEffect(
    () => {
      let urls = {}
      if (
          appConfig.imagebuilder.cluster &&
          appConfig.imagebuilder.gcp_project &&
          appConfig.imagebuilder.namespace &&
          currentProject
      ) {
        // set router url
        urls["router"] = createStackdriverUrl({
          gcp_project: environment.gcp_project,
          cluster: environment.cluster,
          namespace: currentProject.name,
          pod_name: router.name + "-turing-router-"  + router.config.version,
          start_time: router.updated_at,
        }, "router")

        // set enricher url
        if (router.config.enricher.type === "docker") {
          urls["enricher"] = createStackdriverUrl({
            gcp_project: environment.gcp_project,
            cluster: environment.cluster,
            namespace: currentProject.name,
            pod_name: router.name + "-turing-enricher-" + router.config.version,
            start_time: router.updated_at,
          }, "enricher")
        }

        // set ensembler url
        if (router.config.ensembler.type === "docker" || router.config.ensembler.type === "pyfunc") {
          urls["ensembler"] = createStackdriverUrl({
            gcp_project: environment.gcp_project,
            cluster: environment.cluster,
            namespace: currentProject.name,
            pod_name: router.name + "-turing-ensembler-" + router.config.version,
            start_time: router.updated_at,
          }, "ensembler")
        }

        // set image builder url
        if (router.config.ensembler.type === "pyfunc" && ensembler) {
          urls["ensembler_image_builder"] = createStackdriverUrl({
            job_name: "service-" + currentProject.name + "-" + ensembler.name,
            start_time: ensembler.updated_at,
          }, "ensembler_image_builder")
        }

        setStackdriverUrls(urls);
      }
    },
    [currentProject, ensembler, environment, router, appConfig.imagebuilder]
  );

  return (
    <ConfigSection title="Logs">
      <PodLogsViewer
        components={availableComponents}
        emitter={emitter}
        query={query}
        onQueryChange={setQuery}
        batchSize={appConfig.podLogs.batchSize}
        stackdriverUrls={stackdriverUrls}
      />
    </ConfigSection>
  );
};
