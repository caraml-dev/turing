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

import useDynamicScript from "../hooks/useDynamicScript";
import loadComponent from "../utils/remoteComponent";
import { defaultExperimentEngine } from "../config";

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
  // Retrieve script from host dynamically
  const { ready, failed } = useDynamicScript({
    url: defaultExperimentEngine.url,
  });

  if (!ready || failed) {
    const text = !ready
      ? "Loading Experiment Engine ..."
      : "Failed to load Experiment Engine";
    return <FallbackView text={text} />;
  }

  // Load component from remote host
  const ExperimentsLandingPage = React.lazy(
    loadComponent(defaultExperimentEngine.name, "./ExperimentsLandingPage")
  );

  return (
    <React.Suspense fallback={<FallbackView text="Loading Experiments" />}>
      <ExperimentsLandingPage projectId={projectId} />
    </React.Suspense>
  );
};

export const ExperimentsRouter = ({ projectId }) =>
  !!defaultExperimentEngine.url ? (
    <RemoteRouter projectId />
  ) : (
    <FallbackView text="No default Experiment Engine configured" />
  );
