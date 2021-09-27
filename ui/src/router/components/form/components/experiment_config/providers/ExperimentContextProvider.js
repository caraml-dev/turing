import React, { useCallback, useEffect, useState } from "react";
import { useTuringApi } from "../../../../../../hooks/useTuringApi";
import ExperimentContext from "./context";

export const ExperimentContextProvider = ({
  engineProps,
  clientId,
  experimentIds,
  children,
}) => {
  const engine = engineProps.name.toLowerCase();

  const [clientsResponse, fetchClients] = useTuringApi(
    `/experiment-engines/${engine}/clients`,
    {},
    [],
    false
  );

  const [experimentsResponse, fetchExperiments] = useTuringApi(
    `/experiment-engines/${engine}/experiments?client_id=${clientId}`,
    {},
    [],
    false
  );

  const [variablesResponse, fetchVariables] = useTuringApi(
    `/experiment-engines/${engine}/variables?client_id=${clientId}&experiment_id=${experimentIds}`,
    {},
    { config: [] },
    false
  );

  // The following state variables are used to track when the data concerning
  // different components require an update. For example, when the client id
  // changes, experimentsValidated is set to false. This will re-get the experiments
  // list and re-validate the current experiment selection against it (which may or
  // may not cause a change to the selected experimentIds). After this,
  // experimentsValidated is set to true. The dependants of the experiments data
  // (i.e., variables) will only update when experimentsValidated changes back to true.
  const [clientsValidated, setClientsValidated] = useState(true);
  const [experimentsValidated, setExperimentsValidated] = useState(true);
  // variablesValidated is currently redundant since no other components depend on it.
  const [variablesValidated, setVariablesValidated] = useState(true);

  useEffect(() => {
    // Clients, experiments and variables need an update if engine info changes
    engineProps.client_selection_enabled && setClientsValidated(false);
    engineProps.experiment_selection_enabled && setExperimentsValidated(false);
    setVariablesValidated(false);
  }, [engineProps]);

  useEffect(() => {
    // Experiments and variables need an update if client info changes
    engineProps.experiment_selection_enabled && setExperimentsValidated(false);
    setVariablesValidated(false);
  }, [engineProps.experiment_selection_enabled, clientId]);

  useEffect(() => {
    // Variables need an update if experiment info changes
    setVariablesValidated(false);
  }, [experimentIds]);

  useEffect(() => {
    // Fetch clients if engine info changed
    !clientsValidated && fetchClients();
  }, [clientsValidated, fetchClients]);

  useEffect(() => {
    // Fetch experiments if engine / client info changed
    clientsValidated && !experimentsValidated && fetchExperiments();
  }, [clientsValidated, experimentsValidated, fetchExperiments]);

  useEffect(() => {
    // Fetch variables if engine / client / experiment info changed
    clientsValidated &&
      experimentsValidated &&
      !variablesValidated &&
      fetchVariables();
  }, [
    clientsValidated,
    experimentsValidated,
    variablesValidated,
    fetchVariables,
  ]);

  const isLoading = useCallback(
    (val) => {
      switch (val) {
        case "clients":
          return clientsResponse.isLoading;
        case "experiments":
          return experimentsResponse.isLoading;
        case "variables":
          return variablesResponse.isLoading;
        default:
          return false;
      }
    },
    [
      clientsResponse.isLoading,
      experimentsResponse.isLoading,
      variablesResponse.isLoading,
    ]
  );

  const isLoaded = useCallback(
    (val) => {
      switch (val) {
        case "clients":
          return clientsResponse.isLoaded;
        case "experiments":
          return experimentsResponse.isLoaded;
        case "variables":
          return variablesResponse.isLoaded;
        default:
          return false;
      }
    },
    [
      clientsResponse.isLoaded,
      experimentsResponse.isLoaded,
      variablesResponse.isLoaded,
    ]
  );

  return (
    <ExperimentContext.Provider
      value={{
        clients: clientsResponse.data,
        experiments: experimentsResponse.data,
        variables: variablesResponse.data,
        setClientsValidated,
        setExperimentsValidated,
        setVariablesValidated,
        isLoading,
        isLoaded,
      }}
    >
      {children}
    </ExperimentContext.Provider>
  );
};
