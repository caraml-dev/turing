var config = {
  "apiConfig": {
    /*
     * Timeout (in milliseconds) for requests to API
     * "apiTimeout": 10000,
     *
     * Endpoint to Turing API
     * "turingApiUrl": "/api/turing"
     *
     * Endpoint to Merlin API
     * "merlinApiUrl": "/api/merlin"
     *
     * Endpoint to MLP API
     * "mlpApiUrl": "/api"
     */
  },

  "authConfig": {
    /*
     * OAuth2 Client ID
     * "oauthClientId": "CLIENT_ID"
     */
  },

  "appConfig": {
    /*
     * Environment name
     * "environment": "turing-env",
     *
     * "scaling": {
     *   Number of max amount of router/enricher/ensembler replicas configured from UI
     *   "maxAllowedReplica": 20
     * },
     *
     * "pagination": {
     *   Default page size configuration for table views. One of 10, 25, 50
     *   "defaultPageSize": 25
     * }
     */
  },

  "sentryConfig": {
    /*
     * DSN of Sentry project
     * "dsn": "SENTRY_DSN"
     *
     * Sentry environment (if it's different from appConfig.environment)
     * "environment":
     *
     * Additional tags to include with Sentry requests
     * "tags": {
     *   "tag1": "value1",
     *   "tag2": "value2"
     * }
     */
  }
};