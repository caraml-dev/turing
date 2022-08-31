import React from "react";
import {
  EuiPageTemplate, EuiSpacer,
  EuiText,
} from "@elastic/eui";
import { PageTitle } from "../components/page/PageTitle";
import { RemoteComponent } from "../components/remote_component/RemoteComponent";
import { ExperimentEngineLoaderComponent } from "../components/experiments/ExperimentEngineLoaderComponent";

import { useConfig } from "../config";

const FallbackView = ({ text }) => (
  <EuiPageTemplate restrictWidth="90%" paddingSize={"none"}>
    <EuiSpacer size="l" />
    <EuiPageTemplate.Header
      bottomBorder={false}
      pageTitle={<PageTitle title="Experiments" />}
    />

    <EuiSpacer size="m" />
    <EuiPageTemplate.Section restrictWidth="90%" color={"transparent"}>
      <EuiText size="s" color="subdued">
        {text}
      </EuiText>
    </EuiPageTemplate.Section>
  </EuiPageTemplate>
);

const RemoteRouter = ({ projectId }) => {
  const { defaultExperimentEngine } = useConfig();

  // Load component from remote host
  return (
    <React.Suspense
      fallback={<FallbackView text="Loading Experiment Engine config" />}>
      <ExperimentEngineLoaderComponent
        FallbackView={FallbackView}
        experimentEngine={defaultExperimentEngine}>
        <RemoteComponent
          scope={defaultExperimentEngine.name}
          name="./ExperimentsLandingPage"
          fallback={<FallbackView text="Loading Experiment Engine" />}
          projectId={projectId}
        />
      </ExperimentEngineLoaderComponent>
    </React.Suspense>
  );
};

export const ExperimentsRouter = ({ projectId }) => {
  const { defaultExperimentEngine } = useConfig();
  return !!defaultExperimentEngine.url ? (
    <RemoteRouter projectId />
  ) : (
    <FallbackView text="No default Experiment Engine configured" />
  );
};
