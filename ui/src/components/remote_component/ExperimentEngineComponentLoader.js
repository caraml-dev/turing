import React, { useEffect, useState } from "react";

import useDynamicScript from "../../hooks/useDynamicScript";

// Renderless component wrapper
const LoadDynamicScript = ({ url, setReady, setFailed }) => {
  const { ready, failed } = useDynamicScript({
    url: url,
  });

  useEffect(() => {
    setReady(ready);
    setFailed(failed);
  }, [setReady, setFailed, ready, failed]);

  return null;
};

// Dynamic Script Loading component wrapper
export const ExperimentEngineComponentLoader = ({
  FallbackView,
  remoteUi,
  componentName,
  children,
}) => {
  const [urlReady, setUrlReady] = useState(false);
  const [urlFailed, setUrlFailed] = useState(false);
  const [configReady, setConfigReady] = useState(false);
  const [configFailed, setConfigFailed] = useState(false);

  return urlFailed ? (
    <FallbackView text={`Failed to load ${componentName}`} />
  ) : configFailed ? (
    <FallbackView text={`Failed to load ${componentName} Config`} />
  ) : !urlReady || !configReady ? (
    <>
      {!!remoteUi.url && !urlReady && (
        <LoadDynamicScript
          setReady={setUrlReady}
          setFailed={setUrlFailed}
          url={remoteUi.url}
        />
      )}
      {!!remoteUi.config && !configReady && (
        <LoadDynamicScript
          setReady={setConfigReady}
          setFailed={setConfigFailed}
          url={remoteUi.config}
        />
      )}
      <FallbackView text={`Loading ${componentName}...`} />
    </>
  ) : (
    children
  );
};

export default ExperimentEngineComponentLoader;
