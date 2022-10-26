import React from "react";
import { Route, Routes, useParams } from "react-router-dom";
import { ListEnsemblingJobsLandingView } from "./list/ListEnsemblingJobsLandingView";
import { EnsemblingJobDetailsView } from "./details/EnsemblingJobDetailsView";
import { EnsemblersContextProvider } from "../providers/ensemblers/context";

export const EnsemblingJobsRouter = () => {
  const { projectId } = useParams();
  return (
    <EnsemblersContextProvider projectId={projectId}>
      <Routes>
        <Route index element={<ListEnsemblingJobsLandingView />} />
        <Route path=":jobId/*" element={<EnsemblingJobDetailsView />} />
      </Routes>
    </EnsemblersContextProvider>
  )
};
