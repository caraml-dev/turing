import React, { Fragment, useContext } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiPanel } from "@elastic/eui";

import { ConfigSectionPanel } from "../../../../components/config_section";
import { RemoteComponent } from "../../../../components/remote_component/RemoteComponent";
import ExperimentEngineComponentLoader from "../../../../components/remote_component/ExperimentEngineComponentLoader";
import ExperimentEngineContext from "../../../../providers/experiments/context";

import { StandardExperimentConfigGroup } from "./experiment_config_section/StandardExperimentConfigGroup";

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

const CustomExperimentConfigView = ({
  projectId,
  remoteUi,
  config,
  protocol,
}) => {
  // Load component from remote host
  return (
    <React.Suspense
      fallback={<FallbackView text="Loading Experiment Engine config" />}
    >
      <ExperimentEngineComponentLoader
        FallbackView={FallbackView}
        remoteUi={remoteUi}
        componentName="Experiment Engine"
      >
        <RemoteComponent
          scope={remoteUi.name}
          name="./ExperimentEngineConfigDetails"
          fallback={<FallbackView text="Loading Experiment Engine config" />}
          projectId={projectId}
          config={config}
          protocol={protocol}
        />
      </ExperimentEngineComponentLoader>
    </React.Suspense>
  );
};

export const ExperimentConfigSection = ({
  projectId,
  config: { protocol, experiment_engine },
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
            protocol={protocol}
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
