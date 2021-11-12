{{- define "merlin.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "merlin.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version -}}
{{- end -}}

{{- define "merlin.fullname" -}}
{{- if .Values.fullnameOverride -}}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}
{{- end -}}

{{- define "merlin.postgresql.name" -}}
{{- printf "%s-%s" .Release.Name (default "postgresql" .Values.postgresql.nameOverride) -}}
{{- end -}}

{{- define "merlin.db.host" -}}
{{- printf "%s.%s.svc.cluster.local" (include "merlin.postgresql.name" .) .Release.Namespace -}}
{{- end -}}

{{- define "merlin.mlflow.backendStoreUri" -}}
{{- printf "postgresql://%s:%s@%s:5432/%s" .Values.postgresql.postgresqlUsername .Values.postgresql.postgresqlPassword (include "merlin.db.host" .) .Values.postgresql.postgresqlDatabase -}}
{{- end -}}

{{- define "merlin.vault.secretName" -}}
{{- if eq (.Values.vault.enabled | toString) "true" -}}
    {{ include "merlin.fullname" . }}-secrets
{{- else -}}
    {{- .Values.vault.secretName -}}
{{- end -}}
{{- end -}}

{{- define "merlin.vault.address" -}}
{{- printf "http://%s-vault:8200" .Release.Name -}}
{{- end -}}

{{- define "merlin.mlp.encryption.key" -}}
{{- if and .Values.global.mlp .Values.global.mlp.encryption -}}
    {{- .Values.global.mlp.encryption.key | default .Values.mlpApi.encryptionKey -}}
{{- else -}}
    {{ .Values.mlpApi.encryptionKey }}
{{- end -}}
{{- end -}}

{{/*
Renders a value that contains template.
Usage:
{{ include "merlin.tplValue" ( dict "value" .Values.path.to.the.Value "context" $) }}
*/}}
{{- define "merlin.tplValue" -}}
    {{- if typeIs "string" .value }}
        {{- tpl .value .context }}
    {{- else }}
        {{- tpl (.value | toYaml) .context }}
    {{- end }}
{{- end -}}


{{- define "merlin.mlp.apiHost" -}}
{{- end -}}

{{- define "merlin.oauthClientId" -}}
{{- .Values.global.oauthClientId | default .Values.oauthClientId -}}
{{- end -}}

{{- define "merlin.environment" -}}
{{- .Values.global.environment | default .Values.environment -}}
{{- end -}}

{{- define "merlin.alerts.enabled" -}}
{{ eq (.Values.alert.enabled | toString) "true" }}
{{- end -}}

{{- define "merlin.alerts.envVars" -}}
- name: ALERT_ENABLED
  value: {{ include "merlin.alerts.enabled" . | quote }}
{{- if (include "merlin.alerts.enabled" .) }}
- name: GITLAB_BASE_URL
  value: "{{ .Values.alert.gitlab.baseURL }}"
- name: GITLAB_TOKEN
  valueFrom:
    secretKeyRef:
      name: {{ include "merlin.fullname" . }}-secrets
      key: gitlab-token
- name: GITLAB_DASHBOARD_REPOSITORY
  value: "{{ .Values.alert.gitlab.dashboardRepository }}"
- name: GITLAB_DASHBOARD_BRANCH
  value: "{{ .Values.alert.gitlab.dashboardBranch }}"
- name: GITLAB_ALERT_REPOSITORY
  value: "{{ .Values.alert.gitlab.alertRepository }}"
- name: GITLAB_ALERT_BRANCH
  value: "{{ .Values.alert.gitlab.alertBranch }}"
- name: WARDEN_API_HOST
  value: "{{ .Values.alert.warden.apiHost }}"
{{- end -}}
{{- end -}}

{{- define "merlin.imageBuilder.envVars" -}}
- name: IMG_BUILDER_CLUSTER_NAME
  value: "{{ .Values.imageBuilder.clusterName }}"
