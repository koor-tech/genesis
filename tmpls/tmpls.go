package tmpls

import (
	"embed"
	"html/template"
)

//go:embed */*.gotmpl
var embedTemplates embed.FS

func New() (*template.Template, error) {
	tmpls, err := template.ParseFS(embedTemplates, "*/*.gotmpl")
	if err != nil {
		return nil, err
	}

	return tmpls, nil
}
