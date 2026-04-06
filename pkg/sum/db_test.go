package sum

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestDBSetGet(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		urlTemplate string
		inArchive   string
		os          string
		arch        string
		sum         string
		comment     string
	}{
		"basic entry": {
			urlTemplate: "https://example.com/{{.Version}}/tool.tar.gz",
			inArchive:   "tool",
			os:          "linux",
			arch:        "amd64",
			sum:         "abc123",
			comment:     "sha512",
		},
		"windows entry": {
			urlTemplate: "https://example.com/{{.Version}}/tool.zip",
			inArchive:   "tool.exe",
			os:          "windows",
			arch:        "amd64",
			sum:         "def456",
			comment:     "sha512",
		},
		"empty comment": {
			urlTemplate: "https://example.com/tool.tar.gz",
			inArchive:   "bin/tool",
			os:          "darwin",
			arch:        "arm64",
			sum:         "ghi789",
			comment:     "",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db := &DB{entries: map[string]Entry{}}
			db.Set(tc.urlTemplate, tc.inArchive, tc.os, tc.arch, tc.sum, tc.comment)

			got, ok := db.Get(tc.urlTemplate, tc.inArchive, tc.os, tc.arch)
			if !ok {
				t.Fatal("expected entry to be found")
			}
			if got.URLTemplate != tc.urlTemplate {
				t.Errorf("URLTemplate = %q, want %q", got.URLTemplate, tc.urlTemplate)
			}
			if got.InArchive != tc.inArchive {
				t.Errorf("InArchive = %q, want %q", got.InArchive, tc.inArchive)
			}
			if got.OS != tc.os {
				t.Errorf("OS = %q, want %q", got.OS, tc.os)
			}
			if got.Arch != tc.arch {
				t.Errorf("Arch = %q, want %q", got.Arch, tc.arch)
			}
			if got.Sum != tc.sum {
				t.Errorf("Sum = %q, want %q", got.Sum, tc.sum)
			}
			if got.Comment != tc.comment {
				t.Errorf("Comment = %q, want %q", got.Comment, tc.comment)
			}
		})
	}
}

func TestDBSetOverwrite(t *testing.T) {
	t.Parallel()

	db := &DB{entries: map[string]Entry{}}
	db.Set("url", "archive", "linux", "amd64", "old-sum", "sha512")
	db.Set("url", "archive", "linux", "amd64", "new-sum", "sha512")

	got, ok := db.Get("url", "archive", "linux", "amd64")
	if !ok {
		t.Fatal("expected entry to be found")
	}
	if got.Sum != "new-sum" {
		t.Errorf("Sum = %q, want %q after overwrite", got.Sum, "new-sum")
	}
}

func TestDBGetNotFound(t *testing.T) {
	t.Parallel()

	db := &DB{entries: map[string]Entry{}}
	_, ok := db.Get("nonexistent", "path", "linux", "amd64")
	if ok {
		t.Error("expected ok to be false for missing entry")
	}
}

func TestDBGetAll(t *testing.T) {
	t.Parallel()

	db := &DB{entries: map[string]Entry{}}
	db.Set("url-a", "archive-a", "linux", "amd64", "sum1", "sha512")
	db.Set("url-a", "archive-a", "linux", "arm64", "sum2", "sha512")
	db.Set("url-a", "archive-a", "darwin", "amd64", "sum3", "sha512")
	db.Set("url-b", "archive-b", "linux", "amd64", "sum4", "sha512")
	db.Set("url-a", "archive-b", "linux", "amd64", "sum5", "sha512")

	tests := map[string]struct {
		urlTemplate string
		inArchive   string
		wantCount   int
	}{
		"three entries for url-a/archive-a": {
			urlTemplate: "url-a",
			inArchive:   "archive-a",
			wantCount:   3,
		},
		"one entry for url-b/archive-b": {
			urlTemplate: "url-b",
			inArchive:   "archive-b",
			wantCount:   1,
		},
		"one entry for url-a/archive-b": {
			urlTemplate: "url-a",
			inArchive:   "archive-b",
			wantCount:   1,
		},
		"no entries for unknown url": {
			urlTemplate: "url-c",
			inArchive:   "archive-a",
			wantCount:   0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := db.GetAll(tc.urlTemplate, tc.inArchive)
			if len(got) != tc.wantCount {
				t.Errorf("GetAll returned %d entries, want %d", len(got), tc.wantCount)
			}
		})
	}
}

