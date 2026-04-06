package mindl

import (
	"sort"
	"testing"
)

func TestDeduplicateTargets(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		input     []Target
		wantCount int
		// wantTargets lists every unique target that must be present
		wantTargets []Target
	}{
		"no duplicates": {
			input: []Target{
				{OS: "linux", Arch: "amd64"},
				{OS: "darwin", Arch: "arm64"},
			},
			wantCount: 2,
			wantTargets: []Target{
				{OS: "linux", Arch: "amd64"},
				{OS: "darwin", Arch: "arm64"},
			},
		},
		"with duplicates": {
			input: []Target{
				{OS: "linux", Arch: "amd64"},
				{OS: "linux", Arch: "amd64"},
				{OS: "darwin", Arch: "arm64"},
			},
			wantCount: 2,
			wantTargets: []Target{
				{OS: "linux", Arch: "amd64"},
				{OS: "darwin", Arch: "arm64"},
			},
		},
		"all identical": {
			input: []Target{
				{OS: "linux", Arch: "amd64"},
				{OS: "linux", Arch: "amd64"},
				{OS: "linux", Arch: "amd64"},
			},
			wantCount: 1,
			wantTargets: []Target{
				{OS: "linux", Arch: "amd64"},
			},
		},
		"empty input": {
			input:       []Target{},
			wantCount:   0,
			wantTargets: []Target{},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := DeduplicateTargets(tc.input)
			if len(got) != tc.wantCount {
				t.Fatalf("DeduplicateTargets returned %d targets, want %d", len(got), tc.wantCount)
			}

			// Sort both for comparison
			sortTargets := func(ts []Target) {
				sort.Slice(ts, func(i, j int) bool {
					if ts[i].OS != ts[j].OS {
						return ts[i].OS < ts[j].OS
					}
					return ts[i].Arch < ts[j].Arch
				})
			}
			sortTargets(got)
			sortTargets(tc.wantTargets)

			for i, want := range tc.wantTargets {
				if got[i] != want {
					t.Errorf("target[%d] = %v, want %v", i, got[i], want)
				}
			}
		})
	}
}
