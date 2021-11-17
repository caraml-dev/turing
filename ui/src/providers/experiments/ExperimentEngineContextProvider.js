import React, { useMemo } from "react";
import { transformAll } from "@overgear/yup-ast";

import { useTuringApi } from "../../hooks/useTuringApi";
import ExperimentEngineContext from "./context";
import { getExperimentEngineOptions } from "../../router/components/form/components/experiment_config/typeOptions";

export const ExperimentEngineContextProvider = ({ children }) => {
  const [{ data, isLoaded }] = useTuringApi(`/experiment-engines`, {}, []);

  const experimentEngines = useMemo(() => {
    // For all custom experiment managers, parse validation schema
    const parseCustomValidationSchema = (engineProps) => {
      try {
        if (engineProps.type === "custom") {
          const rawSchema =
            engineProps.custom_experiment_manager_config
              .experiment_config_schema;
          return {
            ...engineProps,
            custom_experiment_manager_config: {
              ...engineProps.custom_experiment_manager_config,
              parsed_experiment_config_schema: transformAll(
                JSON.parse(rawSchema)
              ),
            },
          };
        }
      } catch (_) {} // Ignore errors
      return engineProps;
    };
    return data.map((e) => parseCustomValidationSchema(e));
  }, [data]);

  const experimentEngineOptions = useMemo(() => {
    return getExperimentEngineOptions(experimentEngines).map((o) => o.value);
  }, [experimentEngines]);

  const getEngineProperties = (engineType) =>
    experimentEngines.find((eng) => eng.name.toLowerCase() === engineType) ||
    {};

  return (
    <ExperimentEngineContext.Provider
      value={{
        experimentEngines,
        experimentEngineOptions,
        isLoaded,
        getEngineProperties,
      }}>
      {children}
    </ExperimentEngineContext.Provider>
  );
};
