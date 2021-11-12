import React from "react";
import objectAssignDeep from "object-assign-deep";

/*
 * In development environment, we set turingApiUrl, merlinApiUrl and mlpApiUrl
 * to unreachable paths so that the requests will be made to the given API
 * servers through proxying (setupProxy.js). This is required to bypass CORS
 * restrictions imposed by the browser. In production, the env vars REACT_APP_TURING_API,
 * REACT_APP_MERLIN_API and REACT_APP_MLP_API can either be absolute URLs, or relative
 * to the UI if the API is served from the same host. When the API's origin differs from
 * that of the UI, appropriate CORS policies are expected to be in place on the API server.
 */
const apiConfig = {
  apiTimeout: process.env.REACT_APP_API_TIMEOUT,
  useMockData: process.env.REACT_APP_USE_MOCK_DATA || false,
  turingApiUrl:
    process.env.NODE_ENV === "development"
      ? "/api/turing"
      : process.env.REACT_APP_TURING_API,
  merlinApiUrl:
    process.env.NODE_ENV === "development"
      ? "/api/merlin"
      : process.env.REACT_APP_MERLIN_API,
  mlpApiUrl:
    process.env.NODE_ENV === "development"
      ? "/api/mlp"
      : process.env.REACT_APP_MLP_API,
};

const authConfig = {
  oauthClientId: process.env.REACT_APP_OAUTH_CLIENT_ID,
};

const appConfig = {
  environment: process.env.REACT_APP_ENVIRONMENT,
  homepage: process.env.PUBLIC_URL,
  appIcon: "graphApp",
  docsUrl: JSON.parse(process.env.REACT_APP_USER_DOCS_URL),
  privateDockerRegistries: process.env.REACT_APP_PRIVATE_DOCKER_REGISTRIES
    ? process.env.REACT_APP_PRIVATE_DOCKER_REGISTRIES.split(",")
    : [],
  defaultDockerRegistry: process.env.REACT_APP_DEFAULT_DOCKER_REGISTRY,
  scaling: {
    // Max number of router/enricher/ensembler replica allowed to set during the deployment
    maxAllowedReplica: parseInt(process.env.REACT_APP_MAX_ALLOWED_REPLICA),
  },
  pagination: {
    defaultPageSize: 10,
  },
  tables: {
    defaultTextSize: "s",
    defaultIconSize: "s",
    dateFormat: "YYYY-MM-DDTHH:mm.SSZ",
  },
  podLogs: {
    // Interval (in ms) between API calls to Logs API
    pollInterval: 10000,
    // Max number of log entries to be fetched in a single API call
    batchSize: 500,
    // Default number of tail log entries to be fetched
    defaultTailLines: 1000,
  },
};

const sentryConfig = {
  dsn: process.env.REACT_APP_SENTRY_DSN,
  environment: appConfig.environment,
  tags: {
    app: "turing-ui",
  },
};

const resultLoggingConfig = {
  protoUrl: process.env.REACT_APP_RESULT_LOG_PROTO_URL,
};

const alertConfig = {
  enabled: true,
  environment:
    appConfig.environment === "dev" ? "development" : appConfig.environment,
};

const defaultExperimentEngine = process.env.REACT_APP_DEFAULT_EXPERIMENT_ENGINE
  ? JSON.parse(process.env.REACT_APP_DEFAULT_EXPERIMENT_ENGINE)
  : {};

const buildTimeConfig = {
  apiConfig,
  authConfig,
  appConfig,
  sentryConfig,
  resultLoggingConfig,
  alertConfig,
  defaultExperimentEngine,
};

const ConfigContext = React.createContext({});

export const ConfigProvider = ({ children }) => {
  const runTimeConfig = window.config;
  const config = objectAssignDeep({}, buildTimeConfig, runTimeConfig);

  return (
    <ConfigContext.Provider value={config}>{children}</ConfigContext.Provider>
  );
};

export const useConfig = () => React.useContext(ConfigContext);
