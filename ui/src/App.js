import React from "react";
import { Route, Routes } from "react-router-dom";
import {
  AuthProvider,
  Page404,
  ErrorBoundary,
  Login,
  MlpApiContextProvider,
  Toast,
} from "@caraml-dev/ui-lib";
import { PrivateLayout } from "./PrivateLayout";
import { useConfig } from "./config";
import { EuiProvider } from "@elastic/eui";
import AppRoutes from "./AppRoutes";

const App = () => {
  const { apiConfig, authConfig } = useConfig();
  return (
    <EuiProvider>
      <ErrorBoundary>
        <MlpApiContextProvider
          mlpApiUrl={apiConfig.mlpApiUrl}
          timeout={apiConfig.apiTimeout}
          useMockData={apiConfig.useMockData}>
          <AuthProvider clientId={authConfig.oauthClientId}>
            <Routes>
              <Route path="/login" element={<Login />} />
              <Route element={<PrivateLayout />}>
                <Route path="/*" element={<AppRoutes />} />
              </Route>
              <Route path="/pages/404" element={<Page404 />} />
            </Routes>
            <Toast />
          </AuthProvider>
        </MlpApiContextProvider>
      </ErrorBoundary>
    </EuiProvider>
  );
};

export default App;
