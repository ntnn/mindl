package mindl

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createZipArchive creates a zip file containing the given files.
func createZipArchive(t *testing.T, archivePath string, files map[string]string) {
	t.Helper()
	f, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("failed to create zip file: %v", err)
	}

	w := zip.NewWriter(f)

	for name, content := range files {
		fw, err := w.Create(name)
		if err != nil {
			t.Fatalf("failed to create zip entry %q: %v", name, err)
		}
		if _, err := fw.Write([]byte(content)); err != nil {
			t.Fatalf("failed to write zip entry %q: %v", name, err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close zip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close zip file: %v", err)
	}
}

// createTarArchive creates a tar file containing the given files.
func createTarArchive(t *testing.T, files map[string]string) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	w := tar.NewWriter(&buf)

	for name, content := range files {
		hdr := &tar.Header{
			Name: name,
			Mode: 0o644,
			Size: int64(len(content)),
		}
		if err := w.WriteHeader(hdr); err != nil {
			t.Fatalf("failed to write tar header for %q: %v", name, err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatalf("failed to write tar content for %q: %v", name, err)
		}
	}

	if err := w.Close(); err != nil {
		t.Fatalf("failed to close tar writer: %v", err)
	}

	return &buf
}

// writeTarFile writes raw tar bytes to a file.
func writeTarFile(t *testing.T, path string, tarBuf *bytes.Buffer) {
	t.Helper()
	if err := os.WriteFile(path, tarBuf.Bytes(), 0o600); err != nil {
		t.Fatalf("failed to write tar file: %v", err)
	}
}

// writeTarGzFile writes gzip-compressed tar bytes to a file.
func writeTarGzFile(t *testing.T, path string, tarBuf *bytes.Buffer) {
	t.Helper()
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create tar.gz file: %v", err)
	}

	gw := gzip.NewWriter(f)

	if _, err := gw.Write(tarBuf.Bytes()); err != nil {
		t.Fatalf("failed to write gzip data: %v", err)
	}

	if err := gw.Close(); err != nil {
		t.Fatalf("failed to close gzip writer: %v", err)
	}
	if err := f.Close(); err != nil {
		t.Fatalf("failed to close tar.gz file: %v", err)
	}
}

func TestUnzip(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		files       map[string]string
		target      string
		wantContent string
	}{
		"single file": {
			files:       map[string]string{"tool": "binary-content"},
			target:      "tool",
			wantContent: "binary-content",
		},
		"nested path": {
			files:       map[string]string{"dir/tool": "nested-content"},
			target:      "dir/tool",
			wantContent: "nested-content",
		},
		"multiple files extract one": {
			files: map[string]string{
				"file-a": "content-a",
				"file-b": "content-b",
			},
			target:      "file-b",
			wantContent: "content-b",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			archivePath := filepath.Join(dir, "test.zip")
			outputPath := filepath.Join(dir, "output")

			createZipArchive(t, archivePath, tc.files)

			if err := Unzip(archivePath, tc.target, outputPath); err != nil {
				t.Fatalf("Unzip returned unexpected error: %v", err)
			}

			got, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("failed to read output: %v", err)
			}
			if string(got) != tc.wantContent {
				t.Errorf("extracted content = %q, want %q", got, tc.wantContent)
			}
		})
	}
}

func TestUnzipError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	archivePath := filepath.Join(dir, "test.zip")
	createZipArchive(t, archivePath, map[string]string{"existing": "data"})

	err := Unzip(archivePath, "nonexistent", filepath.Join(dir, "output"))
	if err == nil {
		t.Fatal("expected error for missing target in zip, got nil")
	}
}

func TestUntar(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		files       map[string]string
		target      string
		suffix      string // archive file extension
		compress    func(t *testing.T, path string, tarBuf *bytes.Buffer)
		wantContent string
	}{
		"plain tar": {
			files:       map[string]string{"tool": "tar-content"},
			target:      "tool",
			suffix:      ".tar",
			compress:    writeTarFile,
			wantContent: "tar-content",
		},
		"tar.gz": {
			files:       map[string]string{"tool": "gzip-content"},
			target:      "tool",
			suffix:      ".tar.gz",
			compress:    writeTarGzFile,
			wantContent: "gzip-content",
		},
		"nested path in tar.gz": {
			files:       map[string]string{"dir/tool": "nested-gz"},
			target:      "dir/tool",
			suffix:      ".tar.gz",
			compress:    writeTarGzFile,
			wantContent: "nested-gz",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			archivePath := filepath.Join(dir, "archive"+tc.suffix)
			outputPath := filepath.Join(dir, "output")

			tarBuf := createTarArchive(t, tc.files)
			tc.compress(t, archivePath, tarBuf)

			if err := Untar(archivePath, tc.target, outputPath); err != nil {
				t.Fatalf("Untar returned unexpected error: %v", err)
			}

			got, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("failed to read output: %v", err)
			}
			if string(got) != tc.wantContent {
				t.Errorf("extracted content = %q, want %q", got, tc.wantContent)
			}
		})
	}
}

func TestUntarError(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	archivePath := filepath.Join(dir, "archive.tar")
	tarBuf := createTarArchive(t, map[string]string{"existing": "data"})
	writeTarFile(t, archivePath, tarBuf)

	err := Untar(archivePath, "nonexistent", filepath.Join(dir, "output"))
	if err == nil {
		t.Fatal("expected error for missing target in tar, got nil")
	}
	if !strings.Contains(err.Error(), "hit end of archive") {
		t.Errorf("error = %q, want it to contain %q", err, "hit end of archive")
	}
}

func TestUnarchive(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		suffix        string
		createArchive func(t *testing.T, archivePath string)
		target        string
		wantContent   string
	}{
		"zip dispatches to Unzip": {
			suffix: ".zip",
			createArchive: func(t *testing.T, archivePath string) {
				t.Helper()
				createZipArchive(t, archivePath, map[string]string{"tool": "zip-data"})
			},
			target:      "tool",
			wantContent: "zip-data",
		},
		"tar.gz dispatches to Untar": {
			suffix: ".tar.gz",
			createArchive: func(t *testing.T, archivePath string) {
				t.Helper()
				tarBuf := createTarArchive(t, map[string]string{"tool": "tar-data"})
				writeTarGzFile(t, archivePath, tarBuf)
			},
			target:      "tool",
			wantContent: "tar-data",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			archivePath := filepath.Join(dir, "archive"+tc.suffix)
			outputPath := filepath.Join(dir, "output")

			tc.createArchive(t, archivePath)

			if err := Unarchive(archivePath, tc.target, outputPath); err != nil {
				t.Fatalf("Unarchive returned unexpected error: %v", err)
			}

			got, err := os.ReadFile(outputPath)
			if err != nil {
				t.Fatalf("failed to read output: %v", err)
			}
			if string(got) != tc.wantContent {
				t.Errorf("extracted content = %q, want %q", got, tc.wantContent)
			}
		})
	}
}
