import { useCallback, useContext } from "react";
import { CurrentProjectContext } from "@gojek/mlp-ui";
import EnvironmentsContext from "../providers/environments/context";
import { monitoringConfig } from "../config";
import template from "lodash/template";
import templateSettings from "lodash/templateSettings";

const getEnvironmentCluster = (envName, environments) => {
  const env = environments.find(e => e.name === envName);
  return env ? env.cluster : undefined;
};

const getMonitoringLink = (
  clusterName,
  projectName,
  routerName,
  routerVersion
) => {
  templateSettings.interpolate = /{{([\s\S]+?)}}/g;
  // If current page is at router details (not viewing a specific router version), the monitoring link will show
  // metrics for all revisions. Else, the monitoring link will show metrics only for a "specific" revision.
  routerVersion = /routers\/\d+\/details/.test(window.location.pathname) ? "$__all" : routerVersion
  return template(monitoringConfig.dashboardUrl)({
    cluster: clusterName,
    project: projectName,
    router: routerName,
    version: routerVersion
  });
};

export const useMonitoring = () => {
  const environments = useContext(EnvironmentsContext);
  const { project } = useContext(CurrentProjectContext);

  const getMonitoringDashboardUrl = useCallback(
    (envName, routerName, revision) => {
      const clusterName = getEnvironmentCluster(envName, environments);
      const projectName = !!project ? project.name : undefined;

      return getMonitoringLink(clusterName, projectName, routerName, revision);
    },
    [project, environments]
  );

  return [getMonitoringDashboardUrl];
};
