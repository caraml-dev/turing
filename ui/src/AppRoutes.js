import React from "react";
import { Navigate, Route, Routes } from "react-router-dom";
import Home from "./Home";
import { CreateRouterView } from "./router/create/CreateRouterView";
import { ListRoutersView } from "./router/list/ListRoutersView";
import { RouterDetailsView } from "./router/details/RouterDetailsView";
import { RouterVersionDetailsView } from "./router/versions/details/RouterVersionDetailsView";
import { ListEnsemblersView } from "./ensembler/list/ListEnsemblersView";
import { ExperimentsRouter } from "./experiment/ExperimentsRouter";
import { EnsemblingJobsRouter } from "./jobs/EnsemblingJobsRouter";
import { useConfig } from "./config";


const AppRoutes = () => {
  const { appConfig } = useConfig();

  return (
    <Routes>
      <Route path={appConfig.homepage}>
        <Route index element={<Home />} />
        <Route path="projects/:projectId">
          <Route index element={<Navigate to="routers" replace={true} />} />
          {/* BATCH JOBS */}
          <Route path="jobs/*" element={<EnsemblingJobsRouter />} />
          {/* ENSEMBLERS */}
          <Route path="ensemblers" element={<ListEnsemblersView />} />
          {/* EXPERIMENTS */}
          <Route path="experiments" element={<ExperimentsRouter />} />
          {/* ROUTERS */}
          <Route path="routers">
            {/* LIST */}
            <Route index element={<ListRoutersView />} />
            {/* CREATE */}
            <Route path="create" element={<CreateRouterView />} />
            {/* ROUTER ID */}
            <Route path=":routerId/*" element={<RouterDetailsView />} />
            {/* ROUTER VERSION */}
            <Route path=":routerId/versions/:versionId/*" element={<RouterVersionDetailsView />} />
          </Route>
        </Route>
      </Route>
      {/* DEFAULT */}
      <Route path="*" element={<Navigate to="/pages/404" replace={true} />} />
    </Routes>
  );
};

export default AppRoutes;
