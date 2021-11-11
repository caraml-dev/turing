import React, { Fragment, useContext } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiPanel } from "@elastic/eui";

import { ConfigSectionPanel } from "../../../../components/config_section";
import ExperimentEngineContext from "../../../../providers/experiments/context";
import { StandardExperimentConfigGroup } from "./experiment_config_section/StandardExperimentConfigGroup";
import useDynamicScript from "../../../../hooks/useDynamicScript";
import loadComponent from "../../../../utils/remoteComponent";

const StandardExperimentConfigView = ({ projectId, engine }) => (
  <StandardExperimentConfigGroup
    projectId={projectId}
    engineType={engine.type}
    engineConfig={engine.config}
  />
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
}) => {
  // Get engine's properties
  const { getEngineProperties } = useContext(ExperimentEngineContext);
  const engineProps = getEngineProperties(experiment_engine.type);

  return (
    <Fragment>
      {experiment_engine.type === "nop" ? (
        <EuiPanel>Not Configured</EuiPanel>
      ) : !!engineProps.type ? (
        engineProps.type === "custom" ? (
          <CustomExperimentConfigView
            projectId={projectId}
            remoteUi={engineProps.custom_experiment_manager_config.remote_ui}
            config={experiment_engine.config}
          />
        ) : (
          <StandardExperimentConfigView
            projectId={projectId}
            engine={experiment_engine}
          />
        )
      ) : (
        <EuiPanel>Loading ...</EuiPanel>
      )}
    </Fragment>
  );
};
