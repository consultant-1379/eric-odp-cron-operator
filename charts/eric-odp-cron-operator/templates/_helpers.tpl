{{/*
Expand the name of the chart.
*/}}
{{- define "eric-odp-cron-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "eric-odp-cron-operator.fullname" -}}
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

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "eric-odp-cron-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" -}}
{{- end -}}

{{/*
Kubernetes labels
*/}}
{{- define "eric-odp-cron-operator.kubernetes-labels" -}}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
helm.sh/chart: {{ include "eric-odp-cron-operator.chart" . }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "eric-odp-cron-operator.labels" -}}
  {{- $kubernetesLabels := include "eric-odp-cron-operator.kubernetes-labels" . | fromYaml -}}
  {{- $selectorLabels := include "eric-odp-cron-operator.selectorLabels" . | fromYaml -}}
  {{- $globalLabels := (.Values.global).labels -}}
  {{- $serviceLabels := .Values.labels -}}
{{- include "eric-odp-cron-operator.mergeLabels" (dict "location" .Template.Name "sources" (list $kubernetesLabels $selectorLabels $globalLabels $serviceLabels)) }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "eric-odp-cron-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "eric-odp-cron-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create a map from ".Values.global" with defaults if missing in values file.
This hides defaults from values file.
*/}}
{{ define "eric-odp-cron-operator.global" }}
  {{- $globalDefaults := dict "security" (dict "tls" (dict "enabled" true)) -}}
  {{- $globalDefaults := merge $globalDefaults (dict "nodeSelector" (dict)) -}}
  {{- $globalDefaults := merge $globalDefaults (dict "registry" (dict "url" "armdocker.rnd.ericsson.se")) -}}
  {{- $globalDefaults := merge $globalDefaults (dict "pullSecret" "") -}}
  {{- $globalDefaults := merge $globalDefaults (dict "timezone" "UTC") -}}
  {{- $globalDefaults := merge $globalDefaults (dict "externalIPv4" (dict "enabled")) -}}
  {{- $globalDefaults := merge $globalDefaults (dict "externalIPv6" (dict "enabled")) -}}
  {{- $globalDefaults := merge $globalDefaults (dict "log" (dict "outputs" (list  "k8sLevel"))) -}}
  {{ if .Values.global }}
    {{- mergeOverwrite $globalDefaults .Values.global | toJson -}}
  {{ else }}
    {{- $globalDefaults | toJson -}}
  {{ end }}
{{ end }}

{{/*
Create a user defined annotation (DR-D1121-065, DR-D1121-060)
*/}}
{{ define "eric-odp-cron-operator.config-annotations" }}
  {{- $global := (.Values.global).annotations -}}
  {{- $service := .Values.annotations -}}
  {{- include "eric-odp-cron-operator.mergeAnnotations" (dict "location" .Template.Name "sources" (list $global $service)) -}}
{{- end }}

{{/*
Define the common annotations
*/}}
{{- define "eric-odp-cron-operator.annotations" -}}
{{- $productInfo := include "eric-odp-cron-operator.product-info" . | fromYaml -}}
{{- $config := include "eric-odp-cron-operator.config-annotations" . | fromYaml -}}
{{- include "eric-odp-cron-operator.mergeAnnotations" (dict "location" .Template.Name "sources" (list $productInfo $config)) }}
{{- end -}}

{{/*
Annotations with Prometheus
*/}}
{{- define "eric-odp-cron-operator.annotations-with-prometheus" -}}
  {{- $annotations := include "eric-odp-cron-operator.annotations" . | fromYaml -}}
  {{- $prometheus := include "eric-odp-cron-operator.prometheus" . | fromYaml -}}
  {{- include "eric-odp-cron-operator.mergeAnnotations" (dict "location" .Template.Name "sources" (list $annotations $prometheus)) }}
{{- end -}}


{{/*
Create prometheus info
*/}}
{{- define "eric-odp-cron-operator.prometheus" }}
prometheus.io/scrape-role: {{ .Values.prometheus.role | quote }}
prometheus.io/scrape-interval: {{ .Values.prometheus.interval | quote }}
prometheus.io/path: {{ .Values.prometheus.path | quote }}
prometheus.io/port: {{ .Values.service.metrics.port | quote }}
{{- end }}

