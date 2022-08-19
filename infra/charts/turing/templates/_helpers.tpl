{{- define "turing.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{- define "turing.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version -}}
{{- end -}}

{{- define "turing.labels" -}}
app: {{ include "turing.fullname" . }}
chart: {{ include "turing.chart" . }}
release: {{ .Release.Name }}
heritage: {{ .Release.Service }}
{{- if .Values.turing.labels }}
{{ toYaml .Values.turing.labels -}}
{{- end }}
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
{{ if .Values.tags.db }}
    {{- printf "%s.%s.svc.cluster.local" (include "turing.postgresql.name" .) .Release.Namespace -}}
{{- else -}}
    {{- .Values.turing.config.DbConfig.Host -}}
{{- end -}}
{{- end -}}

{{- define "turing.db.port" -}}
{{ if .Values.tags.db }}
    {{ .Values.postgresql.containerPorts.postgresql }}
{{- else -}}
    {{- .Values.turing.config.DbConfig.Port -}}
{{- end -}}
{{- end -}}

{{- define "turing.db.user" -}}
{{ if .Values.tags.db }}
    {{ .Values.postgresql.postgresqlUsername }}
{{- else -}}
    {{- .Values.turing.config.DbConfig.User -}}
{{- end -}}
{{- end -}}

{{- define "turing.db.password" -}}
{{ if .Values.tags.db }}
    {{ .Values.postgresql.postgresqlPassword }}
{{- else -}}
    {{- .Values.turing.config.DbConfig.Password -}}
{{- end -}}
{{- end -}}

{{- define "turing.db.database" -}}
{{ if .Values.tags.db }}
    {{ .Values.postgresql.postgresqlDatabase }}
{{- else -}}
    {{- .Values.turing.config.DbConfig.Database -}}
{{- end -}}
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

{{- define "turing.xpManagement.name" -}}
{{- printf "%s-%s" .Release.Name ((.Values.xpManagement).nameOverride | default "xp-management") -}}
{{- end -}}

{{- define "turing.xpTreatment.name" -}}
{{- printf "%s-%s" .Release.Name ((.Values.xpTreatment).nameOverride | default "xp-treatment") -}}
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

{{- define "turing.plugins.directory" -}}
/app/plugins
{{- end -}}

{{- define "turing.plugins.initContainers" -}}
{{ if .Values.turing.experimentEngines }}
initContainers:
{{ range $expEngine := .Values.turing.experimentEngines }}
{{ if eq (toString $expEngine.type) "rpc-plugin" }}
- name: {{ $expEngine.name }}-plugin
  image: {{ $expEngine.rpcPlugin.image }}
  env:
  - name: PLUGIN_NAME
    value: "{{ $expEngine.name }}"
  - name: PLUGINS_DIR
    value: {{ include "turing.plugins.directory" . }}
  volumeMounts:
  - name: plugins-volume
    mountPath: {{ include "turing.plugins.directory" . }}
{{ end }}
{{ end }}
{{ end }}
{{- end -}}

{{- define "turing.initContainers" -}}
initContainers:
{{ with (include "turing.plugins.initContainers" . | fromYaml) -}}
{{ if .initContainers }}
{{- toYaml .initContainers -}}
{{ end }}
{{- end }}
{{ with .Values.turing.extraInitContainers }}
{{- toYaml . -}}
{{- end }}
{{- end -}}

{{- define "turing.defaultConfig" -}}
ClusterConfig:
  InClusterConfig: {{ .Values.turing.clusterConfig.useInClusterConfig }}
