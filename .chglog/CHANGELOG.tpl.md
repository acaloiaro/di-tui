{{- if .Versions}}
# Changelog

{{range .Versions}}
## [{{.Tag.Name}}]({{$.Info.RepositoryURL}}/releases/tag/{{.Tag.Name}}) ({{.Tag.Date}})
{{range .CommitGroups}}
### {{.Title}}
{{range .Commits}}
- {{.Header}}
{{end}}
{{end}}
{{end}}
{{end}}
