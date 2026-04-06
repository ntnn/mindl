package sum

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathMatchesHash(t *testing.T) {
	t.Parallel()

	// Prepare a temp file with known content and its SHA-512 hash
	dir := t.TempDir()
	p := filepath.Join(dir, "testfile")
	if err := os.WriteFile(p, []byte("hello world"), 0o600); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	correctHash := "MJ7MSJwS1utMxA9QyQLytNDtd+5RGnx6m808qG1M2G+YndNbxf9JlnDaNCVbRbDP2DDoH2Bdz33FVC6TrpzXbw=="

	tests := map[string]struct {
		path     string
		expected string
		want     bool
	}{
		"matching hash": {
			path:     p,
			expected: correctHash,
			want:     true,
		},
		"non-matching hash": {
			path:     p,
			expected: "wronghash",
			want:     false,
		},
		"nonexistent file": {
			path:     filepath.Join(dir, "missing"),
			expected: correctHash,
			want:     false,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := PathMatchesHash(tc.path, Sha512Path, tc.expected)
			if err != nil {
				t.Fatalf("PathMatchesHash returned unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("PathMatchesHash = %v, want %v", got, tc.want)
			}
		})
	}
}
