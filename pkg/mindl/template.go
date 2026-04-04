package mindl

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"
	"text/template"
)

// TemplateData holds template rendering context.
type TemplateData struct {
	Version string `json:"version"`
	OS      string `json:"os"`
	Arch    string `json:"arch"`

	// OSArchive contains an os-dependent archive extension.
	// Default is `tar.gz`.
	// On windows it is `zip`.
	OSArchive string `json:"osarchive"`

	// Exe contains `.exe` on windows.
	Exe string `json:"exe"`
}

// NewTemplateData returns a prepared TemplateData.
// os defaults to MINDL_OS and then runtime.GOOS.
// arch defaults to MINDL_ARCH and then runtime.GOARCH.
func NewTemplateData(os, arch string) *TemplateData {
	td := new(TemplateData)
	td.OS = os
	if td.OS == "" {
		td.OS = getenv("MINDL_OD", runtime.GOOS)
	}
	td.Arch = arch
	if td.Arch == "" {
		td.Arch = getenv("MINDL_ARCH", runtime.GOARCH)
	}

	td.OSArchive = "tar.gz"
	if td.OS == "windows" {
		td.OSArchive = "zip"
		td.Exe = ".exe"
	}

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

// FuncMap contains template functions used by [Template].
var FuncMap = map[string]any{
	"title": Title,
}

// Template parses and executes the given text as a template with the given data.
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
