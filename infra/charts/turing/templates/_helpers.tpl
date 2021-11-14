{{- define "turing.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "turing.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version -}}
{{- end -}}

{{- define "turing.fullname" -}}
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

{{- define "turing.image" -}}
{{- $registryName := .Values.turing.image.registry -}}
{{- $repositoryName := .Values.turing.image.repository -}}
{{- $tag := .Values.turing.image.tag | toString -}}
{{- if $registryName }}
    {{- printf "%s/%s:%s" $registryName $repositoryName $tag -}}
{{- else -}}
    {{- printf "%s:%s" $repositoryName $tag -}}
{{- end -}}
{{- end -}}

{{- define "turing.environment" -}}
{{- .Values.global.environment | default .Values.turing.environment | default "dev" -}}
{{- end -}}

{{- define "turing.postgresql.name" -}}
{{- printf "%s-%s" .Release.Name (.Values.postgresql.nameOverride | default "postgresql" ) -}}
{{- end -}}

{{- define "turing.db.host" -}}
{{- printf "%s.%s.svc.cluster.local" (include "turing.postgresql.name" .) .Release.Namespace -}}
{{- end -}}

{{- define "turing.encryption.key" -}}
{{- $encryptionKey := .Values.turing.config.TuringEncryptionKey | default (randAlpha 12) -}}
{{- if and .Values.global.mlp .Values.global.mlp.encryption -}}
    {{- .Values.global.mlp.encryption.key | default $encryptionKey -}}
{{- else -}}
    {{ $encryptionKey }}
{{- end -}}
{{- end -}}

{{- define "turing.serviceAccount.name" -}}
{{- include "turing.fullname" . -}}
{{- end -}}

{{- define "turing.merlin.name" -}}
{{- printf "%s-%s" .Release.Name (.Values.merlin.nameOverride | default "merlin") -}}
{{- end -}}

{{- define "turing.mlp.name" -}}
{{- printf "%s-%s" .Release.Name (.Values.mlp.nameOverride | default "mlp") -}}
{{- end -}}

{{- define "turing.clusterConfig.useInClusterCredentials" -}}
{{- .Values.tags.mlp | ternary true .Values.turing.clusterConfig.useInClusterConfig -}}
{{- end -}}

{{- define "turing.mlp.encryption.key" -}}
{{- $encryptionKey := .Values.turing.config.MLPConfig.MLPEncryptionKey -}}
{{- if and .Values.global.mlp .Values.global.mlp.encryption -}}
    {{- .Values.global.mlp.encryption.key | default $encryptionKey -}}
{{- else -}}
    {{ $encryptionKey }}
{{- end -}}
{{- end -}}

{{- define "turing.sentry.dsn" -}}
{{- .Values.global.sentry.dsn | default .Values.sentry.dsn -}}
{{- end -}}

{{- define "turing.defaultConfig" -}}
ClusterConfig:
  InClusterConfig: {{ .Values.turing.clusterConfig.useInClusterConfig }}
DbConfig:
  Host: {{ include "turing.db.host" . | quote }}
  Database: {{ .Values.postgresql.postgresqlDatabase }}
  User: {{ .Values.postgresql.postgresqlUsername }}
  Password: {{ .Values.postgresql.postgresqlPassword }}
DeployConfig:
  EnvironmentType: {{ .Values.turing.config.DeployConfig.EnvironmentType | default (include "turing.environment" .) }}
KubernetesLabelConfigs:
  Environment: {{ .Values.turing.config.KubernetesLabelConfigs.Environment | default (include "turing.environment" .) }}
MLPConfig:
  MLPEncryptionKey: {{ include "turing.mlp.encryption.key" . | quote }}
{{ if .Values.tags.mlp }}
  MerlinURL: {{ printf "http://%s:8080/v1" (include "turing.merlin.name" .) }}
  MLPURL: {{ printf "http://%s:8080/v1" (include "turing.mlp.name" .) }}
{{ end }}
TuringEncryptionKey: {{ include "turing.encryption.key" . | quote }}
Sentry:
  DSN: {{ .Values.turing.config.Sentry.DSN | default (include "turing.sentry.dsn" .) | quote }}
{{- end -}}

{{- define "turing.config" -}}
{{- $defaultConfig := include "turing.defaultConfig" . | fromYaml -}}
{{ .Values.turing.config | merge $defaultConfig | toYaml }}
{{- end -}}

{{- define "turing.ui.defaultConfig" -}}
{{- if .Values.turing.uiConfig -}}
alertConfig:
  enabled: {{ eq (quote .Values.turing.uiConfig.alertConfig.enabled) "" | ternary .Values.turing.config.AlertConfig.Enabled .Values.turing.uiConfig.alertConfig.enabled }}
appConfig:
  environment: {{ .Values.turing.uiConfig.appConfig.environment | default (include "turing.environment" .) }}
authConfig:
  oauthClientId: {{ .Values.global.oauthClientId | default .Values.turing.uiConfig.authConfig.oauthClientId | quote }}
sentryConfig:
  environment: {{ .Values.turing.uiConfig.sentryConfig.environment | default (include "turing.environment" .) }}
  dsn: {{ .Values.turing.uiConfig.sentryConfig.dsn | default (include "turing.sentry.dsn" .) | quote }}
{{- end -}}
{{- end -}}

{{- define "turing.ui.config" -}}
{{- $defaultConfig := include "turing.ui.defaultConfig" . | fromYaml -}}
{{ .Values.turing.uiConfig | merge $defaultConfig | toPrettyJson }}
{{- end -}}