import React from "react";
import {
  EuiPage,
  EuiPageBody,
  EuiPageContent,
  EuiPageHeader,
  EuiPageHeaderSection,
  EuiText,
} from "@elastic/eui";
import { PageTitle } from "../components/page/PageTitle";
import { RemoteComponent } from "../components/remote_component/RemoteComponent";
import { ExperimentEngineLoaderComponent } from "../components/experiments/ExperimentEngineLoaderComponent";

import { useConfig } from "../config";

const FallbackView = ({ text }) => (
  <EuiPage>
    <EuiPageBody>
      <EuiPageHeader>
        <EuiPageHeaderSection>
          <PageTitle title="Experiments" />
        </EuiPageHeaderSection>
      </EuiPageHeader>
      <EuiPageContent>
        <EuiText size="s" color="subdued">
          {text}
        </EuiText>
      </EuiPageContent>
    </EuiPageBody>
  </EuiPage>
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
