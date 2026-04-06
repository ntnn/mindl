package mindl

import (
	"testing"
)

func TestTitle(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input string
		want  string
	}{
		"single lowercase word": {
			input: "hello",
			want:  "Hello",
		},
		"multiple words": {
			input: "hello world",
			want:  "Hello World",
		},
		"already capitalized": {
			input: "Hello",
			want:  "Hello",
		},
		"all caps": {
			input: "HELLO",
			want:  "HELLO",
		},
		"empty string": {
			input: "",
			want:  "",
		},
		"non-alpha first char": {
			input: "123abc",
			want:  "123abc",
		},
		"leading space": {
			input: " hello",
			want:  " Hello",
		},
		"mixed case words": {
			input: "hELLO wORLD",
			want:  "HELLO WORLD",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := Title(tc.input)
			if got != tc.want {
				t.Errorf("Title(%q) = %q, want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestTemplate(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		text string
		data *TemplateData
		want string
	}{
		"version substitution": {
			text: "v{{.Version}}",
			data: &TemplateData{Version: "1.2.3"},
			want: "v1.2.3",
		},
		"all fields": {
			text: "{{.OS}}-{{.Arch}}-{{.Version}}.{{.OSArchive}}{{.Exe}}",
			data: &TemplateData{
				Version:   "2.0",
				OS:        "linux",
				Arch:      "amd64",
				OSArchive: "tar.gz",
				Exe:       "",
			},
			want: "linux-amd64-2.0.tar.gz",
		},
		"title function": {
			text: "{{title .OS}}",
			data: &TemplateData{OS: "linux"},
			want: "Linux",
		},
		"static text only": {
			text: "no-templates-here",
			data: &TemplateData{},
			want: "no-templates-here",
		},
		"windows with exe": {
			text: "tool{{.Exe}}",
			data: &TemplateData{Exe: ".exe"},
			want: "tool.exe",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := Template(tc.text, tc.data)
			if err != nil {
				t.Fatalf("Template returned unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("Template(%q) = %q, want %q", tc.text, got, tc.want)
			}
		})
	}
}

func TestTemplateError(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		text string
	}{
		"unclosed action": {
			text: "{{.Invalid",
		},
		"unknown function": {
			text: "{{unknownfunc .OS}}",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := Template(tc.text, &TemplateData{})
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
