import React, { useEffect } from "react";
import { replaceBreadcrumbs } from "@gojek/mlp-ui";
import { RouterConfigDetails } from "../../../components/configuration/RouterConfigDetails";
import { get } from "../../../../components/form/utils";
import { ExperimentEngineContextProvider } from "../../../../providers/experiments/ExperimentEngineContextProvider";

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

  return (
    <ExperimentEngineContextProvider>
      <RouterConfigDetails projectId={projectId} config={config} />
    </ExperimentEngineContextProvider>
  );
};
