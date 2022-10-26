import React from "react";
import {
  ApplicationsContext,
  ApplicationsContextProvider,
  Header,
  PrivateRoute,
  ProjectsContextProvider,
} from "@gojek/mlp-ui";
import urlJoin from "proper-url-join";
import { Outlet, useNavigate } from "react-router-dom";
import { useConfig } from "./config";
import { EnvironmentsContextProvider } from "./providers/environments/context";
import "./PrivateLayout.scss";

export const PrivateLayout = () => {
  const navigate = useNavigate();
  const { appConfig } = useConfig();
  return (
    <PrivateRoute>
      <ApplicationsContextProvider>
        <ProjectsContextProvider>
          <EnvironmentsContextProvider>
            <ApplicationsContext.Consumer>
              {({ currentApp }) => (
                <Header
                  onProjectSelect={pId =>
                    navigate(urlJoin(currentApp?.href, "projects", pId, "routers"))
                  }
                  docLinks={appConfig.docsUrl}
                />
              )}
            </ApplicationsContext.Consumer>
            <Outlet />
          </EnvironmentsContextProvider>
        </ProjectsContextProvider>
      </ApplicationsContextProvider>
    </PrivateRoute>
  );
};
