import React, { Fragment } from "react";
import { EuiPanel } from "@elastic/eui";
import { ExperimentConfigGroup } from "./experiment_config_section/ExperimentConfigGroup";
import { ExperimentEngineContextProvider } from "../../../../providers/experiments/ExperimentEngineContextProvider";

export const ExperimentConfigSection = ({ config: { experiment_engine } }) => {
  const View = React.lazy(() =>
    import("expEngine/ExperimentEngineConfigDetails")
  );

  return (
    /* This a temporary hard-coded check. It should be replaced by a check
       for the default engine property. */
    experiment_engine.type === "xp" ? (
      <React.Suspense fallback="Loading Experiments">
        <View
          client={experiment_engine.config.client}
          experiments={experiment_engine.config.experiments}
          variables={experiment_engine.config.variables}
        />
      </React.Suspense>
    ) : (
      <Fragment>
        {experiment_engine.type === "nop" ? (
          <EuiPanel>Not Configured</EuiPanel>
        ) : (
          <ExperimentEngineContextProvider>
            <ExperimentConfigGroup
              engineType={experiment_engine.type}
              engineConfig={experiment_engine.config}
            />
          </ExperimentEngineContextProvider>
        )}
      </Fragment>
    )
  );
};
