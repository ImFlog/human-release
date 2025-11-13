# Release Notes: {{ .FromRef }}...{{ .ToRef }}

*Generated on {{ .Date.Format "Mon, Jan 2, 2006" }}*

{{ range $category, $summaries := .Organized }}
## {{ categoryName $category }}

{{ range $summary := $summaries }}
- **{{ $summary.Title }}** - {{ $summary.Description }}
  - Commit: `{{ truncateHash $summary.CommitHash }}`
{{ end }}

{{ end }}
