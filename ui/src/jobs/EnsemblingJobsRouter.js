import React from "react";
import { Redirect, Router } from "@reach/router";
import { ListEnsemblingJobsLandingView } from "./list/ListEnsemblingJobsLandingView";
import { EnsemblingJobDetailsView } from "./details/EnsemblingJobDetailsView";
import { EnsemblersContextContextProvider } from "../providers/ensemblers/context";

export const EnsemblingJobsRouter = ({ projectId }) => (
  <EnsemblersContextContextProvider projectId={projectId}>
    <Router>
      <ListEnsemblingJobsLandingView path="/" />

      <EnsemblingJobDetailsView path=":jobId/*" />

      <Redirect from="any" to="/error/404" default noThrow />
    </Router>
  </EnsemblersContextContextProvider>
);
