import React, { Fragment } from "react";
import { DockerConfigViewGroup } from "./docker_config_section/DockerConfigViewGroup";
import { ExperimentEngineContextProvider } from "../../../../providers/experiments/ExperimentEngineContextProvider";
import { NopConfigViewGroup } from "./nop_config_section/NopConfigViewGroup";
import { PyFuncConfigViewGroup } from "./pyfunc_config_section/PyFuncConfigViewGroup";
import { StandardConfigViewGroup } from "./standard_config_section/StandardConfigViewGroup"
import { EnsemblersContextContextProvider } from "../../../../providers/ensemblers/context";

export const EnsemblerConfigSection = ({
  projectId,
  config: {
    routes,
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
        <StandardConfigViewGroup
          projectId={projectId}
          routes={routes}
          standardConfig={ensembler.standard_config}
          experimentConfig={(experimentConfig || {}).experiments || []}
          type={type}
        />
      </ExperimentEngineContextProvider>
    )}
  </Fragment>
);
