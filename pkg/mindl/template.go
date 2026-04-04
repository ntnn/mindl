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

const asciiUpperCaseOffset = 32

// Title is a simplistic implementation of the deprecated strings.Title.
// It does not handle unicode.
func Title(in string) string {
	split := strings.Split(in, " ")
	for i := range split {
		if split[i][0] >= 'a' || split[i][0] <= 'z' {
			split[i] = string(split[i][0]-asciiUpperCaseOffset) + split[i][1:]
		}
	}
	return strings.Join(split, " ")
}

var FuncMap = map[string]any{
	"title": Title,
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
