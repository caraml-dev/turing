/*
 * In development environment, we set turingApiUrl, merlinApiUrl and mlpApiUrl
 * to unreachable paths so that the requests will be made to the given API
 * servers through proxying (setupProxy.js). This is required to bypass CORS
 * restrictions imposed by the browser. In production, the env vars REACT_APP_TURING_API,
 * REACT_APP_MERLIN_API and REACT_APP_MLP_API can either be absolute URLs, or relative
 * to the UI if the API is served from the same host. When the API's origin differs from
 * that of the UI, appropriate CORS policies are expected to be in place on the API server.
 */
export const apiConfig = {
  apiTimeout: process.env.REACT_APP_API_TIMEOUT || 5000,
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

export const authConfig = {
  oauthClientId: process.env.REACT_APP_OAUTH_CLIENT_ID,
};

export const appConfig = {
  environment: process.env.REACT_APP_ENVIRONMENT || "dev",
  homepage: process.env.REACT_APP_HOMEPAGE || process.env.PUBLIC_URL,
  appIcon: "graphApp",
  docsUrl: process.env.REACT_APP_USER_DOCS_URL
    ? JSON.parse(process.env.REACT_APP_USER_DOCS_URL)
    : [{ href: "https://github.com/gojek/turing", label: "Turing User Guide" }],
  privateDockerRegistries: process.env.REACT_APP_PRIVATE_DOCKER_REGISTRIES
    ? process.env.REACT_APP_PRIVATE_DOCKER_REGISTRIES.split(",")
    : [],
  defaultDockerRegistry:
    process.env.REACT_APP_DEFAULT_DOCKER_REGISTRY || "docker.io", // User Docker Hub as the default
  scaling: {
    maxAllowedReplica: process.env.REACT_APP_MAX_ALLOWED_REPLICA
      ? parseInt(process.env.REACT_APP_MAX_ALLOWED_REPLICA)
      : 10,
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

export const sentryConfig = {
  dsn: process.env.REACT_APP_SENTRY_DSN,
  environment: appConfig.environment,
};

export const resultLoggingConfig = {
  protoUrl: process.env.REACT_APP_RESULT_LOG_PROTO_URL,
};

export const monitoringConfig = {
  dashboardUrl: process.env.REACT_APP_ROUTER_MONITORING_URL || "",
};

export const alertConfig = {
  environment:
    appConfig.environment === "dev" ? "development" : appConfig.environment,
};
