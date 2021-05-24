import React from "react";
import { navigate } from "@reach/router";
import {
  ApplicationsContextProvider,
  CurrentProjectContextProvider,
  Header,
  ProjectsContextProvider
} from "@gojek/mlp-ui";
import { appConfig } from "./config";
import { EnvironmentsContextProvider } from "./providers/environments/context";
import "./PrivateLayout.scss";

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
            docLinks={appConfig.docsUrl}
          />
          <div className="main-component-layout">
            <EnvironmentsContextProvider>
              <Component {...props} />
            </EnvironmentsContextProvider>
          </div>
        </CurrentProjectContextProvider>
      </ProjectsContextProvider>
    </ApplicationsContextProvider>
  );
};
