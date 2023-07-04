{{/*
Expand the name of the chart.
*/}}
{{- define "runtime.name" -}}
    {{- printf "%s-%s" (include "cf-runtime.name" .) "runtime" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "runtime.fullname" -}}
    {{- printf "%s-%s" (include "cf-runtime.fullname" .) "runtime" | trunc 63 | trimSuffix "-" }}
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
  {{- if .Values.global.codefresh.userToken.token }}
valueFrom:
  secretKeyRef:
    name: {{ include "runtime.installation-token-secret-name" . }}
    key: codefresh-api-token
  {{- else if .Values.global.codefresh.userToken.secretKeyRef  }}
valueFrom:
  secretKeyRef:
  {{- .Values.global.codefresh.userToken.secretKeyRef | toYaml | nindent 4 }}
  {{- else }}
    {{- fail "global.codefresh.userToken is mandatory. Set token or secretKeyRef!" }}
  {{- end }}
{{- end }}

{{- define "runtime.installation-token-secret-name" }}
{{- print "codefresh-user-token" }}
{{- end }}