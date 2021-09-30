{{/* vim: set filetype=mustache: */}}
{{/*
Expand the name of the chart.
*/}}
{{- define "brigade-metrics.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "brigade-metrics.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- if contains $name .Release.Name -}}
{{- .Release.Name | trunc 63 | trimSuffix "-" -}}
{{- else -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end -}}
{{- end -}}

{{- define "brigade-metrics.exporter.fullname" -}}
{{ include "brigade-metrics.fullname" . | printf "%s-exporter" }}
{{- end -}}

{{- define "brigade-metrics.prometheus.fullname" -}}
{{ include "brigade-metrics.fullname" . | printf "%s-prometheus" }}
{{- end -}}

{{- define "brigade-metrics.grafana.fullname" -}}
{{ include "brigade-metrics.fullname" . | printf "%s-grafana" }}
{{- end -}}

{{- define "brigade-metrics.nginx.fullname" -}}
{{ include "brigade-metrics.fullname" . | printf "%s-nginx" }}
{{- end -}}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "brigade-metrics.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Common labels
*/}}
{{- define "brigade-metrics.labels" -}}
helm.sh/chart: {{ include "brigade-metrics.chart" . }}
{{ include "brigade-metrics.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end -}}

{{/*
Selector labels
*/}}
{{- define "brigade-metrics.selectorLabels" -}}
app.kubernetes.io/name: {{ include "brigade-metrics.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "brigade-metrics.exporter.labels" -}}
app.kubernetes.io/component: exporter
{{- end -}}

{{- define "brigade-metrics.prometheus.labels" -}}
app.kubernetes.io/component: prometheus
{{- end -}}

{{- define "brigade-metrics.grafana.labels" -}}
app.kubernetes.io/component: grafana
{{- end -}}

{{- define "brigade-metrics.nginx.labels" -}}
app.kubernetes.io/component: nginx
{{- end -}}

{{- define "call-nested" }}
{{- $dot := index . 0 }}
{{- $subchart := index . 1 }}
{{- $template := index . 2 }}
{{- include $template (dict "Chart" (dict "Name" $subchart) "Values" (index $dot.Values $subchart) "Release" $dot.Release "Capabilities" $dot.Capabilities) }}
{{- end }}

{{/*
Return the appropriate apiVersion for a networking object.
*/}}
{{- define "networking.apiVersion" -}}
{{- if semverCompare ">=1.19-0" .Capabilities.KubeVersion.GitVersion -}}
{{- print "networking.k8s.io/v1" -}}
{{- else -}}
{{- print "networking.k8s.io/v1beta1" -}}
{{- end -}}
{{- end -}}

{{- define "networking.apiVersion.isStable" -}}
  {{- eq (include "networking.apiVersion" .) "networking.k8s.io/v1" -}}
{{- end -}}

{{- define "networking.apiVersion.supportIngressClassName" -}}
  {{- semverCompare ">=1.18-0" .Capabilities.KubeVersion.GitVersion -}}
{{- end -}}
