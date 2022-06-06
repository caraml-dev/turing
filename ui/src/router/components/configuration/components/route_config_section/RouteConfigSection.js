import React from "react";
import { EuiDescriptionList } from "@elastic/eui";
import { ConfigSectionPanel } from "../../../../../components/config_section";

export const RouteConfigSection = ({ panelTitle, items }) => (
  <ConfigSectionPanel title={panelTitle}>
    <EuiDescriptionList
      compressed
      textStyle="reverse"
      type="responsiveColumn"
      listItems={items}
      titleProps={{ style: { width: "30%" } }}
      descriptionProps={{ style: { width: "70%" } }}
    />
  </ConfigSectionPanel>
);
