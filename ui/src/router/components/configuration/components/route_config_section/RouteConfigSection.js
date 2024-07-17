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
      columnWidths={[1, 7/3]}
    />
  </ConfigSectionPanel>
);
