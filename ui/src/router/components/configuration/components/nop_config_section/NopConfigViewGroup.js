import React from "react";
import { RouteConfigSection } from "../route_config_section/RouteConfigSection";

export const NopConfigViewGroup = ({ nopConfig }) => {
  const items = [
    {
      title: "Final Response Route",
      description: nopConfig.final_response_route_id,
    },
  ];
  return <RouteConfigSection panelTitle={"Response"} items={items} />;
};
