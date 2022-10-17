import React, { useEffect, useMemo } from "react";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { EuiCallOut, EuiLoadingChart, EuiTextAlign } from "@elastic/eui";
import { RouterAlertDetails } from "./details/RouterAlertDetails";
import { Routes, Route, useLocation } from "react-router-dom";
import { EditAlertsView } from "./edit/EditAlertsView";
import { useTuringApi } from "../../hooks/useTuringApi";

export const RouterAlertsView = ({ projectId, router }) => {
  const location = useLocation();
  const routerId = router?.id;
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
    if ((location.state || {}).refresh) {
      fetchAlertDetails();
    }
  }, [fetchAlertDetails, location.state]);

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
    <Routes>
      <Route index element={<RouterAlertDetails alertsData={existingAlerts} routerStatus={router.status} />} />
      <Route path="edit" element={<EditAlertsView projectId={projectId} router={router} alerts={existingAlerts} />} />
    </Routes>
  );
};
