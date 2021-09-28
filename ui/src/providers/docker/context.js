import React from "react";
import { useConfig } from "../../config";

const DockerRegistriesContext = React.createContext([]);

export const DockerRegistriesContextProvider = ({ children }) => {
  const { appConfig } = useConfig();

  const registries = [
    ...appConfig.privateDockerRegistries.map((registry) => ({
      value: registry,
      inputDisplay: registry,
    })),
    {
      value: "docker.io",
      inputDisplay: "Docker Hub",
    },
  ];

  return (
    <DockerRegistriesContext.Provider value={registries}>
      {children}
    </DockerRegistriesContext.Provider>
  );
};

export default DockerRegistriesContext;
