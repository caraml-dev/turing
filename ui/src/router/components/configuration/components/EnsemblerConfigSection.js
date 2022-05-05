import React, { Fragment } from "react";
import { DockerConfigViewGroup } from "./docker_config_section/DockerConfigViewGroup";
import { FallbackRouteConfigSection } from "./standard_config_section/FallbackRouteConfigSection";
import { TreatmentMappingConfigSection } from "./standard_config_section/TreatmentMappingConfigSection";
import { ExperimentEngineContextProvider } from "../../../../providers/experiments/ExperimentEngineContextProvider";
import { NopConfigViewGroup } from "./nop_config_section/NopConfigViewGroup";
import { PyFuncConfigViewGroup } from "./pyfunc_config_section/PyFuncConfigViewGroup";
import { EnsemblersContextContextProvider } from "../../../../providers/ensemblers/context";
import { EuiFlexGroup, EuiFlexItem } from "@elastic/eui";

export const EnsemblerConfigSection = ({
  projectId,
  config: {
    ensembler,
    experiment_engine: { type, config: experimentConfig },
  },
}) => (
  <Fragment>
    {ensembler.type === "nop" && (
      <NopConfigViewGroup nopConfig={ensembler.nop_config} />
    )}
    {ensembler.type === "pyfunc" && (
      <EnsemblersContextContextProvider
        projectId={projectId}
        ensemblerType={"pyfunc"}>
        <PyFuncConfigViewGroup
          componentName="Ensembler"
          pyfuncConfig={ensembler.pyfunc_config}
          dockerConfig={ensembler.docker_config}
        />
      </EnsemblersContextContextProvider>
    )}
    {ensembler.type === "docker" && (
      <DockerConfigViewGroup
        componentName="Ensembler"
        dockerConfig={ensembler.docker_config}
      />
    )}
    {ensembler.type === "standard" && (
      <ExperimentEngineContextProvider>
        <EuiFlexGroup direction="column">
          <EuiFlexItem>
            <TreatmentMappingConfigSection
              engine={type}
              experiments={(experimentConfig || {}).experiments || []}
              mappings={ensembler.standard_config.experiment_mappings}
            />
          </EuiFlexItem>
          <EuiFlexItem>
            <FallbackRouteConfigSection
              fallbackResponseRouteId={
                ensembler.standard_config.fallback_response_route_id
              }
            />
          </EuiFlexItem>
        </EuiFlexGroup>
      </ExperimentEngineContextProvider>
    )}
  </Fragment>
);
