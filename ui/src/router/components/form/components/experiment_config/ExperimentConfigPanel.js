import React, { useContext, useState } from "react";
import { EuiFlexItem, EuiSpacer } from "@elastic/eui";

import { RemoteComponent } from "../../../../../components/remote_component/RemoteComponent";
import ExperimentEngineContext from "../../../../../providers/experiments/context";
import useDynamicScript, {
  LoadDynamicScript,
} from "../../../../../hooks/useDynamicScript";
import { Panel } from "../Panel";

import { StandardExperimentConfigGroup } from "./StandardExperimentConfigGroup";

const FallbackView = ({ text }) => (
  <EuiFlexItem grow={true}>
    <Panel title="Configuration">
      <EuiSpacer size="m" />
      {text}
    </Panel>
  </EuiFlexItem>
);

const CustomExperimentEngineConfigGroup = ({
  remoteUi,
  projectId,
  config,
  onChangeHandler,
  errors,
}) => {
  const [configReady, setConfigReady] = useState(false);
  const [configFailed, setConfigFailed] = useState(false);

  // Retrieve script from host dynamically
  const { ready, failed } = useDynamicScript({
    url: remoteUi.url,
  });

  if (!ready || failed) {
    const text = failed
      ? "Failed to load Experiment Engine"
      : "Loading Experiment Engine ...";
    return <FallbackView text={text} />;
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
    <RemoteComponent
      scope={remoteUi.name}
      name="./EditExperimentEngineConfig"
      fallback={<FallbackView text="Loading Experiment Engine config" />}
      projectId={projectId}
      config={config}
      onChangeHandler={onChangeHandler}
      errors={errors}
    />
  );
};

export const ExperimentConfigPanel = ({
  projectId,
  engine,
  onChangeHandler,
  errors,
}) => {
  // Get engine's properties
  const { getEngineProperties, isLoaded } = useContext(ExperimentEngineContext);
  const engineProps = getEngineProperties(engine.type);

  return isLoaded ? (
    engineProps.type === "custom" ? (
      <CustomExperimentEngineConfigGroup
        remoteUi={engineProps.custom_experiment_manager_config.remote_ui}
        projectId={projectId}
        config={engine.config}
        onChangeHandler={onChangeHandler}
        errors={errors}
      />
    ) : (
      <StandardExperimentConfigGroup
        engineType={engine.type}
        experimentConfig={engine.config}
        onChangeHandler={onChangeHandler}
        errors={errors}
      />
    )
  ) : (
    <FallbackView text={"Loading ..."} />
  );
};
