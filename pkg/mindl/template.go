package mindl

import "runtime"

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
