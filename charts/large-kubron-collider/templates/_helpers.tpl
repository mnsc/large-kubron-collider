{{- define "lkc.magnet.fullname" -}}
{{- printf "%s-%s" .Release.Name "magnet" | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{- define "lkc.magnets.serviceName" -}}
{{- printf "%s-%s" .Release.Name "magnets" | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{- define "lkc.experiment.fullname" -}}
{{- printf "%s-%s" .Release.Name .Values.experiment.name | trunc 63 | trimSuffix "-" -}}
{{- end }}