{{/*
Create Ericsson product specific annotations
*/}}
{{- define "eric-odp-cron-operator.helm-annotations_product_name" -}}
{{- $productname := (fromYaml (.Files.Get "eric-product-info.yaml")).productName -}}
{{- print $productname | quote }}
{{- end -}}
{{- define "eric-odp-cron-operator.helm-annotations_product_number" -}}
{{- $productNumber := (fromYaml (.Files.Get "eric-product-info.yaml")).productNumber -}}
{{- print $productNumber | quote }}
{{- end -}}
{{- define "eric-odp-cron-operator.helm-annotations_product_revision" -}}
{{- $ddbMajorVersion := mustRegexFind "^([0-9]+)\\.([0-9]+)\\.([0-9]+)((-|\\+)EP[0-9]+)*((-|\\+)[0-9]+)*" .Chart.Version -}}
{{- print $ddbMajorVersion | quote }}
{{- end -}}

{{/*
Create a dict of annotations for the product information (DR-D1121-064, DR-D1121-067).
*/}}
{{- define "eric-odp-cron-operator.product-info" }}
ericsson.com/product-name: {{ template "eric-odp-cron-operator.helm-annotations_product_name" . }}
ericsson.com/product-number: {{ template "eric-odp-cron-operator.helm-annotations_product_number" . }}
ericsson.com/product-revision: {{ template "eric-odp-cron-operator.helm-annotations_product_revision" . }}
{{- end }}

{{/*
Create a merged set of nodeSelectors from global and service level.
*/}}
{{ define "eric-odp-cron-operator.nodeSelector" }}
  {{- $g := fromJson (include "eric-odp-cron-operator.global" .) -}}
  {{- if .Values.nodeSelector -}}
    {{- range $key, $localValue := .Values.nodeSelector -}}
      {{- if hasKey $g.nodeSelector $key -}}
          {{- $globalValue := index $g.nodeSelector $key -}}
          {{- if ne $globalValue $localValue -}}
            {{- printf "nodeSelector \"%s\" is specified in both global (%s: %s) and service level (%s: %s) with differing values which is not allowed." $key $key $globalValue $key $localValue | fail -}}
          {{- end -}}
      {{- end -}}
    {{- end -}}
    {{- toYaml (merge $g.nodeSelector .Values.nodeSelector) | trim -}}
  {{- else -}}
    {{- toYaml $g.nodeSelector | trim -}}
  {{- end -}}
{{ end }}

{{/*
Create image pull secrets
*/}}
{{- define "eric-odp-cron-operator.pullSecrets" -}}
{{- $g := fromJson (include "eric-odp-cron-operator.global" .) -}}
{{- $pullSecret := $g.pullSecret -}}
{{- if .Values.imageCredentials -}}
    {{- if .Values.imageCredentials.pullSecret -}}
        {{- $pullSecret = .Values.imageCredentials.pullSecret -}}
    {{- end -}}
{{- end -}}
{{- print $pullSecret -}}
{{- end -}}

{{/*
The image path (DR-D1121-067, DR-D1121-104, DR-D1121-105, DR-D1121-106)
*/}}
{{- define "eric-odp-cron-operator.mainImagePath" }}
    {{- $productInfo := fromYaml (.Files.Get "eric-product-info.yaml") -}}
    {{- $registryUrl := $productInfo.images.ericodpcronoperator.registry -}}
    {{- $repoPath := $productInfo.images.ericodpcronoperator.repoPath -}}
    {{- $name := $productInfo.images.ericodpcronoperator.name -}}
    {{- $tag := $productInfo.images.ericodpcronoperator.tag -}}
    {{- if ((.Values).global).registry -}}
        {{- if .Values.global.registry.url -}}
            {{- $registryUrl = .Values.global.registry.url -}}
        {{- end -}}
        {{- if not (kindIs "invalid" .Values.global.registry.repoPath) -}}
            {{- $repoPath = .Values.global.registry.repoPath -}}
        {{- end -}}
    {{- end -}}
    {{- if .Values.imageCredentials -}}
        {{- if (((.Values).imageCredentials).registry).url -}}
            {{- $registryUrl = .Values.imageCredentials.registry.url -}}
        {{- end -}}
        {{- if not (kindIs "invalid" .Values.imageCredentials.repoPath) -}}
            {{- $repoPath = .Values.imageCredentials.repoPath -}}
        {{- end -}}
        {{- if .Values.imageCredentials.ericodpcronoperator -}}
            {{- if ((((.Values).imageCredentials).ericodpcronoperator).registry).url -}}
                {{- $registryUrl = .Values.imageCredentials.ericodpcronoperator.registry.url -}}
            {{- end -}}
            {{- if not (kindIs "invalid" .Values.imageCredentials.ericodpcronoperator.repoPath) -}}
                {{- $repoPath = .Values.imageCredentials.ericodpcronoperator.repoPath -}}
            {{- end -}}
        {{- end -}}
    {{- end -}}
    {{- if $repoPath -}}
        {{- $repoPath = printf "%s/" $repoPath -}}
    {{- end -}}
    {{- printf "%s/%s%s:%s" $registryUrl $repoPath $name $tag -}}
{{- end -}}

