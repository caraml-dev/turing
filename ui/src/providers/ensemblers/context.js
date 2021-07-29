import React, { useEffect, useState } from "react";
import { useTuringApi } from "../../hooks/useTuringApi";

const EnsemblersContext = React.createContext([]);

export const EnsemblersContextContextProvider = ({
  projectId,
  children,
  ensemblerType = ""
}) => {
  const [ensemblers, setEnsemblers] = useState({});
  const [
    {
      data: { results },
      isLoaded,
      error
    }
  ] = useTuringApi(
    `/projects/${projectId}/ensemblers`,
    {
      query: !!ensemblerType
        ? {
            type: ensemblerType,
            pageSize: Number.MAX_SAFE_INTEGER
          }
        : { pageSize: Number.MAX_SAFE_INTEGER }
    },
    { results: [] }
  );

  useEffect(() => {
    if (isLoaded && !error) {
      setEnsemblers(
        results.reduce((ensemblers, e) => {
          ensemblers[e.id] = e;
          return ensemblers;
        }, {})
      );
    }
  }, [results, isLoaded, error]);

  return (
    <EnsemblersContext.Provider value={ensemblers}>
      {children}
    </EnsemblersContext.Provider>
  );
};

export default EnsemblersContext;
