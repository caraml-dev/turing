import React from "react";
import { EuiDescriptionList } from "@elastic/eui";

export const ResourcesConfigTable = ({
  resourceRequest: { cpu_request, memory_request, min_replica, max_replica },
}) => {
  const items = [
    {
      title: "CPU Request",
      description: cpu_request,
    },
    {
      title: "Memory Request",
      description: memory_request,
    },
    {
      title: "Min Replicas",
      description: min_replica,
    },
    {
      title: "Max Replicas",
      description: max_replica,
    },
  ];

  return (
    <EuiDescriptionList
      compressed
      textStyle="reverse"
      type="responsiveColumn"
      listItems={items}
    />
  );
};
