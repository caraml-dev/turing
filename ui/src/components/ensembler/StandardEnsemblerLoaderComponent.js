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
export const StandardEnsemblerLoaderComponent = ({
  FallbackView,
  experimentEngine,
  children,
}) => {
  const [urlReady, setUrlReady] = useState(false);
  const [urlFailed, setUrlFailed] = useState(false);
  const [configReady, setConfigReady] = useState(false);
  const [configFailed, setConfigFailed] = useState(false);

  return urlFailed ? (
    <FallbackView text={"Failed to load Standard Ensembler for the selected Custom Experiment Engine"} />
  ) : configFailed ? (
    <FallbackView text={"Failed to load Standard Ensembler config for the selected Custom Experiment Engine"} />
  ) : !urlReady || !configReady ? (
    <>
      {!!experimentEngine.url && !urlReady && (
        <LoadDynamicScript
          setReady={setUrlReady}
          setFailed={setUrlFailed}
          url={experimentEngine.url}
        />
      )}
      {!!experimentEngine.config && !configReady && (
        <LoadDynamicScript
          setReady={setConfigReady}
          setFailed={setConfigFailed}
          url={experimentEngine.config}
        />
      )}
      <FallbackView text={"Loading Standard Ensembler for the selected Custom Experiment Engine..."} />
    </>
  ) : (
    children
  );
};

export default StandardEnsemblerLoaderComponent;
