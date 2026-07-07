{{- define "servasec.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "servasec.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{- define "servasec.labels" -}}
helm.sh/chart: {{ include "servasec.name" . }}-{{ .Chart.Version | replace "+" "_" }}
{{ include "servasec.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{- define "servasec.selectorLabels" -}}
app.kubernetes.io/name: {{ include "servasec.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "servasec.backendName" -}}
{{ include "servasec.fullname" . }}-backend
{{- end }}

{{- define "servasec.frontendName" -}}
{{ include "servasec.fullname" . }}-frontend
{{- end }}

{{- define "servasec.postgresqlName" -}}
{{ include "servasec.fullname" . }}-postgresql
{{- end }}

{{- define "servasec.databaseUrl" -}}
{{- if .Values.postgresql.internal -}}
postgres://{{ .Values.postgresql.user }}:{{ .Values.secrets.postgresPassword | default .Values.postgresql.password | urlquery }}@{{ include "servasec.postgresqlName" . }}:5432/{{ .Values.postgresql.database }}?sslmode=disable
{{- else -}}
{{- .Values.postgresql.externalUrl }}
{{- end -}}
{{- end }}
