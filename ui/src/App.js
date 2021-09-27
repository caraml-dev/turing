import React from "react";
import { Redirect, Router } from "@reach/router";
import {
  AuthProvider,
  Empty,
  ErrorBoundary,
  Login,
  MlpApiContextProvider,
  PrivateRoute,
  Toast,
} from "@gojek/mlp-ui";
import Home from "./Home";
import { CreateRouterView } from "./router/create/CreateRouterView";
import { apiConfig, appConfig, authConfig } from "./config";
import { ListRoutersView } from "./router/list/ListRoutersView";
import { RouterDetailsView } from "./router/details/RouterDetailsView";
import { RouterVersionDetailsView } from "./router/versions/details/RouterVersionDetailsView";
import { PrivateLayout } from "./PrivateLayout";
import { ListEnsemblersView } from "./ensembler/list/ListEnsemblersView";
import { ExperimentsRouter } from "./experiment/ExperimentsRouter";
import { EnsemblingJobsRouter } from "./jobs/EnsemblingJobsRouter";

const App = () => (
  <ErrorBoundary>
    <MlpApiContextProvider
      mlpApiUrl={apiConfig.mlpApiUrl}
      timeout={apiConfig.apiTimeout}
      useMockData={apiConfig.useMockData}
    >
      <AuthProvider clientId={authConfig.oauthClientId}>
        <Router role="group">
          <Login path="/login" />

          <Redirect from="/" to={appConfig.homepage} noThrow />

          <Redirect
            from={`${appConfig.homepage}/projects/:projectId`}
            to={`${appConfig.homepage}/projects/:projectId/routers`}
            noThrow
          />

          {/* HOME */}
          <PrivateRoute
            path={appConfig.homepage}
            render={PrivateLayout(Home)}
          />

          {/* BATCH JOBS */}
          <PrivateRoute
            path={`${appConfig.homepage}/projects/:projectId/jobs/*`}
            render={PrivateLayout(EnsemblingJobsRouter)}
          />

          {/* ENSEMBLERS */}
          <PrivateRoute
            path={`${appConfig.homepage}/projects/:projectId/ensemblers`}
            render={PrivateLayout(ListEnsemblersView)}
          />

          {/* CREATE ROUTER */}
          <PrivateRoute
            path={`${appConfig.homepage}/projects/:projectId/routers/create`}
            render={PrivateLayout(CreateRouterView)}
          />

          {/* LIST ROUTER */}
          <PrivateRoute
            path={`${appConfig.homepage}/projects/:projectId/routers`}
            render={PrivateLayout(ListRoutersView)}
          />

          {/* EXPERIMENTS */}
          <PrivateRoute
            path={`${appConfig.homepage}/projects/:projectId/experiments/*`}
            render={PrivateLayout(ExperimentsRouter)}
          />

          {/* ROUTER DETAILS */}
          <PrivateRoute
            path={`${appConfig.homepage}/projects/:projectId/routers/:routerId/*`}
            render={PrivateLayout(RouterDetailsView)}
          />

          {/* ROUTER VERSION DETAILS */}
          <PrivateRoute
            path={`${appConfig.homepage}/projects/:projectId/routers/:routerId/versions/:versionId/*`}
            render={PrivateLayout(RouterVersionDetailsView)}
          />

          {/* DEFAULT */}
          <Empty default />
        </Router>
        <Toast />
      </AuthProvider>
    </MlpApiContextProvider>
  </ErrorBoundary>
);

export default App;
