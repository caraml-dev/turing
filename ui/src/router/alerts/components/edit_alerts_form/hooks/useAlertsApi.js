import { useCallback, useEffect, useState } from "react";
import { useTuringApi } from "../../../../../hooks/useTuringApi";

// TODO: requires refactoring
export const useAlertsApi = (
  projectId,
  routerId,
  environmentName,
  onCancel,
  onSuccess
) => {
  const [request, setRequest] = useState({
    alerts: [],
    team: undefined
  });
  const [sendingAlert, setSendingAlert] = useState({
    method: undefined,
    alertId: undefined,
    submitBody: {}
  });
  const [isSubmitSuccess, setIsSubmitSuccess] = useState(true);

  const [createAlertResponse, createAlert] = useTuringApi(
    `/projects/${projectId}/routers/${routerId}/alerts`,
    {
      method: "POST",
      headers: { "Content-Type": "application/json" }
    },
    {},
    false
  );

  const [updateAlertResponse, updateAlert] = useTuringApi(
    `/projects/${projectId}/routers/${routerId}/alerts/${sendingAlert.alertId}`,
    {
      method: "PUT",
      headers: { "Content-Type": "application/json" }
    },
    {},
    false
  );

  const [deleteAlertResponse, deleteAlert] = useTuringApi(
    `/projects/${projectId}/routers/${routerId}/alerts/${sendingAlert.alertId}`,
    {
      method: "DELETE",
      headers: { "Content-Type": "application/json" }
    },
    {},
    false
  );

  useEffect(() => {
    if (request.alerts.length > 0 && !sendingAlert.method) {
      const body = {
        environment: environmentName,
        team: request.team
      };
      const changedAlerts = [...request.alerts];
      const a = changedAlerts.pop();
      setRequest({
        ...request,
        alerts: changedAlerts
      });
      const method = a.method;
      const alertId = a.id;
      let submitBody = {
        ...a,
        ...body
      };
      delete submitBody.method;
      setSendingAlert({
        method: method,
        alertId: alertId,
        submitBody: submitBody
      });
    }
  }, [request, sendingAlert.method, environmentName]);

  useEffect(() => {
    if (sendingAlert.method) {
      switch (sendingAlert.method) {
        case "create":
          createAlert({ body: JSON.stringify(sendingAlert.submitBody) });
          break;
        case "update":
          updateAlert({ body: JSON.stringify(sendingAlert.submitBody) });
          break;
        case "delete":
          deleteAlert();
          break;
        default:
          console.log("method not found");
      }
    }
  }, [sendingAlert, createAlert, updateAlert, deleteAlert]);

  useEffect(() => {
    if (
      deleteAlertResponse.isLoaded ||
      createAlertResponse.isLoaded ||
      updateAlertResponse.isLoaded
    ) {
      if (
        deleteAlertResponse.error ||
        createAlertResponse.error ||
        updateAlertResponse.error
      ) {
        setIsSubmitSuccess(false);
      }
      deleteAlertResponse.isLoaded = createAlertResponse.isLoaded = updateAlertResponse.isLoaded = false;
      setSendingAlert({
        ...sendingAlert,
        method: undefined
      });
      if (request.alerts.length === 0) {
        onSuccess(isSubmitSuccess);
      }
    }
  }, [
    deleteAlertResponse,
    createAlertResponse,
    updateAlertResponse,
    isSubmitSuccess,
    setIsSubmitSuccess,
    sendingAlert,
    request.alerts,
    onSuccess
  ]);

  const alertMethod = (oldAlert, newAlert, hasTeamChanged) => {
    if (!oldAlert && !!newAlert) {
      return "create";
    }
    if (!!oldAlert && !newAlert) {
      return "delete";
    }
    if (!!oldAlert && !!newAlert) {
      if (hasTeamChanged) {
        return "recreate";
      } else if (JSON.stringify(oldAlert) !== JSON.stringify(newAlert)) {
        return "update";
      }
    }
    return "";
  };

  const submitAlerts = useCallback(
    (existingAlerts, newAlerts, isExpanded) => {
      let changedAlerts = [];
      const teamChange = newAlerts.team !== existingAlerts.team;
      for (const [metric, newAlert] of Object.entries(newAlerts.alerts)) {
        const oldAlert = existingAlerts.alerts[metric];
        const method = alertMethod(
          oldAlert,
          isExpanded[metric] && newAlert,
          teamChange
        );
        switch (method) {
          case "create":
            changedAlerts.push({
              ...newAlert,
              method: "create",
              metric: metric
            });
            break;
          case "delete":
            changedAlerts.push({
              ...oldAlert,
              method: "delete"
            });
            break;
          case "update":
            changedAlerts.push({
              ...newAlert,
              method: "update"
            });
            break;
          case "recreate":
            changedAlerts.push(
              {
                ...newAlert,
                method: "create",
                metric: metric
              },
              {
                ...oldAlert,
                method: "delete"
              }
            );
            break;
          default:
        }
      }
      if (changedAlerts.length > 0) {
        setRequest({
          team: newAlerts.team,
          alerts: changedAlerts
        });
      } else {
        onCancel(); // Form not changed
      }
    },
    [onCancel]
  );

  return { submitAlerts };
};