func TestRead(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input     string
		wantCount int
	}{
		"empty input": {
			input:     "",
			wantCount: 0,
		},
		"single record": {
			input:     "url,archive,linux,amd64,sum1,sha512\n",
			wantCount: 1,
		},
		"multiple records": {
			input:     "url,archive,linux,amd64,sum1,sha512\nurl,archive,darwin,arm64,sum2,sha512\n",
			wantCount: 2,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			db, err := Read(strings.NewReader(tc.input))
			if err != nil {
				t.Fatalf("Read returned unexpected error: %v", err)
			}
			// Count entries by writing and counting CSV lines
			var buf bytes.Buffer
			if err := db.Write(&buf); err != nil {
				t.Fatalf("Write returned unexpected error: %v", err)
			}
			lines := strings.TrimSpace(buf.String())
			gotCount := 0
			if lines != "" {
				gotCount = strings.Count(lines, "\n") + 1
			}
			if gotCount != tc.wantCount {
				t.Errorf("Read produced %d entries, want %d", gotCount, tc.wantCount)
			}
		})
	}
}

func TestReadError(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input string
	}{
		"wrong column count": {
			input: "url,archive,linux,amd64,sum1\n",
		},
		"too many columns": {
			input: "url,archive,linux,amd64,sum1,sha512,extra\n",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := Read(strings.NewReader(tc.input))
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestWrite(t *testing.T) {
	t.Parallel()

	db := &DB{entries: map[string]Entry{}}
	db.Set("b-url", "archive", "linux", "amd64", "sum-b", "sha512")
	db.Set("a-url", "archive", "linux", "amd64", "sum-a", "sha512")

	var buf bytes.Buffer
	if err := db.Write(&buf); err != nil {
		t.Fatalf("Write returned unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}

	// Entries should be sorted by key; a-url comes before b-url
	if !strings.HasPrefix(lines[0], "a-url,") {
		t.Errorf("first line should start with a-url, got %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "b-url,") {
		t.Errorf("second line should start with b-url, got %q", lines[1])
	}
}

func TestWriteEmpty(t *testing.T) {
	t.Parallel()

	db := &DB{entries: map[string]Entry{}}
	var buf bytes.Buffer
	if err := db.Write(&buf); err != nil {
		t.Fatalf("Write returned unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Errorf("expected empty output for empty DB, got %q", buf.String())
	}
}

func TestReadWriteRoundTrip(t *testing.T) {
	t.Parallel()

	db := &DB{entries: map[string]Entry{}}
	db.Set("url-1", "archive-1", "linux", "amd64", "sum1", "sha512")
	db.Set("url-2", "archive-2", "darwin", "arm64", "sum2", "fnv128a")

	var buf bytes.Buffer
	if err := db.Write(&buf); err != nil {
		t.Fatalf("Write error: %v", err)
	}

	db2, err := Read(&buf)
	if err != nil {
		t.Fatalf("Read error: %v", err)
	}

	for _, entry := range []struct {
		url, archive, os, arch, sum, comment string
	}{
		{"url-1", "archive-1", "linux", "amd64", "sum1", "sha512"},
		{"url-2", "archive-2", "darwin", "arm64", "sum2", "fnv128a"},
	} {
		got, ok := db2.Get(entry.url, entry.archive, entry.os, entry.arch)
		if !ok {
			t.Errorf("entry %s/%s/%s/%s not found after round-trip", entry.url, entry.archive, entry.os, entry.arch)
			continue
		}
		if got.Sum != entry.sum {
			t.Errorf("Sum = %q, want %q", got.Sum, entry.sum)
		}
		if got.Comment != entry.comment {
			t.Errorf("Comment = %q, want %q", got.Comment, entry.comment)
		}
	}
}

func TestOpenSave(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	p := filepath.Join(dir, "test.sum")

	// Open nonexistent path creates empty DB
	db, err := Open(p)
	if err != nil {
		t.Fatalf("Open error: %v", err)
	}

	_, ok := db.Get("any", "thing", "linux", "amd64")
	if ok {
		t.Error("expected empty DB after opening nonexistent path")
	}

	// Set entries and save
	db.Set("url", "archive", "linux", "amd64", "sum1", "sha512")
	if err := db.Save(); err != nil {
		t.Fatalf("Save error: %v", err)
	}

	// Re-open and verify
	db2, err := Open(p)
	if err != nil {
		t.Fatalf("re-Open error: %v", err)
	}

	got, ok := db2.Get("url", "archive", "linux", "amd64")
	if !ok {
		t.Fatal("expected entry to be found after Save/Open")
	}
	if got.Sum != "sum1" {
		t.Errorf("Sum = %q, want %q", got.Sum, "sum1")
	}
}
