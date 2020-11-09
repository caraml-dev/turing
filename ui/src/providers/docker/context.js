import React from "react";
import { appConfig } from "../../config";

const DockerRegistriesContext = React.createContext([]);

export const DockerRegistriesContextProvider = ({ children }) => {
  const registries = [
    {
      value: `asia.gcr.io/gods-${appConfig.environment}`,
      inputDisplay: `asia.gcr.io/gods-${appConfig.environment}`
    },
    {
      value: "",
      inputDisplay: "Docker Hub"
    }
  ];

  return (
    <DockerRegistriesContext.Provider value={registries}>
      {children}
    </DockerRegistriesContext.Provider>
  );
};

export default DockerRegistriesContext;
