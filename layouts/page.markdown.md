{{- /*
  Markdown version of a page for agents (served as <slug>.md). Replaces
  hextra's template, which outputs raw source with unrendered shortcodes.
*/ -}}
# {{ .Title }}
{{ with .Description }}
> {{ . }}
{{ end }}
Canonical: {{ with .OutputFormats.Get "html" }}{{ .Permalink }}{{ else }}{{ .Permalink }}{{ end }}
{{- $links := slice }}
{{- with site.Home.OutputFormats.Get "llms" }}
  {{- $links = $links | append (printf "[llms.txt](%s)" .Permalink) }}
{{- end }}
{{- with site.Home.OutputFormats.Get "llmsfull" }}
  {{- $links = $links | append (printf "[all languages](%s)" .Permalink) }}
{{- end }}
{{- range partial "samples/languages.html" . }}
  {{- $name := .name }}
  {{- with site.Home.OutputFormats.Get (printf "llms%s" .folder) }}
    {{- $links = $links | append (printf "[%s](%s)" $name .Permalink) }}
  {{- end }}
{{- end }}
{{- with $links }}
Full guide for agents: {{ delimit . " · " }}
{{- end }}

{{ partial "samples/absolute-links.html" (dict "page" . "content" .RenderShortcodes) | strings.TrimSpace }}
