{{/*
Expand the name of the chart.
*/}}
{{- define "mlp.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
*/}}
{{- define "mlp.fullname" -}}
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

{{- define "mlp.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version -}}
{{- end -}}

{{- define "mlp.postgresql.name" -}}
{{- printf "%s-%s" .Release.Name (default "postgresql" .Values.postgresql.nameOverride) -}}
{{- end -}}

{{- define "mlp.db.host" -}}
{{- printf "%s.%s.svc.cluster.local" (include "mlp.postgresql.name" .) .Release.Namespace -}}
{{- end -}}

{{- define "mlp.encryption.key" -}}
{{- $encryptionKey := default (randAlpha 12) .Values.encryption.key -}}
{{- if and .Values.global.mlp .Values.global.mlp.encryption -}}
    {{- .Values.global.mlp.encryption.key | default $encryptionKey -}}
{{- else -}}
    {{ $encryptionKey }}
{{- end -}}
{{- end -}}

{{- define "mlp.oauthClientId" -}}
{{- .Values.global.oauthClientId | default .Values.oauthClientId -}}
{{- end -}}

{{- define "mlp.gitlab.envVars" -}}
{{- $gitlabEnabled := eq (.Values.gitlab.enabled | toString) "true" -}}
- name: GITLAB_ENABLED
  value: {{ $gitlabEnabled | quote }}
{{- if $gitlabEnabled }}
- name: GITLAB_HOST
  value: "{{ .Values.gitlab.host }}"
- name: GITLAB_REDIRECT_URL
  value: "{{ .Values.gitlab.redirectUrl }}"
- name: GITLAB_OAUTH_SCOPES
  value: "{{ .Values.gitlab.oauthScopes }}"
- name: GITLAB_CLIENT_ID
  valueFrom:
    secretKeyRef:
      name: {{ include "mlp.fullname" . }}-secrets
      key: gitlab-client-id
- name: GITLAB_CLIENT_SECRET
  valueFrom:
    secretKeyRef:
      name: {{ include "mlp.fullname" . }}-secrets
      key: gitlab-client-secret
{{- end -}}
{{- end -}}