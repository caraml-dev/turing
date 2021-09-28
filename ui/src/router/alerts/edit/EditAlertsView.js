import React, { useEffect } from "react";
import { addToast, replaceBreadcrumbs } from "@gojek/mlp-ui";
import { FormContextProvider } from "../../../components/form/context";
import { EditAlertsForm } from "../components/edit_alerts_form/components/EditAlertsForm";
import { TuringAlerts } from "../../../services/alerts/TuringAlerts";
import { useConfig } from "../../../config";

export const EditAlertsView = ({ router, alerts, ...props }) => {
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
    props.navigate("../", { state: { refresh: true } });
  };

  return (
    <FormContextProvider data={TuringAlerts.fromJson(alerts)}>
      <EditAlertsForm
        existingData={alerts}
        onCancel={() => props.navigate("../")}
        onSuccess={onSuccess}
        environment={alertConfig.environment}
        {...props}
      />
    </FormContextProvider>
  );
};
