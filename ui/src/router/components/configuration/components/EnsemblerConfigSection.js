import React, { Fragment } from "react";
import { EuiPanel } from "@elastic/eui";
import { DockerConfigViewGroup } from "./docker_config_section/DockerConfigViewGroup";
import { TreatmentMappingConfigSection } from "./TreatmentMappingConfigSection";
import { ExperimentEngineContextProvider } from "../../../../providers/experiments/ExperimentEngineContextProvider";

export const EnsemblerConfigSection = ({
  config: {
    ensembler,
    experiment_engine: { type, config: experimentConfig },
  },
}) => {
  return !ensembler ? (
    <EuiPanel>Not Configured</EuiPanel>
  ) : (
    <Fragment>
      {ensembler.type === "docker" && (
        <DockerConfigViewGroup
          componentName="Ensembler"
          dockerConfig={ensembler.docker_config}
        />
      )}
      {ensembler.type === "standard" && (
        <ExperimentEngineContextProvider>
          <TreatmentMappingConfigSection
            engine={type}
            experiments={(experimentConfig || {}).experiments || []}
            mappings={ensembler.standard_config.experiment_mappings}
          />
        </ExperimentEngineContextProvider>
      )}
    </Fragment>
  );
};
