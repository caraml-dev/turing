import React, { useMemo } from "react";
import { useTuringApi } from "../../hooks/useTuringApi";
import ExperimentEngineContext from "./context";
import { getExperimentEngineOptions } from "../../router/components/form/components/experiment_config/typeOptions";

export const ExperimentEngineContextProvider = ({ children }) => {
  const [{ data: experimentEngines }] = useTuringApi(
    `/experiment-engines`,
    {},
    []
  );

  const getEngineProperties = engineType => {
    return (
      experimentEngines.find(eng => eng.name.toLowerCase() === engineType) || {}
    );
  };

  const experimentEngineOptions = useMemo(() => {
    return getExperimentEngineOptions(experimentEngines).map(o => o.value);
  }, [experimentEngines]);

  return (
    <ExperimentEngineContext.Provider
      value={{
        experimentEngines,
        experimentEngineOptions,
        getEngineProperties
      }}>
      {children}
    </ExperimentEngineContext.Provider>
  );
};
