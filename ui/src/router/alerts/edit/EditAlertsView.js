import React, { useEffect } from "react";
import {
  addToast,
  replaceBreadcrumbs,
  FormContextProvider,
} from "@gojek/mlp-ui";
import { useNavigate } from "react-router-dom";
import { EditAlertsForm } from "../components/edit_alerts_form/components/EditAlertsForm";
import { TuringAlerts } from "../../../services/alerts/TuringAlerts";
import { useConfig } from "../../../config";

export const EditAlertsView = ({ projectId, router, alerts }) => {
  const navigate = useNavigate();
  const { alertConfig } = useConfig();

  useEffect(() => {
    replaceBreadcrumbs([
      {
        text: "Routers",
        href: `../../../`,
      },
      {
        text: router.name,
        href: `../`,
      },
      {
        text: "Alerts",
        href: "./",
      },
      {
        text: "Edit",
      },
    ]);
  }, [router]);

  const onSuccess = (isSubmitSuccess) => {
    if (isSubmitSuccess) {
      addToast({
        id: `submit-success-alerts`,
        title: `Alerts have been updated!`,
        color: "success",
        iconType: "check",
      });
    }
    navigate("../", { state: { refresh: true } });
  };

  return (
    <FormContextProvider data={TuringAlerts.fromJson(alerts)}>
      <EditAlertsForm
        existingData={alerts}
        onCancel={() => navigate("../")}
        onSuccess={onSuccess}
        environment={alertConfig.environment}
        projectId={projectId}
        routerId={router.id}
      />
    </FormContextProvider>
  );
};
