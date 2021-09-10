import React from "react";
import { EuiDescriptionList } from "@elastic/eui";

export const ContainerConfigTable = ({
  config: { image, endpoint, port, timeout, service_account },
}) => {
  const items = [
    {
      title: "Image",
      description: image,
    },
    {
      title: "Endpoint",
      description: endpoint,
    },
    {
      title: "Port",
      description: port,
    },
    {
      title: "Timeout",
      description: timeout,
    },
  ];

  if (service_account) {
    items.push({
      title: "Service Account",
      description: service_account,
    });
  }

  return (
    <EuiDescriptionList
      compressed
      textStyle="reverse"
      type="responsiveColumn"
      listItems={items}
      titleProps={{ style: { width: "30%" } }}
      descriptionProps={{ style: { width: "70%" } }}
    />
  );
};
