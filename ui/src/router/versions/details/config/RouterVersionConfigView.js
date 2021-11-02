import React, { useEffect } from "react";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { RouterConfigDetails } from "../../../components/configuration/RouterConfigDetails";
import { get } from "../../../../components/form/utils";

export const RouterVersionConfigView = ({ projectId, config }) => {
  useEffect(() => {
    replaceBreadcrumbs([
      {
        text: "Routers",
        href: `../../../`,
      },
      {
        text: get(config, "router.name") || "",
        href: `../../`,
      },
      {
        text: "Versions",
        href: `../`,
      },
      {
        text: `Version ${config.version}`,
        href: `./`,
      },
    ]);
  }, [config]);

  return <RouterConfigDetails projectId={projectId} config={config} />;
};
