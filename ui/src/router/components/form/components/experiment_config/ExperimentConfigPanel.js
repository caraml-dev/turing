import React from "react";
import { EuiFlexItem, EuiSpacer } from "@elastic/eui";

import useDynamicScript from "../../../../../hooks/useDynamicScript";
import loadComponent from "../../../../../utils/remoteComponent";
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

const CustomExperimentEngineConfig = ({
  projectId,
  remoteUi,
  config,
  onChangeHandler,
  errors,
}) => {
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
  const EditExperimentEngineConfig = React.lazy(
    loadComponent(remoteUi.name, "./EditExperimentEngineConfig")
  );

  return (
    <React.Suspense
      fallback={<FallbackView text="Loading Experiment Engine config" />}>
      <EditExperimentEngineConfig
        projectId={projectId}
        config={config}
        onChangeHandler={onChangeHandler}
        errors={errors}
      />
    </React.Suspense>
  );
};

export const ExperimentConfigPanel = ({
  projectId,
  engine,
  onChangeHandler,
  errors,
}) =>
  !!engine.custom_experiment_manager_config ? (
    <CustomExperimentEngineConfig
      projectId={projectId}
      remoteUi={engine.custom_experiment_manager_config.remote_ui}
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
  );
