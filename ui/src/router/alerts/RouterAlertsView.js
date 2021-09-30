import React, { useEffect, useMemo } from "react";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { EuiCallOut, EuiLoadingChart, EuiTextAlign } from "@elastic/eui";
import { RouterAlertDetails } from "./details/RouterAlertDetails";
import { Redirect, Router } from "@reach/router";
import { EditAlertsView } from "./edit/EditAlertsView";
import { useTuringApi } from "../../hooks/useTuringApi";

export const RouterAlertsView = ({ projectId, routerId, router, ...props }) => {
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
        text: "Alerts",
      },
    ]);
  }, [projectId, routerId, router.name]);

  const [{ data: alerts, isLoaded, error }, fetchAlertDetails] = useTuringApi(
    `/projects/${projectId}/routers/${routerId}/alerts`,
    {},
    []
  );

  useEffect(() => {
    if ((props.location.state || {}).refresh) {
      fetchAlertDetails();
    }
  }, [fetchAlertDetails, props.location.state]);

  const existingAlerts = useMemo(() => {
    let metricObj = {};
    alerts.forEach((a) => (metricObj[a.metric] = a));
    return {
      alerts: metricObj,
      team: alerts.length > 0 ? alerts[0].team : undefined,
    };
  }, [alerts]);

  return !isLoaded ? (
    <EuiTextAlign textAlign="center">
      <EuiLoadingChart size="xl" mono />
    </EuiTextAlign>
  ) : error ? (
    <EuiCallOut
      title="Sorry, there was an error"
      color="danger"
      iconType="alert">
      <p>{error.message}</p>
    </EuiCallOut>
  ) : (
    <Router primary={false}>
      <RouterAlertDetails
        path="/"
        alertsData={existingAlerts}
        routerStatus={router.status}
      />
      <EditAlertsView path="edit" router={router} alerts={existingAlerts} />

      <Redirect from="any" to="/error/404" default noThrow />
    </Router>
  );
};
