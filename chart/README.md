# turing

![Version: 0.1.0](https://img.shields.io/badge/Version-0.1.0-informational?style=flat-square)

Turing: ML Experimentation System

## Requirements

| Repository | Name | Version |
|------------|------|---------|
| https://charts.bitnami.com/bitnami | postgresql | 8.9.8 |

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| dbMigrations.image.tag | string | `"v4.7.1"` | Docker tag for golang-migrate Docker image https://hub.docker.com/r/migrate/migrate |
| postgresql.persistence.enabled | bool | `true` | Persist Postgresql data in a Persistent Volume Claim  |
| postgresql.postgresqlDatabase | string | `"turing"` | Database name for Turing Postgresql database |
| postgresql.postgresqlPassword | string | `nil` | Password for Turing Postgresql database (required) |
| postgresql.postgresqlUsername | string | `"turing"` | Username for Turing Postgresql database |
| postgresql.resources | object | `{}` | Resources requests and limits for Turing database. This should be set  according to your cluster capacity and service level objectives. Reference: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| swaggerUi.apiHost | string | `"127.0.0.1"` | Host for the Swagger UI |
| swaggerUi.basePath | string | `"/v1"` | Base URL path to serve the Swagger UI |
| swaggerUi.image | object | `{"tag":"v3.24.3"}` | Docker tag for Swagger UI https://hub.docker.com/r/swaggerapi/swagger-ui |
| swaggerUi.service.externalPort | int | `8080` | Swagger UI Kubernetes service port number |
| swaggerUi.service.internalPort | int | `8081` | Swagger UI container port number |
| turing.config.alert.enabled | bool | `false` | Enable alerting in Turing |
| turing.config.alert.gitlabBaseurl | string | `nil` | GitLab server URL for GitOps based alerting |
| turing.config.alert.gitlabBranch | string | `nil` | GitLab branch to commit the alert configuration file |
| turing.config.alert.gitlabPathprefix | string | `nil` | GitLab path prefix within the project to save the alert configuration file |
| turing.config.alert.gitlabProjectid | string | `nil` | GitLab project ID |
| turing.config.alert.gitlabToken | string | `nil` | GitLab token to authentication API request to GitLab |
| turing.config.authorization.enabled | bool | `false` | Enable authorization middleware in Turing API |
| turing.config.authorization.serverURL | string | `nil` | Authorization server URL |
| turing.config.deployment.deletionTimeout | string | `"30s"` | Maximum wait duration to delete Turing router |
| turing.config.deployment.environmentType | string | `"id-dev"` | Environment name associated with Turing router |
| turing.config.deployment.gcpProject | string | `""` | Google Cloud Project ID associcated with Turing router |
| turing.config.deployment.maxCPU | string | `"8"` | Hard limit on the maximum CPU Turing router can request |
| turing.config.deployment.maxMemory | string | `"8Gi"` | Hard limit on the maximum memory Turing router can request |
| turing.config.deployment.timeout | string | `"3m"` | Maximum wait duration to create or update Turing router |
| turing.config.encryption.key | string | `nil` | Encryption key used by Turing to secure sensitive values (required) |
| turing.config.ingress.enabled | bool | `false` | Enable ingress to provision Ingress resource for external access to Turing API  |
| turing.config.merlin.endpoint | string | `nil` | Merlin API endpoint (required). Reference: https://github.com/gojek/merlin |
| turing.config.mlp.encryption.key | string | `nil` | Encryption key used by MLP to secure sensitive values (required) |
| turing.config.mlp.endpoint | string | `nil` | MLP API endpoint(required). Reference: https://github.com/gojek/mlp |
| turing.config.newrelic.appname | string | `nil` | Application name monitored in New Relic |
| turing.config.newrelic.enabled | bool | `false` | Enable integrarion with New Relic application monitoring https://newrelic.com |
| turing.config.router.customMetrics | bool | `true` | Enable custom metrics |
| turing.config.router.fiberDebugLog | bool | `true` | Enable debugging for Fiber library used by Turing router |
| turing.config.router.fluentd.flushIntervalSeconds | int | `90` | How often should Fluentd flush the buffered log |
| turing.config.router.fluentd.image | string | `nil` | Docker image for Fluentd log forwarder. User is expected to specify the Fluentd image for now as there is no publicly available image.  Currently this is required only if users needs to save Turing logs in BigQuery. |
| turing.config.router.image.registry | string | `"docker.io/"` | Docker registry for Turing router image. User is expected to override the registry for now as there is no publicly available Turing image |
| turing.config.router.image.repository | string | `"turing-router"` | Docker image repository for Turing router |
| turing.config.router.image.tag | string | `"latest"` | Docker image tag for Turing router |
| turing.config.router.jaeger.collectorEndpoint | string | `nil` | Jaeger tracing collector endpoint  |
| turing.config.router.jaeger.enabled | bool | `false` | Enable Jaeger tracing |
| turing.config.router.logLevel | string | `"DEBUG"` | Log level for Turing router |
| turing.config.sentry.dsn | string | `nil` | Data source name for the Sentry project  |
| turing.config.sentry.enabled | bool | `false` | Enable integration with Sentry application monitoring https://sentry.io |
| turing.config.serviceAccount.enabled | bool | `false` | Enable usage of Google Cloud service account JSON key |
| turing.config.serviceAccount.secretKey | string | `nil` | Secret key for Google Cloud service account JSON key |
| turing.config.serviceAccount.secretName | string | `nil` | Secret name for Google Cloud service account JSON key |
| turing.config.vault.address | string | `nil` | Vault server address (required) |
| turing.config.vault.token | string | `nil` | Vault authentication token (required) |
| turing.image.registry | string | `"docker.io/"` | Docker registry for Turing API image. User is required to override the registry for now as there is no publicly available Turing image |
| turing.image.repository | string | `"turing"` | Docker image repository for Turing API |
| turing.image.tag | string | `"latest"` | Docker image tag for Turing API |
| turing.livenessProbe.path | string | `"/v1/internal/live"` | HTTP path for liveness check |
| turing.readinessProbe.path | string | `"/v1/internal/ready"` | HTTP path for readiness check |
| turing.resources | object | `{}` | Resources requests and limits for Turing API. This should be set  according to your cluster capacity and service level objectives. Reference: https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/ |
| turing.service.externalPort | int | `8080` | Turing API Kubernetes service port number |
| turing.service.internalPort | int | `8080` | Turing API container port number |

