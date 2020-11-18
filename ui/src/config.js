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
      : process.env.REACT_APP_MLP_API
};

export const authConfig = {
  oauthClientId: process.env.REACT_APP_OAUTH_CLIENT_ID
};

export const appConfig = {
  environment: process.env.REACT_APP_ENVIRONMENT || "dev",
  homepage: process.env.REACT_APP_HOMEPAGE || process.env.PUBLIC_URL,
  appIcon: "graphApp",
  docsUrl: process.env.REACT_APP_USER_DOCS_URL,
  privateDockerRegistries: process.env.REACT_APP_PRIVATE_DOCKER_REGISTRIES
    ? process.env.REACT_APP_PRIVATE_DOCKER_REGISTRIES.split(",")
    : [],
  scaling: {
    maxAllowedReplica: 10
  }
};

export const sentryConfig = {
  dsn: process.env.REACT_APP_SENTRY_DSN,
  environment: appConfig.environment
};

export const monitoringConfig = {
  dashboardUrl: process.env.REACT_APP_ROUTER_MONITORING_URL || ""
};
