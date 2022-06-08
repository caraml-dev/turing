import React, { Fragment, useContext, useState } from "react";
import { EuiFlexGroup, EuiFlexItem, EuiPanel } from "@elastic/eui";

import { ConfigSectionPanel } from "../../../../components/config_section";
import { RemoteComponent } from "../../../../components/remote_component/RemoteComponent";
import { LoadDynamicScript } from "../../../../hooks/useDynamicScript";
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

const CustomExperimentConfigView = ({ projectId, remoteUi, config }) => {
  const [urlReady, setUrlReady] = useState(false);
  const [urlFailed, setUrlFailed] = useState(false);
  const [configReady, setConfigReady] = useState(false);
  const [configFailed, setConfigFailed] = useState(false);

  // Retrieve script from host dynamically
  if (!!remoteUi.url && !urlReady) {
    return urlFailed ? (
      <FallbackView text={"Failed to load Experiment Engine"} />
    ) : (
      <>
        <LoadDynamicScript
          setReady={setUrlReady}
          setFailed={setUrlFailed}
          url={remoteUi.url}
        />
        <FallbackView text={"Loading Experiment Engine..."} />
      </>
    );
  } else if (!!remoteUi.config && !configReady) {
    return configFailed ? (
      <FallbackView text={"Failed to load Experiment Engine Config"} />
    ) : (
      <>
        <LoadDynamicScript
          setReady={setConfigReady}
          setFailed={setConfigFailed}
          url={remoteUi.config}
        />
        <FallbackView text={"Loading Experiment Engine Config..."} />
      </>
    );
  }

  // Load component from remote host
  return (
    <React.Suspense
      fallback={<FallbackView text="Loading Experiment Engine config" />}>
      <RemoteComponent
        scope={remoteUi.name}
        name="./ExperimentEngineConfigDetails"
        fallback={<FallbackView text="Loading Experiment Engine config" />}
        projectId={projectId}
        config={config}
      />
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
