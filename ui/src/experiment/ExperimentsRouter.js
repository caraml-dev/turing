import React, { Fragment } from "react";
import {
  EuiPageTemplate,
} from "@elastic/eui";
import { useParams } from "react-router-dom";

import { RemoteComponent } from "../components/remote_component/RemoteComponent";
import ExperimentEngineComponentLoader from "../components/remote_component/ExperimentEngineComponentLoader";

import { useConfig } from "../config";

const FallbackView = ({ text }) => {
  const { appConfig } = useConfig();

  return (
    <EuiPageTemplate restrictWidth={appConfig.pageTemplate.restrictWidth} paddingSize={appConfig.pageTemplate.paddingSize}>
      <EuiPageTemplate.EmptyPrompt
        iconType={appConfig.appIcon}
        title={<h2>Experiments</h2>}
        body={
          <Fragment>
            <p>{text}</p>
          </Fragment>
        }
      />
    </EuiPageTemplate>
  );
};

const RemoteRouter = ({ projectId }) => {
  const { defaultExperimentEngine } = useConfig();

  // Load component from remote host
  return (
    <React.Suspense
      fallback={<FallbackView text="Loading Experiment Engine config" />}>
      <ExperimentEngineComponentLoader
        FallbackView={FallbackView}
        remoteUi={defaultExperimentEngine}
        componentName="Experiment Engine"
      >
        <RemoteComponent
          scope={defaultExperimentEngine.name}
          name="./ExperimentsLandingPage"
          fallback={<FallbackView text="Loading Experiment Engine" />}
          projectId={projectId}
        />
      </ExperimentEngineComponentLoader>
    </React.Suspense>
  );
};

export const ExperimentsRouter = () => {
  const { projectId } = useParams();
  const { defaultExperimentEngine } = useConfig();
  return !!defaultExperimentEngine.url ? (
    <RemoteRouter projectId={projectId} />
  ) : (
    <FallbackView text="No default Experiment Engine configured" />
  );
};
