import React from "react";
import { EuiDescriptionList } from "@elastic/eui";
import { ConfigSectionPanel } from "../../../../../components/config_section";

export const NopConfigViewGroup = ({ nopConfig }) => {
  const items = [
    {
      title: "Final Response Route",
      description: nopConfig.final_response_route_id,
    },
  ];
  return (
    <ConfigSectionPanel title="Response">
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
};
