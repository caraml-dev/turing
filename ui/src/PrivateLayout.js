import {
  ApplicationsContextProvider,
  CurrentProjectContextProvider,
  Header,
  NavDrawer,
  ProjectsContextProvider
} from "@gojek/mlp-ui";
import { appConfig } from "./config";
import { navigate } from "@reach/router";
import { EnvironmentsContextProvider } from "./providers/environments/context";
import React from "react";

export const PrivateLayout = Component => {
  return props => (
    <ApplicationsContextProvider>
      <ProjectsContextProvider>
        <CurrentProjectContextProvider {...props}>
          <Header
            homeUrl={appConfig.homepage}
            appIcon={appConfig.appIcon}
            onProjectSelect={projectId =>
              navigate(`${appConfig.homepage}/projects/${projectId}/routers`)
            }
            helpLink={appConfig.docsUrl}
          />
          <NavDrawer homeUrl={appConfig.homepage} />
          <EnvironmentsContextProvider>
            <Component {...props} />
          </EnvironmentsContextProvider>
        </CurrentProjectContextProvider>
      </ProjectsContextProvider>
    </ApplicationsContextProvider>
  );
};
