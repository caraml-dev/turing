import React, { useEffect, useState } from "react";
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
import useDynamicScript, { LoadDynamicScript } from "../hooks/useDynamicScript";

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
  const [configStatusReady, setConfigStatusReady] = useState(false);
  const [configStatusFailed, setConfigStatusFailed] = useState(false);
  const [configStatusLoaded, setConfigStatusLoaded] = useState(false);

  // Retrieve script from host dynamically
  const { ready, failed } = useDynamicScript({
    url: defaultExperimentEngine.url,
  });

  // Re-render to get updated config loading status
  useEffect(() => {
    return () => {
      setConfigStatusLoaded(false);
    };
  });

  if (!ready || failed) {
    const text = failed
      ? "Failed to load Experiment Engine"
      : "Loading Experiment Engine ...";
    return <FallbackView text={text} />;
  } else if (
    defaultExperimentEngine.config &&
    configStatusLoaded &&
    (!configStatusReady || configStatusFailed)
  ) {
    const text = configStatusFailed
      ? "Failed to load Experiment Engine Config"
      : "Loading Experiment Engine Config...";
    return <FallbackView text={text} />;
  }

  // Load component from remote host
  return (
    <React.Suspense
      fallback={<FallbackView text="Loading Experiment Engine config" />}>
      <LoadDynamicScript
        setConfigStatusReady={setConfigStatusReady}
        setConfigStatusFailed={setConfigStatusFailed}
        setConfigStatusLoaded={setConfigStatusLoaded}
        url={defaultExperimentEngine.config}
      />
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
