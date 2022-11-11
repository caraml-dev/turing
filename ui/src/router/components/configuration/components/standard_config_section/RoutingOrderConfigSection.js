import React from "react";
import { EuiDescriptionList } from "@elastic/eui";
import { ConfigSectionPanel } from "../../../../../components/config_section";


export const RoutingOrderConfigSection = ({ isLazyRouting }) => {
  const items = [
    {
      title: "Lazy Routing",
      description: isLazyRouting,
    },
  ];
  return (<ConfigSectionPanel title={"Routing Order"}>
    <EuiDescriptionList
      compressed
      textStyle="reverse"
      type="responsiveColumn"
      listItems={items}
      titleProps={{ style: { width: "30%" } }}
      descriptionProps={{ style: { width: "70%" } }}
    />
  </ConfigSectionPanel>);
};