{{/*
Create pull policy
*/}}
{{- define "eric-odp-cron-operator.imagePullPolicy" -}}
{{- $globalRegistryPullPolicy := "IfNotPresent" -}}
    {{- if .Values.imageCredentials.ericodpcronoperator -}}
        {{- if .Values.imageCredentials.ericodpcronoperator.registry -}}
            {{- if .Values.imageCredentials.ericodpcronoperator.registry.imagePullPolicy -}}
                 {{- $globalRegistryPullPolicy = .Values.imageCredentials.ericodpcronoperator.registry.imagePullPolicy -}}
            {{- end -}}
        {{- else if .Values.imageCredentials.pullPolicy -}}
            {{- $globalRegistryPullPolicy = .Values.imageCredentials.pullPolicy -}}
        {{- else if .Values.global -}}
            {{- if .Values.global.registry -}}
                {{- if .Values.global.registry.imagePullPolicy -}}
                    {{- $globalRegistryPullPolicy = .Values.global.registry.imagePullPolicy -}}
                {{- end -}}
            {{- end -}}
        {{- end -}}
    {{- else if .Values.global -}}
        {{- if .Values.global.registry -}}
            {{- if .Values.global.registry.imagePullPolicy -}}
                {{- $globalRegistryPullPolicy = .Values.global.registry.imagePullPolicy -}}
            {{- end -}}
        {{- end -}}
    {{- end -}}
    {{- print $globalRegistryPullPolicy -}}
{{- end -}}

{{/*
Define RoleBinding value
*/}}
{{- define "eric-odp-cron-operator.roleBinding" -}}
{{- $rolebinding := false -}}
{{- if .Values.global -}}
    {{- if .Values.global.security -}}
        {{- if .Values.global.security.policyBinding -}}
            {{- if hasKey .Values.global.security.policyBinding "create" -}}
                {{- $rolebinding = .Values.global.security.policyBinding.create -}}
            {{- end -}}
        {{- end -}}
    {{- end -}}
{{- end -}}
{{- $rolebinding -}}
{{- end -}}

{{/*
Define reference to SecurityPolicy
*/}}
{{- define "eric-odp-cron-operator.securityPolicyReference" -}}
{{- $policyreference := "default-restricted-security-policy" -}}
{{- if .Values.global -}}
    {{- if .Values.global.security -}}
        {{- if .Values.global.security.policyReferenceMap -}}
            {{- if hasKey .Values.global.security.policyReferenceMap "default-restricted-security-policy" -}}
                {{- $policyreference = index .Values "global" "security" "policyReferenceMap" "default-restricted-security-policy" -}}
            {{- end -}}
        {{- end -}}
    {{- end -}}
{{- end -}}
{{- $policyreference -}}
{{- end -}}


{{/*
To support Dual stack.
*/}}
{{- define "eric-odp-cron-operator.internalIPFamily" -}}
{{- if  .Values.global -}}
  {{- if  .Values.global.internalIPFamily -}}
    {{- .Values.global.internalIPFamily | toString -}}
  {{- else -}}
    {{- "none" -}}
  {{- end -}}
{{- else -}}
{{- "none" -}}
{{- end -}}
{{- end -}}


