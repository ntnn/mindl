package mindl

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"text/template"
)

type TemplateData struct {
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`
}

func NewTemplateData() *TemplateData {
	td := new(TemplateData)
	td.OS = runtime.GOOS
	td.Arch = runtime.GOARCH
	return td
}

var FuncMap = map[string]any{
	"title": strings.Title,
}

func Template(text string, data *TemplateData) (string, error) {
	t := template.New("").Funcs(FuncMap)

	t, err := t.Parse(text)
	if err != nil {
		return "", fmt.Errorf("error parsing template: %w", err)
	}

	b := &bytes.Buffer{}
	if err := t.Execute(b, data); err != nil {
		return "", fmt.Errorf("error executing template: %w", err)
	}

	return b.String(), nil
}
