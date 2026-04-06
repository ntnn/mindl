package mindl

import (
	"testing"
)

func TestHandle(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		tool          Tool
		data          *TemplateData
		wantURL       string
		wantInArchive string
	}{
		"simple substitution": {
			tool: Tool{
				URLTemplate: "https://example.com/v{{.Version}}/tool-{{.OS}}-{{.Arch}}.{{.OSArchive}}",
				InArchive:   "tool-{{.Version}}-{{.OS}}-{{.Arch}}/tool{{.Exe}}",
			},
			data: &TemplateData{
				Version:   "1.0.0",
				OS:        "linux",
				Arch:      "amd64",
				OSArchive: "tar.gz",
				Exe:       "",
			},
			wantURL:       "https://example.com/v1.0.0/tool-linux-amd64.tar.gz",
			wantInArchive: "tool-1.0.0-linux-amd64/tool",
		},
		"windows with exe": {
			tool: Tool{
				URLTemplate: "https://example.com/{{.Version}}/tool.{{.OSArchive}}",
				InArchive:   "tool{{.Exe}}",
			},
			data: &TemplateData{
				Version:   "2.0",
				OS:        "windows",
				Arch:      "amd64",
				OSArchive: "zip",
				Exe:       ".exe",
			},
			wantURL:       "https://example.com/2.0/tool.zip",
			wantInArchive: "tool.exe",
		},
		"with title function": {
			tool: Tool{
				URLTemplate: "https://example.com/{{title .OS}}/tool.tar.gz",
				InArchive:   "tool",
			},
			data:          &TemplateData{OS: "darwin"},
			wantURL:       "https://example.com/Darwin/tool.tar.gz",
			wantInArchive: "tool",
		},
		"no template actions": {
			tool: Tool{
				URLTemplate: "https://example.com/static.tar.gz",
				InArchive:   "bin/tool",
			},
			data:          &TemplateData{},
			wantURL:       "https://example.com/static.tar.gz",
			wantInArchive: "bin/tool",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			th, err := Handle(tc.tool, tc.data)
			if err != nil {
				t.Fatalf("Handle returned unexpected error: %v", err)
			}
			if th.url != tc.wantURL {
				t.Errorf("url = %q, want %q", th.url, tc.wantURL)
			}
			if th.inArchive != tc.wantInArchive {
				t.Errorf("inArchive = %q, want %q", th.inArchive, tc.wantInArchive)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		tool Tool
	}{
		"invalid URL template": {
			tool: Tool{
				URLTemplate: "{{.Invalid",
				InArchive:   "tool",
			},
		},
		"invalid InArchive template": {
			tool: Tool{
				URLTemplate: "https://example.com/tool.tar.gz",
				InArchive:   "{{.Invalid",
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := Handle(tc.tool, &TemplateData{})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
