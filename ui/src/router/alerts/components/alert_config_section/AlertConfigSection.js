import React from "react";
import { EuiDescriptionList } from "@elastic/eui";
import "./AlertConfigSection.scss";

export const AlertConfigSection = ({
  alert: { warning_threshold, critical_threshold, duration },
  unit,
}) => {
  const items = [
    {
      title: "Warning Threshold",
      description: warning_threshold ? warning_threshold + unit : "-",
    },
    {
      title: "Critical Threshold",
      description: critical_threshold ? critical_threshold + unit : "-",
    },
    {
      title: "Duration",
      description: duration,
    },
  ];

  return (
    <EuiDescriptionList
      className="euiDescriptionList--alertConfigSection"
      compressed
      align="center"
      textStyle="reverse"
      listItems={items}
    />
  );
};
