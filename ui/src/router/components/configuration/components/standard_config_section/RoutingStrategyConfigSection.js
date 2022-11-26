import React from "react";
import { EuiDescriptionList } from "@elastic/eui";
import { ConfigSectionPanel } from "../../../../../components/config_section";
import { routingStrategyOptions } from "../../../form/components/ensembler_config/standard_ensembler/typeOptions";

export const RoutingStrategyConfigSection = ({ isLazyRouting }) => {
  const strategy = routingStrategyOptions.find(e => Boolean(isLazyRouting).toString() === e.flag) || {};
  const items = [
    {
      title: "Strategy",
      description: strategy.inputDisplay,
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
