import React, { useState } from "react";
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
import { LoadDynamicScript } from "../hooks/useDynamicScript";

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
  const [urlReady, setUrlReady] = useState(false);
  const [urlFailed, setUrlFailed] = useState(false);
  const [configReady, setConfigReady] = useState(false);
  const [configFailed, setConfigFailed] = useState(false);

  // Retrieve script from host dynamically
  if (!!defaultExperimentEngine.url && !urlReady) {
    return urlFailed ? (
      <FallbackView text={"Failed to load Experiment Engine"} />
    ) : (
      <>
        <LoadDynamicScript
          setReady={setUrlReady}
          setFailed={setUrlFailed}
          url={defaultExperimentEngine.url}
        />
        <FallbackView text={"Loading Experiment Engine..."} />
      </>
    );
  } else if (!!defaultExperimentEngine.config && !configReady) {
    return configFailed ? (
      <FallbackView text={"Failed to load Experiment Engine Config"} />
    ) : (
      <>
        <LoadDynamicScript
          setReady={setConfigReady}
          setFailed={setConfigFailed}
          url={defaultExperimentEngine.config}
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
        scope={defaultExperimentEngine.name}
        name="./ExperimentsLandingPage"
        fallback={<FallbackView text="Loading Experiment Engine" />}
        projectId={projectId}
      />
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
