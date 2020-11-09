import React, { useCallback, useContext, useState } from "react";
import { MetricPanel } from "./metric_panel/MetricPanel";
import { ConfigSectionTitle } from "../../../../components/configuration/components/section";
import { AccordionForm } from "../../../../../components/accordion_form";
import { TeamPanel } from "./team_panel/TeamPanel";
import { FormContext } from "../../../../../components/form/context";
import schema from "../validation/schema";
import { TeamsProvider } from "../../../../../providers/teams/TeamsProvider";
import { supportedAlerts } from "../../../config";
import FormValidationContext from "../../../../../components/form/validation";
import { get } from "../../../../../components/form/utils";
import { useOnChangeHandler } from "../../../../../components/form/hooks/useOnChangeHandler";
import { useAlertsApi } from "../hooks/useAlertsApi";

export const EditAlertsForm = ({
  existingData,
  onCancel,
  onSuccess,
  ...props
}) => {
  const { data: newAlerts, onChangeHandler } = useContext(FormContext);
  const { onChange } = useOnChangeHandler(onChangeHandler);

  const { submitAlerts } = useAlertsApi(
    props.projectId,
    props.routerId,
    props.environment,
    onCancel,
    onSuccess
  );

  const [isExpanded, setIsExpanded] = useState(
    supportedAlerts.reduce((acc, alert) => {
      acc[alert.metric] = !!existingData.alerts[alert.metric];
      return acc;
    }, {})
  );

  const toggleExpanded = useCallback(
    metric => {
      return () =>
        setIsExpanded(state => ({
          ...state,
          [metric]: !state[metric]
        }));
    },
    [setIsExpanded]
  );

  const sections = [
    {
      title: "Team",
      iconType: "user",
      children: (
        <FormValidationContext.Consumer>
          {({ errors }) => (
            <TeamsProvider>
              <TeamPanel
                team={newAlerts.team}
                onChange={onChange("team")}
                errors={(errors || {}).team}
              />
            </TeamsProvider>
          )}
        </FormValidationContext.Consumer>
      ),
      validationSchema: schema["team"]
    },
    ...supportedAlerts.map((alertType, idx) => ({
      title: alertType.title,
      iconType: alertType.iconType,
      children: (
        <FormValidationContext.Consumer>
          {({ errors }) => (
            <MetricPanel
              title={alertType.title}
              comparator={alertType.comparator}
              unit={alertType.unit}
              alert={newAlerts.alerts[alertType.metric]}
              onChangeHandler={onChange(`alerts.${alertType.metric}`)}
              errors={get(errors, `alerts.${alertType.metric}`)}
              isExpanded={isExpanded[alertType.metric]}
              toggleExpanded={toggleExpanded(alertType.metric)}
            />
          )}
        </FormValidationContext.Consumer>
      ),
      validationSchema: isExpanded[alertType.metric] && schema[alertType.metric]
    }))
  ];

  return (
    <AccordionForm
      name="Configure Alerts"
      sections={sections}
      onCancel={onCancel}
      onSubmit={() => submitAlerts(existingData, newAlerts, isExpanded)}
      submitLabel="Update"
      renderTitle={(title, iconType) => (
        <ConfigSectionTitle title={title} iconType={iconType} />
      )}
    />
  );
};
