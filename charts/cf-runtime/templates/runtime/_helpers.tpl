{{/*
Expand the name of the chart.
*/}}
{{- define "runtime.name" -}}
    {{- printf "%s" (include "cf-runtime.name" .)  | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "runtime.fullname" -}}
    {{- printf "%s" (include "cf-runtime.fullname" .) | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "runtime.labels" -}}
{{ include "cf-runtime.labels" . }}
codefresh.io/application: runtime
{{- end }}

{{/*
Selector labels
*/}}
{{- define "runtime.selectorLabels" -}}
{{ include "cf-runtime.selectorLabels" . }}
codefresh.io/application: runtime
{{- end }}

{{/*
Return runtime image (classic runtime) with private registry prefix
*/}}
{{- define "runtime.runtimeImageName" -}}
  {{- if .registry -}}
    {{- $imageName :=  (trimPrefix "quay.io/" .imageFullName) -}}
    {{- printf "%s/%s" .registry $imageName -}}
  {{- else -}}
    {{- printf "%s" .imageFullName -}}
  {{- end -}}
{{- end -}}

{{/*
Environment variable value of Codefresh installation token
*/}}
{{- define "runtime.installation-token-env-var-value" -}}
  {{- if .Values.global.codefreshToken }}
valueFrom:
  secretKeyRef:
    name: {{ include "runtime.installation-token-secret-name" . }}
    key: codefresh-api-token
  {{- else if .Values.global.codefreshTokenSecretKeyRef  }}
valueFrom:
  secretKeyRef:
  {{- .Values.global.codefreshTokenSecretKeyRef | toYaml | nindent 4 }}
  {{- end }}
{{- end }}

{{/*
Print Codefresh API token secret name
*/}}
{{- define "runtime.installation-token-secret-name" }}
{{- print "codefresh-user-token" }}
{{- end }}

{{/*
Print runtime-environment name
*/}}
{{- define "runtime.runtime-environment-spec.name" }}
{{- if and (not .Values.runtime.agent) }}
  {{- if not (hasPrefix "system/" .Values.global.runtimeName) }}
    {{- fail "ERROR: .runtime.agent is set to false! .global.runtimeName should start with system/ prefix" }}
  {{- else }}
    {{- printf "%s" (required ".global.runtimeName is required" .Values.global.runtimeName) }}
  {{- end }}
{{- else }}
{{- printf "%s" (required ".global.runtimeName is required" .Values.global.runtimeName) }}
{{- end }}
{{- end }}

{{- define "runtime.runtime-environment-spec.codefresh-host" }}
{{- if and (not .Values.global.codefreshHost) }}
  {{- fail "ERROR: .global.codefreshHost is required" }}
{{- else }}
  {{- printf "%s" (trimSuffix "/" .Values.global.codefreshHost) }}
{{- end }}
{{- end }}