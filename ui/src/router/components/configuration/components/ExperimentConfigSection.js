import React, { Fragment } from "react";
import { EuiPanel } from "@elastic/eui";
import { ExperimentConfigGroup } from "./experiment_config_section/ExperimentConfigGroup";
import { ExperimentEngineContextProvider } from "../../../../providers/experiments/ExperimentEngineContextProvider";

export const ExperimentConfigSection = ({
  projectId,
  config: { experiment_engine },
}) => {
  return (
    <Fragment>
      {experiment_engine.type === "nop" ? (
        <EuiPanel>Not Configured</EuiPanel>
      ) : (
        <ExperimentEngineContextProvider>
          <ExperimentConfigGroup
            projectId={projectId}
            engineType={experiment_engine.type}
            engineConfig={experiment_engine.config}
          />
        </ExperimentEngineContextProvider>
      )}
    </Fragment>
  );
};
