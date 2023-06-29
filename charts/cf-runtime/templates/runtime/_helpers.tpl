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
