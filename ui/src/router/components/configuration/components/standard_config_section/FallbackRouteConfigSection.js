import React from "react";
import { RouteConfigSection } from "../route_config_section/RouteConfigSection";

export const FallbackRouteConfigSection = ({ fallbackResponseRouteId }) => {
  const items = [
    {
      title: "Fallback Response Route",
      description: fallbackResponseRouteId,
    },
  ];
  return <RouteConfigSection panelTitle={"Fallback"} items={items} />;
};
