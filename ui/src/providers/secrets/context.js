import React from "react";
import { useMerlinApi } from "../../hooks/useMerlinApi";

const SecretsContext = React.createContext([]);

export const SecretsContextProvider = ({ projectId, children }) => {
  const [{ data: secrets }] = useMerlinApi(
    `/projects/${projectId}/secrets`,
    {},
    []
  );

  return (
    <SecretsContext.Provider value={secrets}>
      {children}
    </SecretsContext.Provider>
  );
};

export default SecretsContext;
