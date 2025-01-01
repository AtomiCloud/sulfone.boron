{{- $secrets := secret "eb0ea61f-2637-482f-ae76-340eabf367a0" "raichu" "/" -}}
{{- range $secrets -}}

{{- if eq .Key "TOFU_BACKEND" }}
tofu_backend = "{{ .Value }}"
{{- end }}

{{- if eq .Key "SELF_CLIENT_ID" }}
infisical_client_id = "{{ .Value }}"
{{- end }}

{{- if eq .Key "SELF_CLIENT_SECRET" }}
infisical_client_secret = "{{ .Value }}"
{{- end }}

{{- end }}