{{/*
DR-D470222-010
Configuration of Log Collection Streaming Method
*/}}
{{- define "eric-odp-cron-operator.log.streamingMethod" -}}
{{- $defaultMethod := "indirect" }}
{{- if .Values.global -}}
    {{- if .Values.global.log -}}
        {{- if .Values.global.log.streamingMethod -}}
            {{- $streamingMethod := .Values.global.log.streamingMethod }}
                {{- if not $streamingMethod }}
                    {{- if (.Values.global.log).streamingMethod -}}
                        {{- $streamingMethod = (.Values.global.log).streamingMethod }}
                    {{- else -}}
                        {{- $streamingMethod = $defaultMethod -}}
                    {{- end }}
                {{- end }}

                {{- if or (eq $streamingMethod "direct") (eq $streamingMethod "indirect") }}
                    {{- $streamingMethod -}}
                {{- else }}
                    {{- $defaultMethod -}}
                {{- end }}
            {{- end }}
        {{- end }}
    {{- end }}
{{- end }}

{{/*
DR-D1126-005 Set resources for the container
*/}}
{{- define "eric-odp-cron-operator.containerResources" -}}
{{- $container := index . -}}
requests:
{{- if $container.requests.cpu }}
  cpu: {{ $container.requests.cpu | quote }}
{{- end }}
{{- if $container.requests.memory }}
  memory: {{ $container.requests.memory | quote }}
{{- end }}
{{- if (index $container "requests" "ephemeral-storage") }}
  ephemeral-storage: {{ (index $container "requests" "ephemeral-storage" | quote) }}
{{- end }}
limits:
{{- if $container.limits.cpu }}
  cpu: {{ $container.limits.cpu | quote }}
{{- end }}
{{- if $container.limits.memory }}
  memory: {{ $container.limits.memory | quote }}
{{- end }}
{{- if (index $container "limits" "ephemeral-storage") }}
  ephemeral-storage: {{ (index $container "limits" "ephemeral-storage" | quote) }}
{{- end }}
{{- end -}}

{{/*
The image path (DR-D1121-067, DR-D1121-104, DR-D1121-105, DR-D1121-106)
*/}}
{{- define "eric-odp-cron-operator.cronWrapperImagePath" }}
    {{- $productInfo := fromYaml (.Files.Get "eric-product-info.yaml") -}}
    {{- $registryUrl := $productInfo.images.ericodpcronwrapper.registry -}}
    {{- $repoPath := $productInfo.images.ericodpcronwrapper.repoPath -}}
    {{- $name := $productInfo.images.ericodpcronwrapper.name -}}
    {{- $tag := $productInfo.images.ericodpcronwrapper.tag -}}
    {{- if ((.Values).global).registry -}}
        {{- if .Values.global.registry.url -}}
            {{- $registryUrl = .Values.global.registry.url -}}
        {{- end -}}
        {{- if not (kindIs "invalid" .Values.global.registry.repoPath) -}}
            {{- $repoPath = .Values.global.registry.repoPath -}}
        {{- end -}}
    {{- end -}}
    {{- if .Values.imageCredentials -}}
        {{- if (((.Values).imageCredentials).registry).url -}}
            {{- $registryUrl = .Values.imageCredentials.registry.url -}}
        {{- end -}}
        {{- if not (kindIs "invalid" .Values.imageCredentials.repoPath) -}}
            {{- $repoPath = .Values.imageCredentials.repoPath -}}
        {{- end -}}
        {{- if .Values.imageCredentials.ericodpcronwrapper -}}
            {{- if ((((.Values).imageCredentials).ericodpcronwrapper).registry).url -}}
                {{- $registryUrl = .Values.imageCredentials.ericodpcronwrapper.registry.url -}}
            {{- end -}}
            {{- if not (kindIs "invalid" .Values.imageCredentials.ericodpcronwrapper.repoPath) -}}
                {{- $repoPath = .Values.imageCredentials.ericodpcronwrapper.repoPath -}}
            {{- end -}}
        {{- end -}}
    {{- end -}}
    {{- if $repoPath -}}
        {{- $repoPath = printf "%s/" $repoPath -}}
    {{- end -}}
    {{- printf "%s/%s%s:%s" $registryUrl $repoPath $name $tag -}}
{{- end -}}
