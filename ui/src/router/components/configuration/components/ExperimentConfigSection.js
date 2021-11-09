import React, { Fragment } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiPanel } from "@elastic/eui";

import { StandardExperimentConfigGroup } from "./experiment_config_section/StandardExperimentConfigGroup";
import { ExperimentEngineContextProvider } from "../../../../providers/experiments/ExperimentEngineContextProvider";
import { ConfigSectionPanel } from "../../../../components/config_section";
import useDynamicScript from "../../../../hooks/useDynamicScript";
import loadComponent from "../../../../utils/remoteComponent";

const StandardExperimentConfigView = ({ projectId, experiment_engine }) => (
  <ExperimentEngineContextProvider>
    <StandardExperimentConfigGroup
      projectId={projectId}
      engineType={experiment_engine.type}
      engineConfig={experiment_engine.config}
    />
  </ExperimentEngineContextProvider>
);

const FallbackView = ({ text }) => (
  <EuiFlexGroup direction="row" wrap>
    <EuiFlexItem grow={true}>
      <ConfigSectionPanel title="Experiment Engine">{text}</ConfigSectionPanel>
    </EuiFlexItem>
  </EuiFlexGroup>
);

const CustomExperimentConfigView = ({ projectId, remoteUi, config }) => {
  // Retrieve script from host dynamically
  const { ready, failed } = useDynamicScript({
    url: remoteUi.url,
  });

  if (!ready || failed) {
    const text = failed
      ? "Failed to load Experiment Engine"
      : "Loading Experiment Engine ...";
    return <FallbackView text={text} />;
  }

  // Load component from remote host
  const ExperimentEngineConfigDetails = React.lazy(
    loadComponent(remoteUi.name, "./ExperimentEngineConfigDetails")
  );

  return (
    <React.Suspense
      fallback={<FallbackView text="Loading Experiment Engine config" />}>
      <ExperimentEngineConfigDetails projectId={projectId} config={config} />
    </React.Suspense>
  );
};

export const ExperimentConfigSection = ({
  projectId,
  config: { experiment_engine },
}) => (
  <Fragment>
    {experiment_engine.type === "nop" ? (
      <EuiPanel>Not Configured</EuiPanel>
    ) : !!experiment_engine.custom_experiment_manager_config ? (
      <CustomExperimentConfigView
        projectId={projectId}
        remoteUi={experiment_engine.custom_experiment_manager_config.remote_ui}
        config={experiment_engine.config}
      />
    ) : (
      <StandardExperimentConfigView
        projectId={projectId}
        experiment_engine={experiment_engine}
      />
    )}
  </Fragment>
);