DbConfig:
  Host: {{ include "turing.db.host" . | quote }}
  Port: {{ include "turing.db.port" . }}
  Database:  {{ include "turing.db.database" . }}
  User:  {{ include "turing.db.user" . }}
  Password:  {{ include "turing.db.password" . }}
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
{{ if .Values.turing.experimentEngines }}
Experiment:
{{ range $expEngine := .Values.turing.experimentEngines }}
{{ if or (and (eq (toString $expEngine.name) "xp") (.Values).tags.xp ) (ne (toString $expEngine.name) "xp") }}
  {{ $expEngine.name }}:
{{ if $expEngine.options }}
{{ toYaml $expEngine.options | indent 4 }}
{{ end }}
{{ if eq (toString $expEngine.type) "rpc-plugin" }}
    plugin_binary: {{ include "turing.plugins.directory" . }}/{{ $expEngine.name }}
{{ end }}
{{ end }}
{{ end }}
RouterDefaults:
  ExperimentEnginePlugins:
{{ range $expEngine := .Values.turing.experimentEngines }}
{{ if eq (toString $expEngine.type) "rpc-plugin" }}
    {{ $expEngine.name }}:
      Image: {{ $expEngine.rpcPlugin.image }}
{{- end -}}
{{- end -}}
{{ end }}
{{- if .Values.turing.openApiSpecOverrides }}
OpenapiConfig:
  SpecOverrideFile: /etc/openapi/override.yaml
{{- end -}}
{{- end -}}

{{- define "turing.mergedConfig" -}}
{{- $defaultConfig := include "turing.defaultConfig" . | fromYaml -}}
{{ $defaultConfig | merge .Values.turing.config | toYaml }}
{{- end -}}

{{- define "turing.experimentEngines" -}}
{{- $mergedConfig := include "turing.mergedConfig" . | fromYaml -}}
{{ $c := "false" }}
Experiment:
# Add all configured Experiment Engines
{{ range $expEngine := $mergedConfig.Experiment }}
{{ if eq (toString ($expEngine).name) "xp" }}
{{ $c := "true" }}
{{ end }}
  {{ $expEngine }}
{{ end }}
# Add XP Experiment Engine if it is enabled but overwritten by .Values.turing.config previously
{{ if and (eq $c "false") (.Values).tags.xp }}
{{ range $expEngine := .Values.turing.experimentEngines }}
{{ if eq (toString ($expEngine).name) "xp" }}
  {{ ($expEngine).name }}:
{{ if ($expEngine).options }}
{{ toYaml ($expEngine).options | indent 4 }}
{{ end }}
{{ if eq (toString ($expEngine).type) "rpc-plugin" }}
    plugin_binary: {{ include "turing.plugins.directory" . }}/{{ $expEngine.name }}
{{ end }}
{{ end }}
{{ end }}
{{ end }}
{{- end -}}

{{- define "turing.config" -}}
{{- $mergedConfig := include "turing.mergedConfig" . | fromYaml -}}
{{- $experimentEngines := include "turing.experimentEngines" . | fromYaml -}}
{{ $mergedConfig | merge $experimentEngines | toYaml }}
{{- end -}}

{{- define "turing.defaultOpenApiSpecOverrides" -}}
{{ if (.Values).tags.xp }}
components:
  schemas:
    ExperimentEngineType:
      enum:
      - xp
{{ end }}
{{- end -}}

{{- define "turing.mergedOpenApiSpecOverrides" -}}
{{- $defaultOpenApiSpecOverrides := include "turing.defaultOpenApiSpecOverrides" . | fromYaml -}}
{{ $defaultOpenApiSpecOverrides | merge .Values.turing.openApiSpecOverrides | toYaml }}
{{- end -}}

{{- define "turing.openApiSpecOverridesEnums" -}}
{{- $mergedOpenApiSpecOverrides := include "turing.mergedOpenApiSpecOverrides" . | fromYaml -}}
{{- $c := "false" -}}
components:
  schemas:
    ExperimentEngineType:
      enum:
{{- range $enum := $mergedOpenApiSpecOverrides.components.schemas.ExperimentEngineType.enum }}
{{- if eq $enum "xp" -}}
{{- $c := "true" -}}
{{- end -}}
{{- if ne (toString $enum) "xp" }}
        - {{ $enum }}
{{- end -}}
{{ end }}
{{- if and (eq $c "false") (.Values).tags.xp }}
        - xp
{{ end }}
{{- end -}}

{{- define "turing.finalOpenApiSpecOverrides" -}}
{{- $mergedOpenApiSpecOverrides := include "turing.mergedOpenApiSpecOverrides" . | fromYaml -}}
{{- $openApiSpecOverridesEnums := include "turing.openApiSpecOverridesEnums" . | fromYaml -}}
{{ $mergedOpenApiSpecOverrides | merge $openApiSpecOverridesEnums | toYaml }}
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
