package mindl

import (
	"maps"
	"slices"
)

// Target is an OS/Arch combination.
type Target struct {
	OS   string
	Arch string
}

// CurrentTarget returns the [Target] for the current OS and architecture.
func CurrentTarget() Target {
	return Target{OS: OS(), Arch: Arch()}
}

// CommonTargets contains common OS/Arch combinations.
var CommonTargets = []Target{
	{OS: "linux", Arch: "amd64"},
	{OS: "linux", Arch: "arm64"},
	{OS: "darwin", Arch: "amd64"},
	{OS: "darwin", Arch: "arm64"},
	{OS: "windows", Arch: "amd64"},
}

// DeduplicateTargets returns a deduplicated copy of targets.
func DeduplicateTargets(targets []Target) []Target {
	m := map[string]Target{}
	for _, t := range targets {
		m[t.OS+t.Arch] = t
	}
	return slices.Collect(maps.Values(m))
}
