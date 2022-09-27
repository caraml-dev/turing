import React, { useContext } from "react";
import { EuiFlexItem, EuiSpacer } from "@elastic/eui";

import { RemoteComponent } from "../../../../../components/remote_component/RemoteComponent";
import ExperimentEngineContext from "../../../../../providers/experiments/context";
import ExperimentEngineComponentLoader from "../../../../../components/remote_component/ExperimentEngineComponentLoader";
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
  // Load component from remote host
  return (
    <React.Suspense
      fallback={<FallbackView text="Loading Experiment Engine config" />}>
      <ExperimentEngineComponentLoader
        FallbackView={FallbackView}
        remoteUi={remoteUi}
        componentName="Experiment Engine"
      >
        <RemoteComponent
          scope={remoteUi.name}
          name="./EditExperimentEngineConfig"
          fallback={<FallbackView text="Loading Experiment Engine config" />}
          projectId={projectId}
          config={config}
          onChangeHandler={onChangeHandler}
          errors={errors}
        />
      </ExperimentEngineComponentLoader>
    </React.Suspense>
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
