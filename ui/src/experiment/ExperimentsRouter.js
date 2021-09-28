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

export const ExperimentsRouter = ({ projectId }) => {
  // Retrieve script from host dynamically
  const { ready, failed } = useDynamicScript({
    url: defaultExperimentEngine.url,
  });

  if (!ready || failed) {
    return (
      <EuiPage>
        <EuiPageBody>
          <EuiPageHeader>
            <EuiPageHeaderSection>
              <PageTitle title="Experiments" />
            </EuiPageHeaderSection>
          </EuiPageHeader>
          <EuiPageContent>
            <EuiText size="s" color="subdued">
              {!defaultExperimentEngine.url
                ? "No default Experiment Engine configured"
                : !ready
                ? "Loading Experiment Exgine ..."
                : "Failed to load Experiment Exgine"}
            </EuiText>
          </EuiPageContent>
        </EuiPageBody>
      </EuiPage>
    );
  }

  // Load component from remote host
  const Component = React.lazy(
    loadComponent(defaultExperimentEngine.appName, "./ExperimentsLandingPage")
  );

  return (
    <React.Suspense fallback="Loading Experiments">
      <Component projectId={projectId} />
    </React.Suspense>
  );
};