- name: IMG_BUILDER_BUILD_CONTEXT_URI
  value: "{{ .Values.imageBuilder.buildContextURI }}"
- name: IMG_BUILDER_DOCKERFILE_PATH
  value: "{{ .Values.imageBuilder.dockerfilePath }}"
- name: IMG_BUILDER_BASE_IMAGE
  value: "{{ .Values.imageBuilder.baseImage }}"
- name: IMG_BUILDER_PREDICTION_JOB_BUILD_CONTEXT_URI
  value: "{{ .Values.imageBuilder.predictionJobBuildContextURI }}"
- name: IMG_BUILDER_PREDICTION_JOB_DOCKERFILE_PATH
  value: "{{ .Values.imageBuilder.predictionJobDockerfilePath }}"
- name: IMG_BUILDER_PREDICTION_JOB_BASE_IMAGE
  value: "{{ .Values.imageBuilder.predictionJobBaseImage }}"
- name: IMG_BUILDER_NAMESPACE
  value: "{{ .Values.imageBuilder.namespace }}"
- name: IMG_BUILDER_DOCKER_REGISTRY
  value: "{{ .Values.imageBuilder.dockerRegistry }}"
- name: IMG_BUILDER_TIMEOUT
  value: "{{ .Values.imageBuilder.timeout }}"
{{- if .Values.imageBuilder.contextSubPath }}
- name: IMG_BUILDER_CONTEXT_SUB_PATH
  value: "{{ .Values.imageBuilder.contextSubPath }}"
{{- end }}
{{- if .Values.imageBuilder.predictionJobContextSubPath }}
- name: IMG_BUILDER_PREDICTION_JOB_CONTEXT_SUB_PATH
  value: "{{ .Values.imageBuilder.predictionJobContextSubPath }}"
{{- end }}
{{- end -}}

{{- define "merlin.monitoring.enabled" -}}
{{ eq (.Values.monitoring.enabled | toString) "true" }}
{{- end -}}

{{- define "merlin.monitoring.envVars" -}}
 - name: MONITORING_DASHBOARD_ENABLED
  value: {{ include "merlin.monitoring.enabled" . | quote }}
{{- if (include "merlin.monitoring.enabled" .) }}
- name: MONITORING_DASHBOARD_BASE_URL
  value: "{{ .Values.monitoring.baseURL }}"
- name: MONITORING_DASHBOARD_JOB_BASE_URL
  value: "{{ .Values.monitoring.jobBaseURL }}"
{{- end }}
{{- end -}}

{{- define "merlin.sentry.enabled" -}}
{{ eq (.Values.sentry.enabled | toString) "true" }}
{{- end -}}

{{- define "merlin.ui.envVars" -}}
- name: REACT_APP_OAUTH_CLIENT_ID
  value: {{ include "merlin.oauthClientId" . | quote }}
- name: REACT_APP_ENVIRONMENT
  value: {{ include "merlin.environment" . | quote }}
- name: REACT_APP_ALERT_ENABLED
  value: {{ include "merlin.alerts.enabled" . | quote }}
- name: REACT_APP_MONITORING_DASHBOARD_ENABLED
  value: {{ include "merlin.monitoring.enabled" . | quote }}
- name: REACT_APP_MERLIN_DOCS_URL
  value: "{{ .Values.docsURL }}"
- name: REACT_APP_HOMEPAGE
  value: "{{ .Values.homepage }}"
- name: REACT_APP_MERLIN_API
  value: "{{ .Values.apiHost }}"
- name: REACT_APP_MLP_API
  value: "{{ .Values.mlpApi.apiHost }}"
- name: REACT_APP_DOCKER_REGISTRIES
  value: "{{ .Values.dockerRegistries }}"
{{- if (include "merlin.sentry.enabled" .) }}
- name: REACT_APP_SENTRY_DSN
  value: "{{ .Values.sentry.dsn }}"
{{- end -}}
{{- end -}}